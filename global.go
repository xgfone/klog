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

// Std is the default global Logger.
var Std = New()

// DefaultManager is the default global manager of the logger.
var DefaultManager = NewManager()

// GetLogger is equal to DefaultManager.GetLogger(name).
func GetLogger(name string) *Logger {
	return DefaultManager.GetLogger(name)
}

////////////////////////////////////////////////////////////////////////////
/// Setter Interface

// AddKv is equal to Std.AddKv(key, value).
func AddKv(key string, value interface{}) *Logger {
	return Std.AddKv(key, value)
}

// AddField is equal to Std.AddField(fields...).
func AddField(fields ...Field) *Logger {
	return Std.AddField(fields...)
}

// AddHook is equal to Std.AddHook(hooks...).
func AddHook(hooks ...Hook) *Logger {
	return Std.AddHook(hooks...)
}

// AddDepth is equal to Std.AddDepthSelf(depth).
func AddDepth(depth int) *Logger {
	return Std.AddDepthSelf(depth)
}

// SetDepth is equal to Std.SetDepth(depth).
func SetDepth(depth int) *Logger {
	return Std.SetDepth(depth)
}

// SetEncoder is equal to Std.SetEncoder(encoder).
func SetEncoder(encoder Encoder) *Logger {
	return Std.SetEncoder(encoder)
}

// SetWriter is equal to Std.SetWriter(w).
func SetWriter(w Writer) *Logger {
	return Std.SetWriter(w)
}

// SetLevel is equal to Std.SetLevel(level).
func SetLevel(level Level) *Logger {
	return Std.SetLevel(level)
}

// SetLevelString is equal to Std.SetLevelString(level).
func SetLevelString(level string) *Logger {
	return Std.SetLevelString(level)
}

// SetName is equal to Std.SetName(name).
func SetName(name string) *Logger {
	return Std.SetName(name)
}

////////////////////////////////////////////////////////////////////////////
/// For LLog interface

// E is equal to Std.K("err", err).
func E(err error) LLog { return Std.K("err", err) }

// F is equal to Std.F(fields...).
func F(fields ...Field) LLog { return Std.F(fields...) }

// K is equal to Std.K(key, value).
func K(key string, value interface{}) LLog { return Std.K(key, value) }

// V is equal to Std.V(kvs...).
func V(kvs ...KV) LLog { return Std.V(kvs...) }

// Ef is equal to Std.Ef(err, format, args...).
func Ef(err error, format string, args ...interface{}) {
	Std.AddDepth(1).Ef(err, format, args...)
}

// Lf is equal to `Std.Lf(level, msg, args...)` to emit a specified `level` log.
func Lf(level Level, msg string, args ...interface{}) { Std.AddDepth(1).Lf(level, msg, args...) }

// Tracef equal to `Std.Tracef(level, msg, args...)` to emit a TRACE log.
func Tracef(msg string, args ...interface{}) { Std.AddDepth(1).Tracef(msg, args...) }

// Debugf equal to `Std.Debugf(level, msg, args...)` to emit a DEBUG log.
func Debugf(msg string, args ...interface{}) { Std.AddDepth(1).Debugf(msg, args...) }

// Infof equal to `Std.Infof(level, msg, args...)` to emit a INFO log.
func Infof(msg string, args ...interface{}) { Std.AddDepth(1).Infof(msg, args...) }

// Warnf equal to `Std.Warnf(level, msg, args...)` to emit a WARN log.
func Warnf(msg string, args ...interface{}) { Std.AddDepth(1).Warnf(msg, args...) }

// Errorf equal to `Std.Errorf(level, msg, args...)` to emit a ERROR log.
func Errorf(msg string, args ...interface{}) { Std.AddDepth(1).Errorf(msg, args...) }

// Panicf equal to `Std.Panicf(level, msg, args...)` to emit a PANIC log.
func Panicf(msg string, args ...interface{}) { Std.AddDepth(1).Panicf(msg, args...) }

// Fatalf equal to `Std.Fatalf(level, msg, args...)` to emit a FATAL log.
func Fatalf(msg string, args ...interface{}) { Std.AddDepth(1).Fatalf(msg, args...) }

////////////////////////////////////////////////////////////////////////////
/// For Log interface

// L is equal to `Std.L(level, fields...)` to emit a specified `level` log.
func L(level Level, fields ...Field) Log { return Std.L(level, fields...) }

// Trace is equal to `Std.Trace(fields...)` to emit a TRACE log.
func Trace(fields ...Field) Log { return Std.Trace(fields...) }

// Debug is equal to `Std.Debug(fields...)` to emit a DEBUG log.
func Debug(fields ...Field) Log { return Std.Debug(fields...) }

// Info is equal to `Std.Info(fields...)` to emit a INFO log.
func Info(fields ...Field) Log { return Std.Info(fields...) }

// Warn is equal to `Std.Warn(fields...)` to emit a WARN log.
func Warn(fields ...Field) Log { return Std.Warn(fields...) }

// Error is equal to `Std.Error(fields...)` to emit a ERROR log.
func Error(fields ...Field) Log { return Std.Error(fields...) }

// Panic is equal to `Std.Panic(fields...)` to emit a PANIC log.
func Panic(fields ...Field) Log { return Std.Panic(fields...) }

// Fatal is equal to `Std.Fatal(fields...)` to emit a FATAL log.
func Fatal(fields ...Field) Log { return Std.Fatal(fields...) }
