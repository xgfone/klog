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

// +build !windows,!plan9

package klog

import (
	"log/syslog"
	"testing"
)

func TestSyslogWriter(t *testing.T) {
	w, err := SyslogWriter(syslog.LOG_DEBUG, "testsyslog")
	if err != nil {
		t.Error(err)
	} else {
		defer w.Close()
		logger := New("")
		logger.Encoder = TextEncoder(w)
		logger.Info("test", F("writer", "syslog"))
	}
}
