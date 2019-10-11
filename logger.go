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
	"io"
	"os"
	"path/filepath"
)

// KV represents a key-value interface.
type KV interface {
	Key() string
	Value() interface{}
}

// Field represents a key-value pair.
type Field struct {
	Key   string
	Value interface{}
}

// NewErrField returns a Field with the key "err" and the value err.
func NewErrField(err error) Field {
	return Field{Key: "err", Value: err}
}

// Logger is a structured logger based on key-value.
type Logger struct {
	name    string
	depth   int
	level   Level
	writer  Writer
	encoder Encoder
	fields  []Field
	hooks   []Hook
}

// New returns a new Logger with the writer w that is os.Stdout by default.
//
// The default level is `LvlDebug`, and the default encoder is `TextEncoder()`.
func New(w ...Writer) Logger {
	out := SafeWriter(StreamWriter(os.Stdout))
	if len(w) > 0 && w[0] != nil {
		out = w[0]
	}
	return Logger{writer: out, level: LvlDebug, encoder: TextEncoder(true)}
}

// NewSimpleLogger returns a new Logger based on the file writer by using
// `SizedRotatingFile`.
//
// If filePath is "", it will ignore fileSize and fileNum, and use os.Stdout
// as the writer. If fileSize is "", it is "100M" by default. And fileNum is
// 100 by default.
//
// Notice: if the directory in where filePath is does not exist, it will be
// created automatically.
func NewSimpleLogger(level, filePath, fileSize string, fileNum int) (Logger, error) {
	lvl := NameToLevel(level)
	if filePath == "" {
		return New().WithLevel(lvl), nil
	}

	if fileSize == "" {
		fileSize = "100M"
	}
	size, err := ParseSize(fileSize)
	if err != nil {
		return Logger{}, err
	} else if fileNum < 1 {
		fileNum = 100
	}

	os.MkdirAll(filepath.Dir(filePath), 0755)
	file, err := NewSizedRotatingFile(filePath, int(size), fileNum)
	if err != nil {
		return Logger{}, err
	}
	AppendCleaner(func() { file.Close() })
	return New(StreamWriter(file)).WithLevel(lvl), nil
}

type loggerWriter struct {
	Logger
	level Level
}

func (lw loggerWriter) Write(msg []byte) (int, error) {
	if log := newLog(lw.Logger, lw.level, 0); log.ok {
		return emitLog(lw.Logger, lw.level, 1, string(msg), log.fields)
	}
	return 0, nil
}

// ToWriter converts the logger to io.Writer.
func (l Logger) ToWriter(lvl Level) io.Writer {
	return loggerWriter{Logger: l, level: lvl}
}

// IsEnabled reports whether the log with the lvl level can be logged.
func (l Logger) IsEnabled(lvl Level) bool {
	return lvl >= l.level
}

// AddDepth is the same as WithDepth(depth), but it will grow it with depth,
// not reset it to depth.
func (l Logger) AddDepth(depth int) Logger {
	l.depth += depth
	return l
}

// AddDepthSelf increases the depth of the caller of the logger and returns itself
// for chaining call.
func (l *Logger) AddDepthSelf(depth int) *Logger {
	if depth < 0 {
		panic("the log depth must not be less than 0")
	}
	l.depth += depth
	return l
}

// WithName returns a new Logger with the name.
func (l Logger) WithName(name string) Logger {
	l.name = name
	return l
}

// SetName resets the name of the logger and returns itself for chaining call.
func (l *Logger) SetName(name string) *Logger {
	l.name = name
	return l
}

// WithDepth returns a new Logger with the caller depth.
//
// Notice: 0 stands for the stack where the caller is.
func (l Logger) WithDepth(depth int) Logger {
	if depth < 0 {
		panic("the log depth must not be less than 0")
	}
	l.depth = depth
	return l
}

// SetDepth resets the depth of the caller of the logger and returns itself
// for chaining call.
func (l *Logger) SetDepth(depth int) *Logger {
	if depth < 0 {
		panic("the log depth must not be less than 0")
	}
	l.depth = depth
	return l
}

// WithLevel returns a new Logger with the level.
func (l Logger) WithLevel(level Level) Logger {
	if level < 0 {
		panic("the log level must not be less than 0")
	}
	l.level = level
	return l
}

// WithLevelString is equal to l.WithLevel(NameToLevel(level)).
func (l Logger) WithLevelString(level string) Logger {
	return l.WithLevel(NameToLevel(level))
}

