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
	"encoding/json"
	"strings"
	"testing"
)

func TestLoggerTextEncoder(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithCtx(F("caller1", Caller()))
	logger.SetEncoder(TextEncoder(StreamWriter(buf), Quote(), EncodeLevel("lvl")))

	logger.Log(LvlInfo, "test logger", F("caller2", Caller()))
	if buf.String() != "lvl=INFO caller1=logger_test.go:28 caller2=logger_test.go:28 msg=\"test logger\"\n" {
		t.Error(buf.String())
	}
}

func TestLoggerJSONEncoder(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithCtx(F("caller1", Caller()), F("key1", `value1"`))
	logger.SetEncoder(JSONEncoder(StreamWriter(buf), EncodeLevel("lvl")))
	logger.Log(LvlInfo, "test json encoder", F("key2", 123))

	expect := `{"lvl":"INFO","caller1":"logger_test.go:38","key1":"value1\"","key2":123,"msg":"test json encoder"}` + "\n"
	if buf.String() != expect {
		t.Error(buf.String())
	}

	var ms map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if v, ok := ms["key1"].(string); !ok || v != `value1"` {
		t.Error(v)
	}
}

func TestLoggerIsEnabled(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithLevel(LvlInfo)
	logger.SetEncoder(TextEncoder(StreamWriter(buf), EncodeLevel("lvl")))

	if logger.(interface{ IsEnabled(Level) bool }).IsEnabled(LvlDebug) {
		logger.Log(LvlDebug, "debug")
	}
	if logger.(interface{ IsEnabled(Level) bool }).IsEnabled(LvlInfo) {
		logger.Log(LvlInfo, "info")
	}

	if output := buf.String(); strings.Contains(output, "debug") {
		t.Error(output)
	} else if !strings.Contains(output, "info") {
		t.Error(output)
	}
}
