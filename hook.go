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

// Hook is a log hook, which receives the logger name and the log level and
// reports whether the log should continue to be emitted.
//
// Notice: You can use it to sample the log.
type Hook func(name string, level Level) bool

// DisableLogger disables the logger, whose name is in names, to emit the log.
func DisableLogger(names ...string) Hook {
	return func(name string, level Level) bool {
		for _, _name := range names {
			if name == _name {
				return false
			}
		}
		return true
	}
}

// EnableLogger allows the logger, whose name is in names, to emit the log.
func EnableLogger(names ...string) Hook {
	return func(name string, level Level) bool {
		for _, _name := range names {
			if name == _name {
				return true
			}
		}
		return false
	}
}
