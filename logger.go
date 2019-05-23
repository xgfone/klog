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
	"os"
)

// Field represents a key-value pair.
type Field struct {
	Key   string
	Value interface{}
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
	out := StreamWriter(os.Stdout)
	if len(w) > 0 && w[0] != nil {
		out = w[0]
	}
	return Logger{writer: out, level: LvlDebug, encoder: TextEncoder()}
}

// NewSimpleLogger returns a new Logger based on the file writer by using
// `SizedRotatingFile`.
//
// If filePath is "", it will ignore fileSize and fileNum, and use os.Stdout
// as the writer. If fileSize is "", it is "100M" by default. And fileNum is
// 100 by default.
func NewSimpleLogger(level, filePath, fileSize string, fileNum int) (Logger, error) {
	lvl := NameToLevel(level)
	if filePath == "" {
		return New().WithLevel(lvl), nil
	}

	size, err := ParseSize(fileSize)
	if err != nil {
		return Logger{}, err
	} else if fileNum < 1 {
		fileNum = 100
	}

	file, err := NewSizedRotatingFile(filePath, int(size), fileNum)
	if err != nil {
		return Logger{}, err
	}
	AppendCleaner(func() { file.Close() })
	return New(StreamWriter(file)).WithLevel(lvl), nil
}

func (l Logger) clone() Logger {
	return Logger{
		name:    l.name,
		hooks:   l.hooks,
		depth:   l.depth,
		level:   l.level,
		writer:  l.writer,
		fields:  l.fields,
		encoder: l.encoder,
	}
}

// WithName returns a new Logger with the name.
func (l Logger) WithName(name string) Logger {
	logger := l.clone()
	logger.name = name
	return logger
}

// WithDepth returns a new Logger with the caller depth.
//
// Notice: 0 stands for the stack where the caller is.
func (l Logger) WithDepth(depth int) Logger {
	if depth < 0 {
		panic("the log depth must not be less than 0")
	}
	logger := l.clone()
	logger.depth = depth
	return logger
}

// WithLevel returns a new Logger with the level.
func (l Logger) WithLevel(level Level) Logger {
	if level < 0 {
		panic("the log level must not be less than 0")
	}
	logger := l.clone()
	logger.level = level
	return logger
}

// WithWriter returns a new Logger with the writer w.
func (l Logger) WithWriter(w Writer) Logger {
	if w == nil {
		panic("the log writer must not be nil")
	}
	logger := l.clone()
	logger.writer = w
	return logger
}

// WithEncoder returns a new Logger with the encoder.
func (l Logger) WithEncoder(encoder Encoder) Logger {
	if encoder == nil {
		panic("the log encoder must not be nil")
	}
	logger := l.clone()
	logger.encoder = encoder
	return logger
}

// WithHook returns a new Logger with the hook, which will append the hook.
func (l Logger) WithHook(hook ...Hook) Logger {
	logger := l.clone()
	logger.hooks = append(logger.hooks, hook...)
	return logger
}

// WithField returns a new Logger with the new field context.
func (l Logger) WithField(fields ...Field) Logger {
	logger := l.clone()
	logger.fields = append(logger.fields, fields...)
	return logger
}

// WithKv returns a new Logger with the new key-value context, which is equal to
//
//   l.WithField(Field{Key: key, Value: value})
//
func (l Logger) WithKv(key string, value interface{}) Logger {
	return l.WithField(Field{Key: key, Value: value})
}

// GetName returns the logger name.
func (l Logger) GetName() string {
	return l.name
}

// GetDepth returns the depth of the caller stack.
func (l Logger) GetDepth() int {
	return l.depth
}

// GetLevel returns the logger level.
func (l Logger) GetLevel() Level {
	return l.level
}

// GetWriter returns the logger writer.
func (l Logger) GetWriter() Writer {
	return l.writer
}

func (l Logger) emit(level Level, fields ...Field) Log {
	return newLog(l, level, l.depth).F(fields...)
}

// Level emits a log, the level of which is level.
//
// You can gives some key-value field contexts optionally, which is equal to
// call Log.F(fields...).
func (l Logger) Level(level Level, fields ...Field) Log {
	return l.emit(level, fields...)
}

// L is short for Level(level, fields...).
func (l Logger) L(level Level, fields ...Field) Log {
	return l.emit(level, fields...)
}

// Trace is equal to l.Level(LvlTrace, fields...).
func (l Logger) Trace(fields ...Field) Log {
	return l.emit(LvlTrace, fields...)
}

// Debug is equal to l.Level(LvlDebug, fields...).
func (l Logger) Debug(fields ...Field) Log {
	return l.emit(LvlDebug, fields...)
}

// Info is equal to l.Level(LvlInfo, fields...).
func (l Logger) Info(fields ...Field) Log {
	return l.emit(LvlInfo, fields...)
}

// Warn is equal to l.Level(LvlWarn, fields...).
func (l Logger) Warn(fields ...Field) Log {
	return l.emit(LvlWarn, fields...)
}

// Error is equal to l.Level(LvlError, fields...).
func (l Logger) Error(fields ...Field) Log {
	return l.emit(LvlError, fields...)
}

// Panic is equal to l.Level(LvlPanic, fields...).
func (l Logger) Panic(fields ...Field) Log {
	return l.emit(LvlPanic, fields...)
}

// Fatal is equal to l.Level(LvlFatal, fields...).
func (l Logger) Fatal(fields ...Field) Log {
	return l.emit(LvlFatal, fields...)
}
