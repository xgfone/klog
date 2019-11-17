// Copyright 2019 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package klog

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Field represents a key-value pair.
type Field struct {
	Key   string
	Value interface{}
}

// E is equal to F("err", err).
func E(err error) Field { return Field{Key: "err", Value: err} }

// F returns a new Field.
func F(key string, value interface{}) Field { return Field{Key: key, Value: value} }

// Record represents a log record.
type Record struct {
	Name  string    // The logger name, which may be empty
	Time  time.Time // The start time when to emit the log
	Depth int       // The stack depth of the caller

	Lvl    Level   // The log level
	Msg    string  // The log message
	Fields []Field // The key-value pairs
}

// Encoder is used to encode the log record and to write it into the writer.
type Encoder interface {
	// Writer returns the writer.
	Writer() Writer

	// SetWriter resets the writer.
	SetWriter(Writer)

	// Encode encodes the log record and writes it into the writer.
	Encode(Record)
}

// FixRecordEncoder returns an new Encoder, which will fix the record
// then use the encoder to encode the record.
//
// For example, you can use it to adjust the stack.
func FixRecordEncoder(encoder Encoder, fixRecord func(Record) Record) Encoder {
	return wrappedEncoder{Encoder: encoder, fixRecord: fixRecord}
}

type wrappedEncoder struct {
	Encoder

	level     Level
	fixRecord func(Record) Record
}

func (fre wrappedEncoder) Encode(r Record) {
	if fre.level != nil && r.Lvl.Priority() < fre.level.Priority() {
		return
	}

	r.Depth++
	if fre.fixRecord != nil {
		r = fre.fixRecord(r)
	}
	fre.Encoder.Encode(r)
}

// LevelEncoder returns a new Encoder, which will filter the log record
// when the priority of the log record is lower than that of lvl.
func LevelEncoder(lvl Level, enc Encoder) Encoder {
	return wrappedEncoder{Encoder: enc, level: lvl}
}

// EncoderFunc converts a encode function to Encoder, which will encode
// the record into the builder then write the result into the writer.
func EncoderFunc(w Writer, encode func(*Builder, Record)) Encoder {
	if w == nil {
		panic("EncoderFunc: the writer must not be nil")
	} else if encode == nil {
		panic("EncoderFunc: the encode function must not be nil")
	}
	return &encoderFunc{writer: w, encode: encode}
}

type encoderFunc struct {
	writer Writer
	encode func(*Builder, Record)
}

func (ef *encoderFunc) Writer() Writer     { return ef.writer }
func (ef *encoderFunc) SetWriter(w Writer) { ef.writer = w }
func (ef *encoderFunc) Encode(r Record) {
	if ef.encode == nil {
		return
	}

	r.Depth++
	buf := getBuilder()
	ef.encode(buf, r)
	if bs := buf.Bytes(); len(bs) > 0 {
		ef.writer.Write(r.Lvl, bs)
	}
	putBuilder(buf)
}

//////////////////////////////////////////////////////////////////////////////

// EncoderOption represents the option of the encoder to set the encoder.
type EncoderOption interface{}

type option struct {
	Quote      bool
	StringTime bool
	EncodeTime func(*Builder, time.Time)
}

func getOption(options ...EncoderOption) (o option) {
	for _, opt := range options {
		if f, ok := opt.(func(*option)); ok {
			f(&o)
		}
	}
	return
}

// Quote is used by TextEncoder, which will use a pair of double quotation marks
// to surround the string value if it contains the space.
func Quote() EncoderOption { return func(o *option) { o.Quote = true } }

// EncodeTime sets the encoder to encode the time.
func EncodeTime(enc func(*Builder, time.Time)) EncoderOption {
	return func(o *option) { o.EncodeTime = enc }
}

// StringTime encodes the time as the string with the format time.RFC3339Nano
// instead of the integer second.
func StringTime() EncoderOption {
	return func(o *option) {
		o.StringTime = true
		o.EncodeTime = func(buf *Builder, now time.Time) {
			buf.AppendTime(now, time.RFC3339Nano)
		}
	}
}

// IntegerTime encodes the time as the integer second.
func IntegerTime() EncoderOption {
	return EncodeTime(func(buf *Builder, now time.Time) { buf.AppendInt(now.Unix()) })
}

//////////////////////////////////////////////////////////////////////////////

// NothingEncoder encodes nothing.
func NothingEncoder() Encoder { return &encoderFunc{} }

func appendString(buf *Builder, s string, quote bool) {
	if quote && strings.IndexFunc(s, unicode.IsSpace) > -1 {
		buf.AppendJSONString(s)
	} else {
		buf.AppendString(s)
	}
}

