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
	"strconv"
	"strings"

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

// LineNo returns the line number of the caller as string.
func LineNo() Valuer {
	return func(r Record) interface{} {
		caller := fmt.Sprintf("%v", stack.Caller(r.Depth+1))
		if index := strings.IndexByte(caller, ':'); index > -1 {
			return caller[index+1:]
		}
		return ""
	}
}

// LineNoAsInt returns the line number of the caller.
//
// Return 0 if the line is missing.
func LineNoAsInt() Valuer {
	return func(r Record) interface{} {
		caller := fmt.Sprintf("%v", stack.Caller(r.Depth+1))
		if index := strings.IndexByte(caller, ':'); index > -1 {
			lineno, _ := strconv.ParseInt(caller[index+1:], 10, 32)
			return int(lineno)
		}
		return 0
	}
}

// FuncName returns the name of the function where the caller is in.
func FuncName() Valuer {
	return func(r Record) interface{} {
		return fmt.Sprintf("%n", stack.Caller(r.Depth+1))
	}
}

// FuncFullName returns the full name of the function where the caller is in.
func FuncFullName() Valuer {
	return func(r Record) interface{} {
		return fmt.Sprintf("%+n", stack.Caller(r.Depth+1))
	}
}

// FileName returns the short name of the file where the caller is in.
func FileName() Valuer {
	return func(r Record) interface{} {
		return fmt.Sprintf("%s", stack.Caller(r.Depth+1))
	}
}

// FileLongName returns the long name of the file where the caller is in.
func FileLongName() Valuer {
	return func(r Record) interface{} {
		return fmt.Sprintf("%+s", stack.Caller(r.Depth+1))
	}
}

// Package returns the name of the package where the caller is in.
func Package() Valuer {
	return func(r Record) interface{} {
		path := fmt.Sprintf("%+n", stack.Caller(r.Depth+1))
		if index := strings.LastIndexByte(path, '.'); index > -1 {
			return path[:index]
		}
		return path
	}
}
