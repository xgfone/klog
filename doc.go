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
// package main
//
//     import (
//         "fmt"
//
//         "github.com/xgfone/klog/v2"
//     )
//
//     func main() {
//         // Initialize the default logger.
//         log := klog.WithLevel(klog.LvlWarn).WithCtx(klog.F("caller", klog.Caller()))
//         // if file, err := klog.FileWriter("file.log", "100M", 100); err == nil {
//         //     log.Encoder().SetWriter(file)
//         // } else {
//         //     fmt.Println(err)
//         //     return
//         // }
//         klog.SetDefaultLogger(log)
//
//         // Emit the log with the fields.
//         klog.Info("msg", klog.F("k1", "v1"), klog.F("k2", "v2"))
//         klog.Error("msg", klog.F("k1", "v1"), klog.F("k2", "v2"))
//
//         // Emit the log with the formatter.
//         klog.Infof("log %s", "msg")
//         klog.Warnf("log %s", "msg")
//         klog.Ef(fmt.Errorf("e"), "log %s", "msg")
//
//         // Output:
//         // t=2019-11-19T09:54:36.4956708+08:00 lvl=ERROR caller=main.go:22 k1=v1 k2=v2 msg=msg
//         // t=2019-11-19T09:54:36.4973725+08:00 lvl=WARN caller=main.go:26 msg="log msg"
//         // t=2019-11-19T09:54:36.4974311+08:00 lvl=ERROR caller=main.go:27 err=e msg="log msg"
//     }
//
package klog
