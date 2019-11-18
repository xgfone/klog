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

// Lazyer is used to represent a lazy value by calling a function.
type Lazyer interface {
	Lazy() interface{}
}

// LazyerFunc is used to converts the function to Lazyer.
func LazyerFunc(f func() interface{}) Lazyer { return lazyerFunc(f) }

type lazyerFunc func() interface{}

func (f lazyerFunc) Lazy() interface{} { return f() }

/////////////////////////////////////////////////////////////////////////////

// LazyerStack is used to get the stack of the caller.
type LazyerStack interface {
	Lazyer
	Stack(depth int) string
}

// LazyerStackFunc converts a function to LazyerStack.
func LazyerStackFunc(f func(depth int) string) LazyerStack {
	return stackLazyerFunc(f)
}

type stackLazyerFunc func(int) string

func (f stackLazyerFunc) Lazy() interface{}      { panic("cannot be called") }
func (f stackLazyerFunc) Stack(depth int) string { return f(depth + 1) }

// Caller returns a LazyerStack that returns the caller "file:line".
//
// If fullPath is true, the file is the full path but removing the GOPATH prefix.
func Caller(fullPath ...bool) LazyerStack {
	format := "%v"
	if len(fullPath) > 0 && fullPath[0] {
		format = "%+v"
	}

	return LazyerStackFunc(func(depth int) string {
		return fmt.Sprintf(format, stack.Caller(depth+1))
	})
}

// CallerStack returns a LazyerStack returning the caller stack without runtime.
//
// If fullPath is true, the file is the full path but removing the GOPATH prefix.
func CallerStack(fullPath ...bool) LazyerStack {
	format := "%v"
	if len(fullPath) > 0 && fullPath[0] {
		format = "%+v"
	}

	return LazyerStackFunc(func(depth int) string {
		s := stack.Trace().TrimBelow(stack.Caller(depth + 1)).TrimRuntime()
		if len(s) > 0 {
			return fmt.Sprintf(format, s)
		}
		return ""
	})
}
