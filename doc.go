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
//   - Built-in support for logging to files, syslog, and the network.
//   - Child loggers which inherit and add their own private context.
//   - Support lots of `Writer`, such as `StreamWriter`, `NetWriter`, `SyslogWriter`, etc.
//
package klog
