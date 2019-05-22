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
	"strings"
	"testing"
)

func TestFmtLogger(t *testing.T) {
	buf := NewBuilder(128)
	log := ToFmtLogger(Std.WithWriter(StreamWriter(buf)).WithKv("caller", Caller()))
	log.Info("hello %s", "world")

	s := buf.String()
	if index := strings.IndexByte(s, ' '); index > 0 {
		s = s[index:]
	}
	if s != " lvl=INFO caller=fmt_logger_test.go:25 msg=hello world\n" {
		t.Error(s)
	}
}

func TestFmtLoggerError(t *testing.T) {
	buf := NewBuilder(128)
	log := ToFmtLoggerError(Std.WithWriter(StreamWriter(buf)).WithKv("caller", Caller()))
	log.Info("hello %s", "world")

	s := buf.String()
	if index := strings.IndexByte(s, ' '); index > 0 {
		s = s[index:]
	}
	if s != " lvl=INFO caller=fmt_logger_test.go:39 msg=hello world\n" {
		t.Error(s)
	}
}
