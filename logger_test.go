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
	"fmt"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	buf := NewBuilder(128)
	logger := New(StreamWriter(buf)).WithKv("caller1", Caller())
	logger.Info().K("caller2", Caller()).Msgf("test %s", "logger")

	s := buf.String()
	s = s[strings.IndexByte(s, ' '):]
	if s != " lvl=INFO caller1=logger_test.go:27 caller2=logger_test.go:27 msg=\"test logger\"\n" {
		t.Error(s)
	}

	buf.Reset()
	logger.K("caller2", Caller()).Infof("test %s", "logger")

	s = buf.String()
	s = s[strings.IndexByte(s, ' '):]
	if s != " lvl=INFO caller1=logger_test.go:36 caller2=logger_test.go:36 msg=\"test logger\"\n" {
		t.Error(s)
	}

	buf.Reset()
	logger.Infof("test %d", 123)

	s = buf.String()
	s = s[strings.IndexByte(s, ' '):]
	if s != " lvl=INFO caller1=logger_test.go:45 msg=\"test 123\"\n" {
		t.Error(s)
	}
}

func TestJSONEncoder(t *testing.T) {
	buf := NewBuilder(128)
	logger := New(StreamWriter(buf)).WithEncoder(JSONEncoder())
	logger = logger.WithKv("key1", `value1"`)
	logger.Info().K("key2", 123).Msg("test", "logger")

	var ms map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 5 {
		t.Error(ms)
	} else if ms["t"] == nil {
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
	logger.K("key2", 123).Infof("test %s", "logger")

	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 5 {
		t.Error(ms)
	} else if ms["t"] == nil {
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
	logger := New(StreamWriter(buf)).WithEncoder(StdJSONEncoder())
	logger = logger.WithKv("key1", `value1"`)
	logger.Info().K("key2", 123).Msg("test", "logger")

	var ms map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 5 {
		t.Error(ms)
	} else if ms["t"] == nil {
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
	logger.K("key2", 123).Infof("test %s", "logger")

	if err := json.Unmarshal(buf.Bytes(), &ms); err != nil {
		t.Errorf("%s: %v", buf.String(), err)
	} else if len(ms) != 5 {
		t.Error(ms)
	} else if ms["t"] == nil {
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

func TestKV(t *testing.T) {
	buf := NewBuilder(128)
	logger := New(StreamWriter(buf))
	logger.V(keyValueTest{key: "key", value: 123}).Infof("test %s", "kv")

	buf.TrimNewline()
	s := buf.String()
	s = s[strings.IndexByte(s, ' '):]
	if s != ` lvl=INFO key=123 msg="test kv"` {
		t.Error(s)
	}
}

func TestLoggerSetter(t *testing.T) {
	log := New()
	log.AddDepthSelf(100)

	if newDepth := log.GetDepth(); newDepth != 100 {
		t.Errorf("the depth '%d' is not updated", newDepth)
	}
}

func TestGetLogger(t *testing.T) {
	stdlog := &Std
	if GetLogger("") != stdlog {
		t.Error(`GetLogger("") is not Std`)
	}

	log1 := GetLogger("log1").AddDepthSelf(100).AddDepthSelf(100)
	log2 := GetLogger("log1")
	if log1 != log2 || log1.GetName() != log2.GetName() || log2.GetDepth() != 200 {
		t.Error(`GetLogger("log1") != GetLogger("log1")`)
	}

	log1.SetDepth(123)
	if log2.GetDepth() != 123 {
		t.Error("Logger.SetDepth() is not reset")
	}
}

func TestLogger_SetLevelString(t *testing.T) {
	log := New()
	log.SetLevelString("error")
	if lvl := log.GetLevel(); lvl != LvlError {
		t.Error(lvl)
	}
}

func TestLoggerEf(t *testing.T) {
	buf := NewBuilder(128)
	logger := New(StreamWriter(buf)).WithKv("caller", Caller())
	logger.Ef(fmt.Errorf("test_error"), "test %s", "error")

	buf.TrimNewline()
	s := buf.String()
	s = s[strings.IndexByte(s, ' '):]
	if s != ` lvl=ERROR caller=logger_test.go:206 err=test_error msg="test error"` {
		t.Error(s)
	}
}
