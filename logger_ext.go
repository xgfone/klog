// Copyright 2020 xgfone
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
	"log"
	"os"
	"path/filepath"
	"time"
)

var fixDepth = func(depth int) int { return depth }

// ExtLogger is a extended logger implemented the Logger and Loggerf interface.
type ExtLogger struct {
	Name    string
	Ctxs    []Field
	Depth   int
	Level   Level
	Encoder Encoder
}

// New creates a new ExtLogger, which will use TextEncoder as the encoder
// and output the log to os.Stdout.
func New(name string) *ExtLogger {
	w := SafeWriter(StreamWriter(os.Stdout))
	e := TextEncoder(w, Quote(), EncodeLevel("lvl"), EncodeLogger("logger"),
		EncodeTime("t", time.RFC3339Nano))
	return &ExtLogger{Name: name, Level: LvlDebug, Encoder: e}
}

// NewSimpleLogger returns a new simple logger.
func NewSimpleLogger(name, level, filePath, fileSize string, fileNum int) (*ExtLogger, error) {
	log := New(name).WithLevel(NameToLevel(level))
	if filePath != "" {
		os.MkdirAll(filepath.Dir(filePath), 0755)
		wc, err := FileWriter(filePath, fileSize, fileNum)
		if err != nil {
			return nil, err
		}
		log.Encoder.SetWriter(SafeWriter(wc))
	}
	return log, nil
}

// StdLog converts the ExtLogger to the std log.
func (l *ExtLogger) StdLog(prefix string, flags ...int) *log.Logger {
	flag := log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	if len(flags) > 0 {
		flag = flags[0]
	}
	return log.New(ToIOWriter(l.Encoder.Writer(), l.Level), prefix, flag)
}

// Clone clones itself and returns a new one.
func (l *ExtLogger) Clone() *ExtLogger {
	var ctxs []Field
	if len(l.Ctxs) != 0 {
		ctxs = append([]Field{}, l.Ctxs...)
	}

	return &ExtLogger{
		Ctxs:    ctxs,
		Name:    l.Name,
		Depth:   l.Depth,
		Level:   l.Level,
		Encoder: l.Encoder,
	}
}

// WithName returns a new ExtLogger with the new name.
func (l *ExtLogger) WithName(name string) *ExtLogger {
	ll := l.Clone()
	ll.Name = name
	return ll
}

// WithLevel returns a new ExtLogger with the new level.
func (l *ExtLogger) WithLevel(level Level) *ExtLogger {
	ll := l.Clone()
	ll.Level = level
	return ll
}

// WithEncoder returns a new ExtLogger with the new encoder.
func (l *ExtLogger) WithEncoder(e Encoder) *ExtLogger {
	ll := l.Clone()
	ll.Encoder = e
	return ll
}

// WithDepth returns a new ExtLogger, which will increase the depth.
func (l *ExtLogger) WithDepth(depth int) *ExtLogger {
	ll := l.Clone()
	ll.Depth += depth
	return ll
}

// WithCtx returns a new ExtLogger with the new context fields.
func (l *ExtLogger) WithCtx(ctxs ...Field) *ExtLogger {
	ll := l.Clone()
	ll.Ctxs = append(ll.Ctxs, ctxs...)
	return ll
}

// Log emits the logs with the level and the depth.
func (l *ExtLogger) Log(lvl Level, depth int, msg string, args []interface{}, fields []Field) {
	if lvl < l.Level {
		return
	}

	if len(args) != 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	r := Record{
		Name:   l.Name,
		Depth:  l.Depth + 1 + fixDepth(depth),
		Lvl:    lvl,
		Msg:    msg,
		Ctxs:   l.Ctxs,
		Fields: fields,
	}
	l.Encoder.Encode(r)

	if lvl == LvlFatal {
		callOnExit()
		os.Exit(1)
	}
}

// Trace implements the interface Logger.
func (l *ExtLogger) Trace(msg string, fields ...Field) { l.Log(LvlTrace, 1, msg, nil, fields) }

// Debug implements the interface Logger.
func (l *ExtLogger) Debug(msg string, fields ...Field) { l.Log(LvlDebug, 1, msg, nil, fields) }

// Info implements the interface Logger.
func (l *ExtLogger) Info(msg string, fields ...Field) { l.Log(LvlInfo, 1, msg, nil, fields) }

// Warn implements the interface Logger.
func (l *ExtLogger) Warn(msg string, fields ...Field) { l.Log(LvlWarn, 1, msg, nil, fields) }

// Error implements the interface Logger.
func (l *ExtLogger) Error(msg string, fields ...Field) { l.Log(LvlError, 1, msg, nil, fields) }

// Fatal implements the interface Logger, but call the exit functions
// in CallOnExit before the program exits.
func (l *ExtLogger) Fatal(msg string, fields ...Field) { l.Log(LvlFatal, 1, msg, nil, fields) }

// Tracef implements the interface Loggerf.
func (l *ExtLogger) Tracef(msg string, args ...interface{}) { l.Log(LvlTrace, 1, msg, args, nil) }

// Debugf implements the interface Loggerf.
func (l *ExtLogger) Debugf(msg string, args ...interface{}) { l.Log(LvlDebug, 1, msg, args, nil) }

// Infof implements the interface Loggerf.
func (l *ExtLogger) Infof(msg string, args ...interface{}) { l.Log(LvlInfo, 1, msg, args, nil) }

// Warnf implements the interface Loggerf.
func (l *ExtLogger) Warnf(msg string, args ...interface{}) { l.Log(LvlWarn, 1, msg, args, nil) }

// Errorf implements the interface Loggerf.
func (l *ExtLogger) Errorf(msg string, args ...interface{}) { l.Log(LvlError, 1, msg, args, nil) }

// Fatalf implements the interface Loggerf, but call the exit functions
// in CallOnExit before the program exits.
func (l *ExtLogger) Fatalf(msg string, args ...interface{}) { l.Log(LvlFatal, 1, msg, args, nil) }

// Printf is equal to l.Infof(msg, args...).
func (l *ExtLogger) Printf(msg string, args ...interface{}) { l.Log(LvlInfo, 1, msg, args, nil) }
