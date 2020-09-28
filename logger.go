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

// Field represents a key-value pair.
type Field interface {
	Key() string
	Value() interface{}
}

// Logger is an logger interface based on the key-value pairs to emit the log.
type Logger interface {
	Trace(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(mst string, fields ...Field) // Log and exit with the code 1.
}

// Loggerf is an logger interface based on the format string to emit the log.
type Loggerf interface {
	Tracef(msgfmt string, args ...interface{})
	Debugf(msgfmt string, args ...interface{})
	Infof(msgfmt string, args ...interface{})
	Warnf(msgfmt string, args ...interface{})
	Errorf(msgfmt string, args ...interface{})
	Fatalf(msgfmt string, args ...interface{}) // Log and exit with the code 1.
}