// TextEncoder encodes the key-values log as the text.
//
// Notice: the key name of the level is "lvl", that of the message is "msg",
// and that of the time is "t" if enabling the time encoder. If the logger name
// exists, it will encode it and the key name is "logger".
//
// Notice: it will ignore the empty msg.
func TextEncoder(w Writer, options ...EncoderOption) Encoder {
	opt := getOption(options...)
	return EncoderFunc(w, func(buf *Builder, r Record) {
		r.Depth++

		// Time
		if opt.EncodeTime != nil {
			buf.WriteString("t=")
			opt.EncodeTime(buf, r.Time)
			buf.WriteString(" ")
		}

		// Logger Name
		if r.Name != "" {
			buf.WriteString("logger=")
			appendString(buf, r.Name, opt.Quote)
			buf.WriteString(" ")
		}

		// Level
		buf.WriteString("lvl=")
		buf.WriteString(r.Lvl.String())

		// Fields
		for _, field := range r.Fields {
			buf.WriteString(" ")
			buf.WriteString(field.Key)
			buf.WriteString("=")

			switch v := field.Value.(type) {
			case string:
				appendString(buf, v, opt.Quote)
			case error:
				if v == nil {
					buf.AppendAny(nil)
				} else {
					appendString(buf, v.Error(), opt.Quote)
				}
			case time.Time:
				if opt.EncodeTime != nil {
					opt.EncodeTime(buf, v)
				}
			case fmt.Stringer:
				appendString(buf, v.String(), opt.Quote)
			default:
				if err := buf.AppendAnyFmt(field.Value); err != nil {
					buf.WriteString("<klog.TextEncoder:Error:")
					buf.WriteString(err.Error())
					buf.WriteString(">")
				}
			}
		}

		// Message
		if r.Msg != "" {
			buf.WriteString(" msg=")
			appendString(buf, r.Msg, opt.Quote)
		}

		buf.WriteByte('\n')
	})
}

// JSONEncoder encodes the key-values log as json.
//
// Notice: the key name of the level is "lvl", that of the message is "msg",
// and that of the time is "t" if enabling the time encoder. If the logger name
// exists, it will encode it and the key name is "logger".
//
// Notice: it will ignore the empty msg.
func JSONEncoder(w Writer, options ...EncoderOption) Encoder {
	opt := getOption(options...)
	return EncoderFunc(w, func(buf *Builder, r Record) {
		r.Depth++
		buf.WriteByte('{')

		// Time
		if opt.EncodeTime != nil {
			buf.WriteString(`"t":"`)
			opt.EncodeTime(buf, r.Time)
			buf.WriteString(`",`)
		}

		// Logger Name
		if r.Name != "" {
			buf.WriteString(`"logger":`)
			buf.AppendJSONString(r.Name)
			buf.WriteString(`",`)
		}

		// Level
		buf.WriteString(`"lvl":`)
		buf.AppendJSONString(r.Lvl.String())

		// Fields
		for _, field := range r.Fields {
			// Key
			buf.WriteString(`,`)
			buf.AppendJSONString(field.Key)
			buf.WriteString(`:`)

			if t, ok := field.Value.(time.Time); ok {
				if opt.StringTime {
					buf.WriteString(`"`)
					opt.EncodeTime(buf, t)
					buf.WriteString(`"`)
				} else {
					opt.EncodeTime(buf, t)
				}
				continue
			}

			// Value
			if err := buf.AppendJSON(field.Value); err != nil {
				buf.AppendJSONString(fmt.Sprintf(`<klog.TextEncoder:Error:%s>`, err.Error()))
			}
		}

		// Message
		if r.Msg != "" {
			buf.WriteString(`,"msg":`)
			buf.AppendJSONString(r.Msg)
		}

		// End
		buf.WriteString("}\n")
	})
}

// StdJSONEncoder is equal to JSONEncoder, which uses json.Marshal() to encode
// it, but the performance is a little bad.
//
// Notice: it will ignore the empty msg.
func StdJSONEncoder(w Writer, options ...EncoderOption) Encoder {
	opt := getOption(options...)
	return EncoderFunc(w, func(buf *Builder, r Record) {
		r.Depth++

		maps := make(map[string]interface{}, len(r.Fields)+8)
		maps["lvl"] = r.Lvl.String()

		if opt.EncodeTime != nil {
			opt.EncodeTime(buf, r.Time)
			maps["t"] = buf.String()
			buf.Reset()
		}

		if r.Msg != "" {
			maps["msg"] = r.Msg
		}

		if r.Name != "" {
			maps["logger"] = r.Name
		}

		for _, field := range r.Fields {
			if t, ok := field.Value.(time.Time); ok {
				opt.EncodeTime(buf, t)
				continue
			}
			maps[field.Key] = field.Value
		}

		if err := json.NewEncoder(buf).Encode(maps); err != nil {
			panic(fmt.Errorf("<klog.StdJSONEncoder:Error>:%s", err.Error()))
		}

		// Append a newline
		buf.WriteByte('\n')
	})
}
