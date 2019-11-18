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
	"bytes"
	"testing"
)

func TestLevelLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log := New("")
	log.SetEncoder(TextEncoder(StreamWriter(buf), EncodeLevel("lvl")))
	logger := ToLevelLogger(log.WithCtx(F("caller1", Caller())))
	logger.Info("msg", F("caller2", Caller()), F("key", "value"))

	expect := "lvl=INFO caller1=fmt_logger_test.go:27 caller2=fmt_logger_test.go:27 key=value msg=msg\n"
	if buf.String() != expect {
		t.Error(buf.String())
	}
}

func TestPrintfer(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log := New("")
	log.SetEncoder(TextEncoder(StreamWriter(buf), EncodeLevel("lvl")))
	logger := ToPrintfer(log.WithCtx(F("caller", Caller())))
	logger.Printf("test %s", "Printfer")

	expect := "lvl=INFO caller=fmt_logger_test.go:40 msg=test Printfer\n"
	if buf.String() != expect {
		t.Error(buf.String())
	}
}

func TestFmtLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	log := New("")
	log.SetEncoder(TextEncoder(StreamWriter(buf), EncodeLevel("lvl")))
	logger := ToFmtLogger(log.WithCtx(F("caller", Caller())))
	logger.Infof("test %s", "FmtLogger")

	expect := "lvl=INFO caller=fmt_logger_test.go:53 msg=test FmtLogger\n"
	if buf.String() != expect {
		t.Error(buf.String())
	}
}
