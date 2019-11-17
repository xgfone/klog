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
	"fmt"

	"github.com/go-stack/stack"
)

// Valuer is used to represent a lazy value by calling a function.
type Valuer func(Record) (v interface{})

// Caller returns a Valuer that returns the caller "file:line".
//
// If fullPath is true, the file is the full path but removing the GOPATH prefix.
func Caller(fullPath ...bool) Valuer {
	format := "%v"
	if len(fullPath) > 0 && fullPath[0] {
		format = "%+v"
	}

	return func(r Record) interface{} {
		return fmt.Sprintf(format, stack.Caller(r.Depth+1))
	}
}

// CallerStack returns a Valuer returning the caller stack without runtime.
//
// If fullPath is true, the file is the full path but removing the GOPATH prefix.
func CallerStack(fullPath ...bool) Valuer {
	format := "%v"
	if len(fullPath) > 0 && fullPath[0] {
		format = "%+v"
	}

	return func(r Record) interface{} {
		s := stack.Trace().TrimBelow(stack.Caller(r.Depth + 1)).TrimRuntime()
		if len(s) > 0 {
			return fmt.Sprintf(format, s)
		}
		return ""
	}
}
