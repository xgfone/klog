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
	// FieldSize is the default capacity of []Field to be created in the pool.
	FieldSize = 16

	// BuilderSize is the default capacity of Builder to be created in the pool.
	BuilderSize = 1024
)

var (
	fieldPool   = sync.Pool{New: func() interface{} { return make([]Field, 0, FieldSize) }}
	builderPool = sync.Pool{New: func() interface{} { return NewBuilder(BuilderSize) }}
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
	if ok && !logger.IsEnabled(level) {
		ok = false
	}

	var fields []Field
	if ok {
		fields = append(fieldPool.Get().([]Field), logger.fields...)
	}
	return Log{fields: fields, logger: logger, level: level, depth: depth, ok: ok}
}

// AddDepth adds the depth.
func (l Log) AddDepth(depth int) Log {
	l.depth += depth
	return l
}

// E is equal to l.F(NewErrField(err)).
func (l Log) E(err error) Log {
	return l.F(NewErrField(err))
}

// K appends the key-value pair into the structured log.
func (l Log) K(key string, value interface{}) Log {
	return l.F(Field{Key: key, Value: value})
}

// V is equal to F, which will convert the type KV to Field.
func (l Log) V(kvs ...KV) Log {
	if l.ok {
		for _, kv := range kvs {
			l.fields = append(l.fields, Field{Key: kv.Key(), Value: kv.Value()})
		}
	}
	return l
}

// F appends more than one key-value pair into the structured log by the field.
func (l Log) F(fields ...Field) Log {
	if l.ok {
		l.fields = append(l.fields, fields...)
	}
	return l
}

// Ef is short for l.E(err).Msgf(format, args...).
func (l Log) Ef(err error, format string, args ...interface{}) {
	l.AddDepth(1).E(err).Msgf(format, args...)
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
		return
	}
	l.emit(fmt.Sprintf(format, args...))
}

func (l Log) emit(msg string) {
	emitLog(l.logger, l.level, l.depth, msg, l.fields)
}

// emitLog emits the msg log with the context fields unconditionally,
// which won't filter the log by the logger name and the level.
func emitLog(logger Logger, level Level, depth int, msg string, fields []Field) (n int, err error) {
	record := Record{
		Msg:    msg,
		Time:   time.Now(),
		Name:   logger.name,
		Level:  level,
		Depth:  depth + 3,
		Fields: fields,
	}

	for i, field := range fields {
		switch v := field.Value.(type) {
		case Valuer:
			fields[i].Value = v(record)
		case func(Record) interface{}:
			fields[i].Value = v(record)
		}
	}

	buf := getBuilder()
	logger.encoder(buf, record)
	if bs := buf.Bytes(); len(bs) > 0 {
		n, err = logger.writer.Write(level, bs)
	}
	fieldPool.Put(fields[:0])
	putBuilder(buf)

	if level >= LvlFatal {
		exit(1)
	} else if level >= LvlPanic {
		record.Fields = nil
		panic(record)
	}

	return
}

/////////////////////////////////////////////////////////////////////////////

func (l LLog) emit(level Level, format string, args ...interface{}) {
	if !l.logger.IsEnabled(level) {
		return
	}

	ok := true
	for _, hook := range l.logger.hooks {
		if !hook(l.logger.name, level) {
			ok = false
		}
	}
	if !ok {
		return
	}

	if len(args) > 0 {
		format = fmt.Sprintf(format, args...)
	}

	emitLog(l.logger, level, l.depth, format, l.fields)
}

// LLog is another interface of the key-value log with the level.
type LLog struct {
	fields []Field
	logger Logger
	depth  int
}

func newLLog(logger Logger, depth int, fields ...Field) LLog {
	_fields := append(fieldPool.Get().([]Field), logger.fields...)
	_fields = append(_fields, fields...)
	return LLog{logger: logger, depth: depth, fields: _fields}
}

// IsEnabled reports whether the log with the lvl level can be logged.
func (l LLog) IsEnabled(lvl Level) bool {
	return l.logger.IsEnabled(lvl)
}

// AddDepth adds the depth.
func (l LLog) AddDepth(depth int) LLog {
	l.depth += depth
	return l
}

// E is equal to l.F(NewErrField(err)).
func (l LLog) E(err error) LLog {
	return l.F(NewErrField(err))
}

// K appends the key-value pair into the structured log.
func (l LLog) K(key string, value interface{}) LLog {
	return l.F(Field{Key: key, Value: value})
}

// V is equal to F, which will convert the type KV to Field.
func (l LLog) V(kvs ...KV) LLog {
	for _, kv := range kvs {
		l.fields = append(l.fields, Field{Key: kv.Key(), Value: kv.Value()})
	}
	return l
}

// F appends more than one key-value pair into the structured log by the field.
func (l LLog) F(fields ...Field) LLog {
	l.fields = append(l.fields, fields...)
	return l
}

// Ef is short for l.E(err).Errorf(format, args...). If err is nil, however,
// it is eqaul to l.Infof(format, args...).
func (l LLog) Ef(err error, format string, args ...interface{}) {
	if err == nil {
		l.AddDepth(1).Infof(format, args...)
	} else {
		l.AddDepth(1).E(err).Errorf(format, args...)
	}
}

// Lf is short for Level().
func (l LLog) Lf(lvl Level, msg string, args ...interface{}) { l.emit(lvl, msg, args...) }

// Levelf emits a customized level log.
func (l LLog) Levelf(lvl Level, msg string, args ...interface{}) { l.emit(lvl, msg, args...) }

// Tracef emits a TRACE log.
func (l LLog) Tracef(msg string, args ...interface{}) { l.emit(LvlTrace, msg, args...) }

// Debugf emits a DEBUG log.
func (l LLog) Debugf(msg string, args ...interface{}) { l.emit(LvlDebug, msg, args...) }

// Infof emits a INFO log.
func (l LLog) Infof(msg string, args ...interface{}) { l.emit(LvlInfo, msg, args...) }

// Warnf emits a WARN log.
func (l LLog) Warnf(msg string, args ...interface{}) { l.emit(LvlWarn, msg, args...) }

// Errorf emits a ERROR log.
func (l LLog) Errorf(msg string, args ...interface{}) { l.emit(LvlError, msg, args...) }

// Panicf emits a PANIC log.
func (l LLog) Panicf(msg string, args ...interface{}) { l.emit(LvlPanic, msg, args...) }

// Fatalf emits a FATAL log.
func (l LLog) Fatalf(msg string, args ...interface{}) { l.emit(LvlFatal, msg, args...) }
