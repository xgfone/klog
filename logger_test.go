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

func TestLogger(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithCtx((F("caller1", Caller())))
	logger.SetEncoder(TextEncoder(StreamWriter(buf), Quote()))
	logger.Log(LvlInfo, "test logger", F("caller2", Caller()))

	s := buf.String()
	if s != "lvl=INFO caller1=logger_test.go:27 caller2=logger_test.go:27 msg=\"test logger\"\n" {
		t.Error(s)
	}

	buf.Reset()
	logger.Log(LvlInfo, "test logger", F("caller2", Caller()))

	s = buf.String()
	if s != "lvl=INFO caller1=logger_test.go:35 caller2=logger_test.go:35 msg=\"test logger\"\n" {
		t.Error(s)
	}

	buf.Reset()
	logger.Log(LvlInfo, "test 123")

	s = buf.String()
	if s != "lvl=INFO caller1=logger_test.go:43 msg=\"test 123\"\n" {
		t.Error(s)
	}
}

func TestJSONEncoder(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithCtx(F("key1", `value1"`))
	logger.SetEncoder(JSONEncoder(StreamWriter(buf)))
	logger.Log(LvlInfo, "testlogger", F("key2", 123))

	var ms map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 4 {
		t.Error(ms)
	} else if ms["t"] != nil {
		t.Error(ms)
	} else if v, ok := ms["lvl"].(string); !ok || v != "INFO" {
		t.Error(ms)
	} else if v, ok := ms["key1"].(string); !ok || v != `value1"` {
		t.Error(ms)
	} else if v, ok := ms["key2"].(float64); !ok || v != 123 {
		t.Error(ms)
	} else if v, ok := ms["msg"].(string); !ok || v != "testlogger" {
		t.Error(ms)
	}

	buf.Reset()
	ms = make(map[string]interface{}, 10)
	logger.Log(LvlInfo, "test logger", F("key2", 123))

	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 4 {
		t.Error(ms)
	} else if ms["t"] != nil {
		t.Error(ms)
	} else if v, ok := ms["lvl"].(string); !ok || v != "INFO" {
		t.Error(ms)
	} else if v, ok := ms["key1"].(string); !ok || v != `value1"` {
		t.Error(ms)
	} else if v, ok := ms["key2"].(float64); !ok || v != 123 {
		t.Error(ms)
	} else if v, ok := ms["msg"].(string); !ok || v != "test logger" {
		t.Error(ms)
	}
}

func TestStdJSONEncoder(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithCtx(F("key1", `value1"`))
	logger.SetEncoder(StdJSONEncoder(StreamWriter(buf)))
	logger.Log(LvlInfo, "testlogger", F("key2", 123))

	var ms map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 4 {
		t.Error(ms)
	} else if ms["t"] != nil {
		t.Error(ms)
	} else if v, ok := ms["lvl"].(string); !ok || v != "INFO" {
		t.Error(ms)
	} else if v, ok := ms["key1"].(string); !ok || v != `value1"` {
		t.Error(ms)
	} else if v, ok := ms["key2"].(float64); !ok || v != 123 {
		t.Error(ms)
	} else if v, ok := ms["msg"].(string); !ok || v != "testlogger" {
		t.Error(ms)
	}

	buf.Reset()
	ms = make(map[string]interface{}, 10)
	logger.Log(LvlInfo, "test logger", F("key2", 123))

	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 4 {
		t.Error(ms)
	} else if ms["t"] != nil {
		t.Error(ms)
	} else if v, ok := ms["lvl"].(string); !ok || v != "INFO" {
		t.Error(ms)
	} else if v, ok := ms["key1"].(string); !ok || v != `value1"` {
		t.Error(ms)
	} else if v, ok := ms["key2"].(float64); !ok || v != 123 {
		t.Error(ms)
	} else if v, ok := ms["msg"].(string); !ok || v != "test logger" {
		t.Error(ms)
	}
}

type keyValueTest struct {
	key   string
	value interface{}
}

func (kv keyValueTest) Key() string {
	return kv.key
}

func (kv keyValueTest) Value() interface{} {
	return kv.value
}

func TestLogger_IsEnabled(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithLevel(LvlInfo)
	logger.Encoder().SetWriter(StreamWriter(buf))

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
