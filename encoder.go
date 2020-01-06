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
	"fmt"
	"strings"
	"time"
	"unicode"
)

// Field represents a key-value pair.
type Field interface {
	Key() string
	Value() interface{}
}

type field struct {
	key   string
	value interface{}
}

func (f field) Key() string { return f.key }
func (f field) Value() interface{} {
	switch v := f.value.(type) {
	case func() interface{}:
		return v()
	case func() string:
		return v()
	default:
		return v
	}
}

// E is equal to F("err", err).
func E(err error) Field { return field{key: "err", value: err} }

// F returns a new Field. If value is "func() interface{}" or "func() string",
// it will be evaluated when the log is emitted, that's, it is lazy.
func F(key string, value interface{}) Field { return field{key: key, value: value} }

// Record represents a log record.
type Record struct {
	Name  string    // The logger name, which may be empty
	Time  time.Time // The start time when to emit the log
	Depth int       // The stack depth of the caller

	Lvl    Level   // The log level
	Msg    string  // The log message
	Ctxs   []Field // The SHARED key-value contexts. DON'T MODIFY IT!
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
	Quote   bool
	Newline bool

	TimeKey string
	TimeFmt string

	LevelKey  string
	LoggerKey string
}

func getOption(options ...EncoderOption) (o option) {
	o.Newline = true
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

// EncodeTime enables the encoder to encode the time as the format with the key,
// which will encode the time as the integer second if format is missing.
func EncodeTime(key string, format ...string) EncoderOption {
	return func(o *option) {
		o.TimeKey = key
		if len(format) > 0 {
			o.TimeFmt = format[0]
		}
	}
}

// EncodeLevel enables the encoder to encode the level with the key.
func EncodeLevel(key string) EncoderOption {
	return func(o *option) { o.LevelKey = key }
}

// EncodeLogger enables the encoder to encode the logger name with the key.
func EncodeLogger(key string) EncoderOption {
	return func(o *option) { o.LoggerKey = key }
}

// Newline enables the encoder whether or not to append a newline.
//
// It is true by default.
func Newline(newline bool) EncoderOption {
	return func(o *option) { o.Newline = newline }
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

func encodeTime(buf *Builder, t time.Time, format string) {
	if t.IsZero() {
		t = time.Now()
	}

	if format == "" {
		buf.AppendInt(t.Unix())
	} else {
		buf.AppendTime(t, format)
	}
}

func textEncodeFields(buf *Builder, fields []Field, depth int, quote bool, timeFmt string) {
	depth++
	for _, field := range fields {
		buf.WriteString(field.Key())
		buf.WriteByte('=')

		var value interface{}
		if s, ok := field.(FieldStack); ok {
			value = s.Stack(depth)
		} else {
			value = field.Value()
		}

		switch v := value.(type) {
		case string:
			appendString(buf, v, quote)
		case error:
			if v == nil {
				buf.AppendAny(nil)
			} else {
				appendString(buf, v.Error(), quote)
			}
		case time.Time:
			encodeTime(buf, v, timeFmt)
		case fmt.Stringer:
			appendString(buf, v.String(), quote)
		default:
			if err := buf.AppendAnyFmt(v); err != nil {
				buf.WriteString("<klog.TextEncoder:Error:")
				buf.WriteString(err.Error())
				buf.WriteString(">")
			}
		}

		buf.WriteByte(' ')
	}
}

// TextEncoder encodes the key-values log as the text.
//
// Notice: The message will use "msg" as the key.
func TextEncoder(w Writer, options ...EncoderOption) Encoder {
	opt := getOption(options...)
	return EncoderFunc(w, func(buf *Builder, r Record) {
		r.Depth++

		// Time
		if opt.TimeKey != "" {
			buf.WriteString(opt.TimeKey)
			buf.WriteByte('=')
			encodeTime(buf, r.Time, opt.TimeFmt)
			buf.WriteByte(' ')
		}

		// Logger Name
		if r.Name != "" && opt.LoggerKey != "" {
			buf.WriteString(opt.LoggerKey)
			buf.WriteByte('=')
			appendString(buf, r.Name, opt.Quote)
			buf.WriteByte(' ')
		}

		// Level
		if opt.LevelKey != "" {
			buf.WriteString(opt.LevelKey)
			buf.WriteByte('=')
			buf.WriteString(r.Lvl.String())
			buf.WriteByte(' ')
		}

		// Ctxs and Fields
		textEncodeFields(buf, r.Ctxs, r.Depth, opt.Quote, opt.TimeFmt)
		textEncodeFields(buf, r.Fields, r.Depth, opt.Quote, opt.TimeFmt)

		// Message
		buf.WriteString("msg=")
		appendString(buf, r.Msg, opt.Quote)

		if opt.Newline {
			buf.WriteByte('\n')
		}
	})
}

func jsonEncodeFields(buf *Builder, fields []Field, depth int, timeFmt string) {
	depth++
	for _, field := range fields {
		// Key
		buf.AppendJSONString(field.Key())
		buf.WriteString(`:`)

		var value interface{}
		if s, ok := field.(FieldStack); ok {
			value = s.Stack(depth)
		} else {
			value = field.Value()
		}

		// Value
		switch v := value.(type) {
		case time.Time:
			buf.WriteByte('"')
			encodeTime(buf, v, timeFmt)
			buf.WriteByte('"')
		default:
			if err := buf.AppendJSON(v); err != nil {
				buf.AppendJSONString(fmt.Sprintf(`<klog.TextEncoder:Error:%s>`, err.Error()))
			}
		}

		buf.WriteByte(',')
	}
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
		if opt.TimeKey != "" {
			buf.WriteByte('"')
			buf.WriteString(opt.TimeKey)
			buf.WriteString(`":"`)
			encodeTime(buf, r.Time, opt.TimeFmt)
			buf.WriteString(`",`)
		}

		// Logger Name
		if r.Name != "" && opt.LoggerKey != "" {
			buf.WriteByte('"')
			buf.WriteString(opt.LoggerKey)
			buf.WriteString(`":`)
			buf.AppendJSONString(r.Name)
			buf.WriteByte(',')
		}

		// Level
		if opt.LevelKey != "" {
			buf.WriteByte('"')
			buf.WriteString(opt.LevelKey)
			buf.WriteString(`":"`)
			buf.WriteString(r.Lvl.String())
			buf.WriteString(`",`)
		}

		// Ctxs and Fields
		jsonEncodeFields(buf, r.Ctxs, r.Depth, opt.TimeFmt)
		jsonEncodeFields(buf, r.Fields, r.Depth, opt.TimeFmt)

		// Message
		buf.WriteString(`"msg":`)
		buf.AppendJSONString(r.Msg)

		// End
		buf.WriteByte('}')

		if opt.Newline {
			buf.WriteByte('\n')
		}
	})
}