// SetLevel resets the level of the logger and returns itself for chaining call.
func (l *Logger) SetLevel(level Level) *Logger {
	if level < 0 {
		panic("the log level must not be less than 0")
	}
	l.level = level
	return l
}

// SetLevelString is equal to l.SetLevel(NameToLevel(level)).
func (l *Logger) SetLevelString(level string) *Logger {
	return l.SetLevel(NameToLevel(level))
}

// WithWriter returns a new Logger with the writer w.
func (l Logger) WithWriter(w Writer) Logger {
	if w == nil {
		panic("the log writer must not be nil")
	}
	l.writer = w
	return l
}

// SetWriter resets the writer of the logger and returns itself for chaining call.
func (l *Logger) SetWriter(w Writer) *Logger {
	if w == nil {
		panic("the log writer must not be nil")
	}
	l.writer = w
	return l
}

// WithEncoder returns a new Logger with the encoder.
func (l Logger) WithEncoder(encoder Encoder) Logger {
	if encoder == nil {
		panic("the log encoder must not be nil")
	}
	l.encoder = encoder
	return l
}

// SetEncoder resets the encoder of the logger and returns itself for chaining call.
func (l *Logger) SetEncoder(encoder Encoder) *Logger {
	if encoder == nil {
		panic("the log encoder must not be nil")
	}
	l.encoder = encoder
	return l
}

// WithHook returns a new Logger with the hook, which will append the hook.
func (l Logger) WithHook(hook ...Hook) Logger {
	l.hooks = append(l.hooks, hook...)
	return l
}

// AddHook adds the hooks of the logger and returns itself for chaining call.
func (l *Logger) AddHook(hooks ...Hook) *Logger {
	l.hooks = append(l.hooks, hooks...)
	return l
}

// WithField returns a new Logger with the new field context.
func (l Logger) WithField(fields ...Field) Logger {
	l.fields = append(l.fields, fields...)
	return l
}

// AddField adds the fields of the logger and returns itself for chaining call.
func (l *Logger) AddField(fields ...Field) *Logger {
	l.fields = append(l.fields, fields...)
	return l
}

// WithKv returns a new Logger with the new key-value context, which is equal to
//
//   l.WithField(Field{Key: key, Value: value})
//
func (l Logger) WithKv(key string, value interface{}) Logger {
	return l.WithField(Field{Key: key, Value: value})
}

// AddKv adds the key-value as the field context, which is equal to
//
//   l.AddField(Field{Key: key, Value: value})
//
// It returns itself for chaining call.
func (l *Logger) AddKv(key string, value interface{}) *Logger {
	return l.AddField(Field{Key: key, Value: value})
}

// GetName returns the name of the logger.
func (l Logger) GetName() string {
	return l.name
}

// GetDepth returns the depth of the caller stack.
func (l Logger) GetDepth() int {
	return l.depth
}

// GetLevel returns the level of the logger.
func (l Logger) GetLevel() Level {
	return l.level
}

// GetWriter returns the writer of the logger.
func (l Logger) GetWriter() Writer {
	return l.writer
}

// GetEncoder returns the encoder of the logger.
func (l Logger) GetEncoder() Encoder {
	return l.encoder
}

// GetHooks returns the hooks of the logger.
func (l Logger) GetHooks() []Hook {
	return l.hooks
}

func (l Logger) emit(level Level, fields ...Field) Log {
	return newLog(l, level, l.depth).F(fields...)
}

// Level emits a specified `level` log.
//
// You can gives some key-value field contexts optionally, which is equal to
// call `Log.F(fields...)`.
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Level(level Level, fields ...Field) Log { return l.emit(level, fields...) }

// L is short for Level(level, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) L(level Level, fields ...Field) Log { return l.emit(level, fields...) }

// Trace is equal to l.Level(LvlTrace, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Trace(fields ...Field) Log { return l.emit(LvlTrace, fields...) }

// Debug is equal to l.Level(LvlDebug, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Debug(fields ...Field) Log { return l.emit(LvlDebug, fields...) }

// Info is equal to l.Level(LvlInfo, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Info(fields ...Field) Log { return l.emit(LvlInfo, fields...) }

// Warn is equal to l.Level(LvlWarn, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Warn(fields ...Field) Log { return l.emit(LvlWarn, fields...) }

