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
	"strings"
)

// Hook is a log hook, which receives the logger name and the log level and
// reports whether the log should continue to be emitted.
//
// Notice: You can use it to sample the log.
type Hook func(name string, level Level) bool

func getNamesFromEnv(env string, enable bool) (names []string) {
	for _, kv := range os.Environ() {
		if index := strings.IndexByte(kv, '='); index > 0 && kv[:index] == env {
			for _, value := range strings.Split(kv[index+1:], ",") {
				if value = strings.TrimSpace(value); value != "" {
					tmp := strings.Split(value, "=")
					if len(tmp) == 2 {
						v := strings.ToLower(strings.TrimSpace(tmp[1]))
						if enable {
							if v == "on" || v == "1" || v == "true" {
								names = append(names, strings.TrimSpace(tmp[0]))
							}
						} else {
							if v == "off" || v == "0" || v == "false" {
								names = append(names, strings.TrimSpace(tmp[0]))
							}
						}
					}
				}
			}
			break
		}
	}
	return
}

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

// DisableLoggerFromEnv is the same as DisableLogger, but get the names
// from the environ.
//
// The environ format is "environ=value", and `value` is like "name=off|0[,name=off|0]*".
// So, for `DisableLoggerFromEnv("mod")`, the environ variable
//
//    mod=n1=0,n2=off,n3=on,n4=1
//
// the loggers named "n1" and "n2" will be disabled, and others, including "n3"
// and "n4", will be enabled.
func DisableLoggerFromEnv(environ string) Hook {
	names := getNamesFromEnv(environ, false)
	return func(name string, level Level) bool {
		for _, _name := range names {
			if name == _name {
				return false
			}
		}
		return true
	}
}

// EnableLoggerFromEnv is the same as EnableLogger, but get the names
// from the environ.
//
// The environ format is "environ=value", and `value` is like "name=on|1[,name=on|1]*".
// So, for `EnableLoggerFromEnv("mod")`, the environ variable
//
//    mod=n1=0,n2=off,n3=on,n4=1
//
// the loggers named "n3" and "n4" will be enabled, and others, including "n1"
// and "n2", will be disabled.
func EnableLoggerFromEnv(environ string) Hook {
	names := getNamesFromEnv(environ, true)
	return func(name string, level Level) bool {
		for _, _name := range names {
			if name == _name {
				return true
			}
		}
		return false
	}
}
