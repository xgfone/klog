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
	"bytes"
	"io"
	"log/syslog"
)

type syslogWriter struct {
	w *syslog.Writer
}

func (s syslogWriter) Write(level Level, p []byte) (n int, err error) {
	v := string(bytes.TrimSpace(p))
	if level >= LvlFatal {
		err = s.w.Emerg(v)
	} else if level >= LvlPanic {
		err = s.w.Crit(v)
	} else if level >= LvlError {
		err = s.w.Err(v)
	} else if level >= LvlWarn {
		err = s.w.Warning(v)
	} else if level >= LvlInfo {
		err = s.w.Info(v)
	} else {
		err = s.w.Debug(v)
	}

	if err == nil {
		n = len(p)
	}
	return
}

// SyslogWriter opens a connection to the system syslog daemon
// by calling syslog.New and writes all logs to it.
func SyslogWriter(priority syslog.Priority, tag string) (Writer, io.Closer, error) {
	w, err := syslog.New(priority, tag)
	if err != nil {
		return nil, nil, err
	}
	return syslogWriter{w}, w, nil
}

// SyslogNetWriter opens a connection to a log daemon over the network
// and writes all logs to it.
func SyslogNetWriter(net, addr string, priority syslog.Priority, tag string) (Writer, io.Closer, error) {
	w, err := syslog.Dial(net, addr, priority, tag)
	if err != nil {
		return nil, nil, err
	}
	return syslogWriter{w}, w, nil
}
