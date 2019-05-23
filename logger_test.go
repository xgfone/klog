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
	logger := New(StreamWriter(buf)).WithKv("caller1", Caller())
	logger.Info().K("caller2", Caller()).Msgf("test %s", "logger")

	s := buf.String()
	s = s[strings.IndexByte(s, ' '):]
	if s != " lvl=INFO caller1=logger_test.go:26 caller2=logger_test.go:26 msg=test logger\n" {
		t.Error(s)
	}

	buf.Reset()
	logger.K("caller2", Caller()).Info("test %s", "logger")

	s = buf.String()
	s = s[strings.IndexByte(s, ' '):]
	if s != " lvl=INFO caller1=logger_test.go:35 caller2=logger_test.go:35 msg=test logger\n" {
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
	logger.K("key2", 123).Info("test %s", "logger")

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
	logger.K("key2", 123).Info("test %s", "logger")

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
