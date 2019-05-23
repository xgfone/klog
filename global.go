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

var cleaners []func()

// AppendCleaner appends the clean functions, which will be called when emitting
// the FATAL log.
func AppendCleaner(clean ...func()) {
	cleaners = append(cleaners, clean...)
}

// Std is the global default Logger.
var Std = New()

// F is equal to Std.F(fields...).
func F(fields ...Field) LLog { return Std.F(fields...) }

// K is equal to Std.K(key, value).
func K(key string, value interface{}) LLog { return Std.K(key, value) }

// L emits a customized level log.
func L(level Level, fields ...Field) Log { return Std.L(level, fields...) }

// Trace emits a TRACE log.
func Trace(fields ...Field) Log { return Std.Trace(fields...) }

// Debug emits a DEBUG log.
func Debug(fields ...Field) Log { return Std.Debug(fields...) }

// Info emits a INFO log.
func Info(fields ...Field) Log { return Std.Info(fields...) }

// Warn emits a WARN log.
func Warn(fields ...Field) Log { return Std.Warn(fields...) }

// Error emits a ERROR log.
func Error(fields ...Field) Log { return Std.Error(fields...) }

// Panic emits a PANIC log.
func Panic(fields ...Field) Log { return Std.Panic(fields...) }

// Fatal emits a FATAL log.
func Fatal(fields ...Field) Log { return Std.Fatal(fields...) }
