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
	"sync"
	"time"
)

var (
	fieldPool   = sync.Pool{New: func() interface{} { return make([]Field, 0, 16) }}
	builderPool = sync.Pool{New: func() interface{} { return NewBuilder(1024) }}
)

func getBuilder() *Builder  { return builderPool.Get().(*Builder) }
func putBuilder(b *Builder) { b.Reset(); builderPool.Put(b) }

// Record represents a log record.
type Record struct {
	Msg    string    // The log message
	Time   time.Time // The start time when to emit the log
	Name   string    // The logger name
	Depth  int       // The depth of the caller
	Level  Level     // The log level
	Fields []Field   // The key-value logs
}

// Log is used to emit a structured key-value log.
type Log struct {
	fields []Field
	logger Logger
	level  Level
	depth  int
	ok     bool
}

func newLog(logger Logger, level Level, depth int) Log {
	ok := true
	for _, hook := range logger.hooks {
		if !hook(logger.name, level) {
			ok = false
		}
	}
	if ok && level < logger.level {
		ok = false
	}

	var fields []Field
	if ok || len(logger.fields) > 0 {
		fields = append(fieldPool.Get().([]Field), logger.fields...)
	}
	return Log{fields: fields, logger: logger, level: level, depth: depth, ok: ok}
}

// K appends the key-value pair into the structured log.
func (l Log) K(key string, value interface{}) Log {
	return l.F(Field{Key: key, Value: value})
}

// F appends more than one key-value pair into the structured log by the field.
func (l Log) F(fields ...Field) Log {
	if l.ok {
		l.fields = append(l.fields, fields...)
	}
	return l
}

// Print is equal to l.Msg(args...).
func (l Log) Print(args ...interface{}) {
	l.depth++
	l.Msg(args...)
}

// Printf is equal to l.Msgf(format, args...).
func (l Log) Printf(format string, args ...interface{}) {
	l.depth++
	l.Msgf(format, args...)
}

// Msg appends the msg into the structured log with the key "msg" at last.
//
// Notice: "args" will be formatted by `fmt.Sprint(args...)`.
func (l Log) Msg(args ...interface{}) {
	if !l.ok {
		return
	}

	switch len(args) {
	case 0:
		l.emit("")
	case 1:
		switch v := args[0].(type) {
		case string:
			l.emit(v)
		default:
			l.emit(fmt.Sprint(v))
		}
	default:
		l.emit(fmt.Sprint(args...))
	}
}

// Msgf appends the msg into the structured log with the key "msg" at last.
//
// Notice: "format" and "args" will be foramtted by `fmt.Sprintf(format, args...)`.
func (l Log) Msgf(format string, args ...interface{}) {
	if !l.ok {
		return
	}

	if len(args) == 0 {
		l.emit(format)
	}
	l.emit(fmt.Sprintf(format, args...))
}

func (l Log) emit(msg string) {
	record := Record{
		Msg:    msg,
		Time:   time.Now(),
		Name:   l.logger.name,
		Level:  l.level,
		Depth:  l.depth + 2,
		Fields: l.fields,
	}

	for i, field := range l.fields {
		switch v := field.Value.(type) {
		case Valuer:
			l.fields[i].Value = v(record)
		case func(Record) interface{}:
			l.fields[i].Value = v(record)
		}
	}

	buf := getBuilder()
	l.logger.encoder(buf, record)
	if bs := buf.Bytes(); len(bs) > 0 {
		l.logger.writer.Write(l.level, bs)
	}
	fieldPool.Put(l.fields[:0])
	putBuilder(buf)
}
