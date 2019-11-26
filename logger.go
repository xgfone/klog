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
	"path/filepath"
	"time"
)

// Logger is an logger interface to emit the log.
type Logger interface {
	Log(level Level, msg string, fields ...Field)
}

// ExtLogger is a extended logger interface.
type ExtLogger interface {
	Logger

	SetLevel(level Level) // Reset the level
	SetEncoder(Encoder)   // Reset the encoder
	Encoder() Encoder     // Return the encoder of the logger

	WithCtx(fields ...Field) ExtLogger // Return a new Logger with the fields
	WithName(name string) ExtLogger    // Return a new Logger with the new name
	WithLevel(level Level) ExtLogger   // Return a new Logger with the level
	WithEncoder(e Encoder) ExtLogger   // Return a new Logger with the encoder
	WithDepth(depth int) ExtLogger     // Return a new Logger with the increased depth
}

// GetEncoderFromLogger returns the encoder from logger if it has the method
// `func Encoder() Encoder`. Or, it will return defaultEncoder if given or nil.
func GetEncoderFromLogger(logger Logger, defaultEncoder ...Encoder) Encoder {
	if l, ok := logger.(interface{ Encoder() Encoder }); ok {
		return l.Encoder()
	} else if len(defaultEncoder) > 0 {
		return defaultEncoder[0]
	}
	return nil
}

type logger struct {
	name    string
	depth   int
	level   Level
	fields  []Field
	encoder Encoder
}

// New creates a new ExtLogger, which will use TextEncoder as the encoder
// and output the log to os.Stdout.
func New(name string) ExtLogger {
	w := SafeWriter(StreamWriter(os.Stdout))
	e := TextEncoder(w, Quote(), EncodeLevel("lvl"), EncodeLogger("logger"), EncodeTime("t", time.RFC3339Nano))
	return &logger{name: name, level: LvlDebug, encoder: e, depth: 1}
}

func (l *logger) clone() *logger {
	return &logger{
		name:    l.name,
		depth:   l.depth,
		level:   l.level,
		fields:  l.fields,
		encoder: l.encoder,
	}
}

func (l *logger) Level() Level                    { return l.level }
func (l *logger) Encoder() Encoder                { return l.encoder }
func (l *logger) SetEncoder(enc Encoder)          { l.encoder = enc }
func (l *logger) SetLevel(level Level)            { l.level = level }
func (l *logger) WithName(name string) ExtLogger  { ll := l.clone(); ll.name = name; return ll }
func (l *logger) WithLevel(level Level) ExtLogger { ll := l.clone(); ll.level = level; return ll }
func (l *logger) WithEncoder(e Encoder) ExtLogger { ll := l.clone(); ll.encoder = e; return ll }
func (l *logger) WithDepth(depth int) ExtLogger   { ll := l.clone(); ll.depth += depth; return ll }
func (l *logger) WithCtx(fields ...Field) ExtLogger {
	ll := l.clone()
	ll.fields = append(ll.fields, fields...)
	return ll
}

func (l *logger) IsEnabled(lvl Level) bool { return lvl.Priority() >= l.level.Priority() }
func (l *logger) Log(lvl Level, msg string, fields ...Field) {
	if l.IsEnabled(lvl) {
		r := Record{
			Name:  l.name,
			Depth: l.depth,

			Lvl:    lvl,
			Msg:    msg,
			Ctxs:   l.fields,
			Fields: fields,
		}
		l.encoder.Encode(r)
	}
}

// NewSimpleLogger returns a new simple logger.
func NewSimpleLogger(name, level, filePath, fileSize string, fileNum int) (ExtLogger, error) {
	log := New(name).WithLevel(NameToLevel(level))
	if filePath != "" {
		os.MkdirAll(filepath.Dir(filePath), 0755)
		wc, err := FileWriter(filePath, fileSize, fileNum)
		if err != nil {
			return nil, err
		}
		log.Encoder().SetWriter(wc)
	}
	return log, nil
}
