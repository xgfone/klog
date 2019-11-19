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

import "log"

// LevelLogger is a convenient logger interface based on the level.
type LevelLogger interface {
	Trace(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Printfer is a Printf logger interface.
type Printfer interface {
	Printf(msg string, args ...interface{})
}

// FmtLogger is a formatter logger interface.
type FmtLogger interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// ToPrintfer converts ExtLogger to Printfer, which will use the INFO level.
func ToPrintfer(logger ExtLogger) Printfer {
	return flogger{logger.WithDepth(1)}
}

// ToFmtLogger converts ExtLogger to FmtLogger.
func ToFmtLogger(logger ExtLogger) FmtLogger {
	return flogger{logger.WithDepth(1)}
}

// ToLevelLogger converts ExtLogger to LevelLogger.
func ToLevelLogger(logger ExtLogger) LevelLogger {
	return flogger{logger.WithDepth(1)}
}

type flogger struct{ ExtLogger }

func (l flogger) Printf(fmt string, args ...interface{}) { l.Log(LvlInfo, Sprintf(fmt, args...)) }
func (l flogger) Tracef(fmt string, args ...interface{}) { l.Log(LvlTrace, Sprintf(fmt, args...)) }
func (l flogger) Debugf(fmt string, args ...interface{}) { l.Log(LvlDebug, Sprintf(fmt, args...)) }
func (l flogger) Infof(fmt string, args ...interface{})  { l.Log(LvlInfo, Sprintf(fmt, args...)) }
func (l flogger) Warnf(fmt string, args ...interface{})  { l.Log(LvlWarn, Sprintf(fmt, args...)) }
func (l flogger) Errorf(fmt string, args ...interface{}) { l.Log(LvlError, Sprintf(fmt, args...)) }
func (l flogger) Trace(msg string, fields ...Field)      { l.Log(LvlTrace, msg, fields...) }
func (l flogger) Debug(msg string, fields ...Field)      { l.Log(LvlDebug, msg, fields...) }
func (l flogger) Info(msg string, fields ...Field)       { l.Log(LvlInfo, msg, fields...) }
func (l flogger) Warn(msg string, fields ...Field)       { l.Log(LvlWarn, msg, fields...) }
func (l flogger) Error(msg string, fields ...Field)      { l.Log(LvlError, msg, fields...) }

// NewStdLogger returns a new log.Logger with the writer.
//
// If not giving flags, it is log.LstdFlags|log.Lmicroseconds|log.Lshortfile
// by default.
func NewStdLogger(w Writer, prefix string, flags ...int) *log.Logger {
	flag := log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	if len(flags) > 0 {
		flag = flags[0]
	}
	return log.New(ToIOWriter(w), prefix, flag)
}