// Error is equal to l.Level(LvlError, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Error(fields ...Field) Log { return l.emit(LvlError, fields...) }

// Panic is equal to l.Level(LvlPanic, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Panic(fields ...Field) Log { return l.emit(LvlPanic, fields...) }

// Fatal is equal to l.Level(LvlFatal, fields...).
//
// Notice: you must continue to call Msg() or Msgf() to trigger it.
func (l Logger) Fatal(fields ...Field) Log { return l.emit(LvlFatal, fields...) }

// F appends a key-value field log and returns another log interface
// to emit the level log.
//
// Notice: you must continue to call the level method, such as Levelf(),
// Debugf(), Infof(), Errorf(), etc, to trigger it.
func (l Logger) F(fields ...Field) LLog {
	return newLLog(l, l.depth, fields...)
}

// K is equal to l.F(Field{Key: key, Value: value}).
//
// Notice: you must continue to call the level method, such as Levelf(),
// Debugf(), Infof(), Errorf(), etc, to trigger it.
func (l Logger) K(key string, value interface{}) LLog {
	return l.F(Field{Key: key, Value: value})
}

// V is equal to F, which will convert the type KV to Field.
func (l Logger) V(kvs ...KV) LLog {
	switch len(kvs) {
	case 0:
		return newLLog(l, l.depth)
	case 1:
		return l.K(kvs[0].Key(), kvs[0].Value())
	default:
		return l.K(kvs[0].Key(), kvs[0].Value()).V(kvs[1:]...)
	}
}

// E is equal to l.F(NewErrField(err)).
//
// Notice: you must continue to call the level method, such as Levelf(),
// Debugf(), Infof(), Errorf(), etc, to trigger it.
func (l Logger) E(err error) LLog {
	return l.F(NewErrField(err))
}

// Levelf emits a specified `level` log, which is equal to l.F().Levelf(level, msg, args...).
func (l Logger) Levelf(level Level, msg string, args ...interface{}) {
	newLLog(l, l.depth+1).Levelf(level, msg, args...)
}

// Ef is equal to l.E(err).Errorf(msg, args...). If err is nil, however,
// it is eqaul to l.Infof(msg, args...).
func (l Logger) Ef(err error, msg string, args ...interface{}) {
	if err == nil {
		newLLog(l, l.depth+1).Infof(msg, args...)
	} else {
		newLLog(l, l.depth+1, NewErrField(err)).Errorf(msg, args...)
	}
}

// Lf is short for l.Levelf(level, msg, args...).
func (l Logger) Lf(level Level, msg string, args ...interface{}) {
	newLLog(l, l.depth+1).Levelf(level, msg, args...)
}

// Tracef emits a TRACE log, which is equal to l.Levelf(LvlTrace, msg, args...).
func (l Logger) Tracef(msg string, args ...interface{}) { newLLog(l, l.depth+1).Tracef(msg, args...) }

// Debugf emits a DEBUG log, which is equal to l.Levelf(LvlDebug, msg, args...).
func (l Logger) Debugf(msg string, args ...interface{}) { newLLog(l, l.depth+1).Debugf(msg, args...) }

// Infof emits a INFO log, which is equal to l.Levelf(LvlInfo, msg, args...).
func (l Logger) Infof(msg string, args ...interface{}) { newLLog(l, l.depth+1).Infof(msg, args...) }

// Printf is the alias of Infof.
func (l Logger) Printf(msg string, args ...interface{}) { newLLog(l, l.depth+1).Infof(msg, args...) }

// Warnf emits a WARN log, which is equal to l.Levelf(LvlWarn, msg, args...).
func (l Logger) Warnf(msg string, args ...interface{}) { newLLog(l, l.depth+1).Warnf(msg, args...) }

// Errorf emits a ERROR log, which is equal to l.Levelf(LvlError, msg, args...).
func (l Logger) Errorf(msg string, args ...interface{}) { newLLog(l, l.depth+1).Errorf(msg, args...) }

// Panicf emits a PANIC log, which is equal to l.Levelf(LvlPanic, msg, args...).
func (l Logger) Panicf(msg string, args ...interface{}) { newLLog(l, l.depth+1).Panicf(msg, args...) }

// Fatalf emits a FATAL log, which is equal to l.Levelf(LvlFatal, msg, args...).
func (l Logger) Fatalf(msg string, args ...interface{}) { newLLog(l, l.depth+1).Fatalf(msg, args...) }
