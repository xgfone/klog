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

import (
	"encoding/json"
	"testing"
)

func TestLoggerTextEncoder(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithCtx(Caller("caller1"))
	logger.Encoder = TextEncoder(StreamWriter(buf), Quote(), EncodeLevel("lvl"))

	logger.Info("test logger", Caller("caller2"))
	if buf.String() != "lvl=INFO caller1=logger_test.go:27 caller2=logger_test.go:27 msg=\"test logger\"\n" {
		t.Error(buf.String())
	}
}

func TestLoggerJSONEncoder(t *testing.T) {
	buf := NewBuilder(128)
	logger := New("").WithCtx(Caller("caller1"), F("key1", `value1"`))
	logger.Encoder = JSONEncoder(StreamWriter(buf), EncodeLevel("lvl"))
	logger.Info("test json encoder", F("key2", 123))

	expect := `{"lvl":"INFO","caller1":"logger_test.go:37","key1":"value1\"","key2":123,"msg":"test json encoder"}` + "\n"
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

type serror string

func (e serror) Error() string { return string(e) }

func TestFieldError(t *testing.T) {
	err := serror("error")
	buf := NewBuilder(128)
	logger := New("")
	logger.Encoder = JSONEncoder(StreamWriter(buf))
	logger.Info("test FieldError with json encoder", E(FE(err)))
	logger.Info("test FieldError with json encoder", E(FE(err, F("k1", "v1"))))
	logger.Info("test FieldError with json encoder", E(FE(err, F("k2", "v2"), F("k3", "v3"))))

	logger.Encoder = TextEncoder(StreamWriter(buf))
	logger.Info("test FieldError with json encoder", E(FE(err)))
	logger.Info("test FieldError with json encoder", E(FE(err, F("k4", "v4"))))
	logger.Info("test FieldError with text encoder", E(FE(err, F("k5", "v5"), F("k6", "v6"))))

	expect := `{"err":"error","msg":"test FieldError with json encoder"}
{"err":"error","k1":"v1","msg":"test FieldError with json encoder"}
{"err":"error","k2":"v2","k3":"v3","msg":"test FieldError with json encoder"}
err=error msg=test FieldError with json encoder
err=error k4=v4 msg=test FieldError with json encoder
err=error k5=v5 k6=v6 msg=test FieldError with text encoder
`
	if s := buf.String(); s != expect {
		t.Error(s)
	}
}

func TestFieldBuilder_Reset(t *testing.T) {
	fb := FB(4).F("k1", "v1").F("k2", "v2")
	if fields := fb.Fields(); len(fields) != 2 {
		t.Error(fields)
	}

	fb.Reset()
	if fields := fb.Fields(); len(fields) != 0 {
		t.Error(fields)
	}
}
