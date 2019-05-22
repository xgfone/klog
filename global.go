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

// Std is the global default Logger.
var Std = New()

// L emits a customized level log.
func L(level Level) Log { return Std.L(level) }

// Trace emits a TRACE log.
func Trace() Log { return Std.Trace() }

// Debug emits a DEBUG log.
func Debug() Log { return Std.Debug() }

// Info emits a INFO log.
func Info() Log { return Std.Info() }

// Warn emits a WARN log.
func Warn() Log { return Std.Warn() }

// Error emits a ERROR log.
func Error() Log { return Std.Error() }

// Panic emits a PANIC log.
func Panic() Log { return Std.Panic() }

// Fatal emits a FATAL log.
func Fatal() Log { return Std.Fatal() }
