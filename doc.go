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

// Package klog provides an simple, flexible, extensible, powerful and
// structured logging tool based on the level, which has done the better balance
// between the flexibility and the performance.
//
// Features
//
//   - The better performance.
//   - Lazy evaluation of expensive operations.
//   - Simple, Flexible, Extensible, Powerful and Structured.
//   - Avoid to allocate the memory on heap as far as possible.
//   - Child loggers which inherit and add their own private context.
//   - Built-in support for logging to files, syslog, and the network. See `Writer`.
//
// Example
//
//      package main
//
//      import (
//      	"fmt"
//
//      	"github.com/xgfone/klog/v2"
//      )
//
//      func main() {
//      	// Initialize the default logger.
//      	log := klog.WithLevel(klog.LvlWarn).WithCtx(klog.F("caller", klog.Caller()))
//      	klog.SetDefaultLogger(log)
//
//      	// Emit the log with the fields.
//      	klog.Info("info log msg", klog.F("key1", "value1"), klog.F("key2", "value2"))
//      	klog.Error("error log msg", klog.F("key1", "value1"), klog.F("key2", "value2"))
//
//      	// Emit the log with the formatter.
//      	klog.Infof("%s log msg", "infof")
//      	klog.Errorf("%s log msg", "errorf")
//      	klog.Ef(fmt.Errorf("error"), "%s log msg", "errorf")
//
//      	// Output:
//      	// t=2019-11-18T14:01:08.7345586+08:00 lvl=ERROR caller=main.go:15 key1=value1 key2=value2 msg="error log msg"
//      	// t=2019-11-18T14:01:08.735969+08:00 lvl=ERROR caller=main.go:18 msg="errorf log msg"
//      	// t=2019-11-18T14:01:08.7360115+08:00 lvl=ERROR caller=main.go:19 err=error msg="errorf log msg"
//      }
package klog
