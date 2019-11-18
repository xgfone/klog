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
	"fmt"
	"strings"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := WithEncoder(TextEncoder(StreamWriter(buf), EncodeLevel("lvl"))).WithCtx(F("caller", Caller()))
	SetDefaultLogger(logger)

	Info("msg1", F("key", "value"))
	Infof("%s", "msg2")
	Ef(fmt.Errorf("error"), "msg3")

	expectedLines := []string{
		"lvl=INFO caller=global_test.go:29 key=value msg=msg1",
		"lvl=INFO caller=global_test.go:30 msg=msg2",
		"lvl=ERROR caller=global_test.go:31 err=error msg=msg3",
		"",
	}

	lines := strings.Split(buf.String(), "\n")
	if len(lines) != len(expectedLines) {
		t.Errorf("expected %d lines, but got %d", len(expectedLines), len(lines))
	} else {
		for i := range lines {
			if lines[i] != expectedLines[i] {
				t.Errorf("expected '%s', got '%s'", expectedLines[i], lines[i])
			}
		}
	}
}
