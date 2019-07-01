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

// Encoder is the encoder of the log record.
//
// Notice: w is the buffer writer.
type Encoder func(buf *Builder, r Record)

// NothingEncoder encodes nothing.
func NothingEncoder() Encoder { return func(buf *Builder, r Record) {} }

// TextEncoder encodes the key-values log as the text.
//
// Notice: the key name of the level is "lvl", that of the time is "t"
// with time.RFC3339Nano, and that of the message is "msg".
// If the logger name exists, it will encode it and the key name is "logger".
//
// If quote is true and the string value contains the space, it will be surrounded
// by a pair of the double quotation marks.
func TextEncoder(quote ...bool) Encoder {
	var _quote bool
	if len(quote) > 0 && quote[0] {
		_quote = true
	}

	appendString := func(buf *Builder, s string) {
		if _quote && strings.IndexFunc(s, unicode.IsSpace) > -1 {
			buf.AppendJSONString(s)
		} else {
			buf.AppendString(s)
		}
	}

	return func(buf *Builder, r Record) {
		// Time
		buf.WriteString("t=")
		buf.AppendTime(r.Time, time.RFC3339Nano)

		// Logger Name
		if r.Name != "" {
			buf.WriteString(" logger=")
			appendString(buf, r.Name)
		}

		// Level
		buf.WriteString(" lvl=")
		buf.WriteString(r.Level.String())

		// Fields
		for _, field := range r.Fields {
			buf.WriteString(" ")
			buf.WriteString(field.Key)
			buf.WriteString("=")

			switch v := field.Value.(type) {
			case string:
				appendString(buf, v)
			case error:
				if v == nil {
					buf.AppendAny(nil)
				} else {
					appendString(buf, v.Error())
				}
			case time.Time:
				buf.AppendTime(v, time.RFC3339Nano)
			case fmt.Stringer:
				appendString(buf, v.String())
			default:
				if err := buf.AppendAnyFmt(field.Value); err != nil {
					buf.WriteString("<klog.TextEncoder:Error:")
					buf.WriteString(err.Error())
					buf.WriteString(">")
				}
			}
		}

		// Message
		buf.WriteString(" msg=")
		appendString(buf, r.Msg)
		buf.WriteByte('\n')
	}
}

// JSONEncoder encodes the key-values log as json.
//
// Notice: the key name of the level is "lvl" and that of the time is "t"
// with time.RFC3339Nano, and that of the message is "msg".
// If the logger name exists, it will encode it and the key name is "logger".
func JSONEncoder() Encoder {
	return func(buf *Builder, r Record) {
		// Start and Time
		buf.WriteString(`{"t":"`)
		buf.AppendTime(r.Time, time.RFC3339Nano)
		buf.WriteString(`"`)

		// Logger Name
		if r.Name != "" {
			buf.WriteString(`,"logger":`)
			buf.AppendJSONString(r.Name)
		}

		// Level
		buf.WriteString(`,"lvl":`)
		buf.AppendJSONString(r.Level.String())

		// Fields
		for _, field := range r.Fields {
			// Key
			buf.WriteString(`,`)
			buf.AppendJSONString(field.Key)
			buf.WriteString(`:`)

			// Value
			if err := buf.AppendJSON(field.Value); err != nil {
				buf.AppendJSONString(fmt.Sprintf(`<klog.TextEncoder:Error:%s>`, err.Error()))
			}
		}

		// Message
		buf.WriteString(`,"msg":`)
		buf.AppendJSONString(r.Msg)

		// End
		buf.WriteString("}\n")
	}
}

// StdJSONEncoder is equal to JSONEncoder, which uses json.Marshal() to encode
// it, but the performance is a little bad.
func StdJSONEncoder() Encoder {
	return func(buf *Builder, r Record) {
		maps := make(map[string]interface{}, len(r.Fields)+8)
		maps["t"] = r.Time.Format(time.RFC3339Nano)
		maps["lvl"] = r.Level.String()
		maps["msg"] = r.Msg

		if r.Name != "" {
			maps["logger"] = r.Name
		}

		for _, field := range r.Fields {
			maps[field.Key] = field.Value
		}

		data, err := json.Marshal(maps)
		if err != nil {
			panic(fmt.Errorf("<klog.StdJSONEncoder:Error>:%s", err.Error()))
		}
		buf.Write(data)

		// Append a newline
		buf.WriteByte('\n')
	}
}
