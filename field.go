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

import "sync"

type field struct {
	key   string
	value interface{}
}

func (f field) Key() string { return f.key }
func (f field) Value() interface{} {
	switch v := f.value.(type) {
	case func() interface{}:
		return v()
	case func() string:
		return v()
	default:
		return v
	}
}

// E is equal to F("err", err).
func E(err error) Field { return field{key: "err", value: err} }

// F returns a new Field. If value is "func() interface{}" or "func() string",
// it will be evaluated when the log is emitted, that's, it is lazy.
func F(key string, value interface{}) Field { return field{key: key, value: value} }

// FieldFunc returns a higher-order function to create the field with the key.
func FieldFunc(key string) func(value interface{}) Field {
	return func(value interface{}) Field {
		return F(key, value)
	}
}

// FieldReleaser is used to get and release the fields.
type FieldReleaser interface {
	Fields() []Field
	Release()
}

// FieldError is the error interface with some fields.
type FieldError interface {
	FieldReleaser
	error
}

// FE is the alias of NewFieldError.
func FE(err error, fields FieldReleaser) FieldError {
	return NewFieldError(err, fields)
}

// NewFieldError returns a new FieldError.
func NewFieldError(err error, fields FieldReleaser) FieldError {
	return fieldError{error: err, FieldReleaser: fields}
}

type fieldError struct {
	FieldReleaser
	error
}

func (e fieldError) Unwrap() error { return e.error }

var fbPool4 = sync.Pool{New: func() interface{} { return make([]Field, 0, 4) }}
var fbPool8 = sync.Pool{New: func() interface{} { return make([]Field, 0, 8) }}
var fbPool16 = sync.Pool{New: func() interface{} { return make([]Field, 0, 16) }}
var fbPool32 = sync.Pool{New: func() interface{} { return make([]Field, 0, 32) }}

// FieldBuilder is used to build a set of fields.
type FieldBuilder struct {
	fields []Field
}

// NewFieldBuilder returns a new FieldBuilder with the capacity of the fields.
func NewFieldBuilder(cap int) FieldBuilder {
	var fields []Field
	if cap > 16 {
		fields = fbPool32.Get().([]Field)
	} else if cap > 8 {
		fields = fbPool16.Get().([]Field)
	} else if cap > 4 {
		fields = fbPool8.Get().([]Field)
	} else {
		fields = fbPool4.Get().([]Field)
	}
	return FieldBuilder{fields: fields}
}

// FB is the alias of NewFieldBuilder.
func FB(cap int) FieldBuilder { return NewFieldBuilder(cap) }

// Field appends the field with the key and value.
func (fb FieldBuilder) Field(key string, value interface{}) FieldBuilder {
	fb.fields = append(fb.fields, F(key, value))
	return fb
}

// F is the alias of Field.
func (fb FieldBuilder) F(key string, value interface{}) FieldBuilder {
	return fb.Field(key, value)
}

// E appends the error field if err is not equal to nil.
func (fb FieldBuilder) E(err error) FieldBuilder {
	if err != nil {
		fb.fields = append(fb.fields, E(err))
	}
	return fb
}

// Fields returns the built fields.
func (fb FieldBuilder) Fields() []Field { return fb.fields }

// Reset resets the fields and reuses the underlying memory.
func (fb *FieldBuilder) Reset() { fb.fields = fb.fields[:0] }

// Release releases the fields into the pool.
func (fb FieldBuilder) Release() {
	fb.Reset()
	if cap := len(fb.fields); cap > 16 {
		fbPool32.Put(fb.fields)
	} else if cap > 8 {
		fbPool16.Put(fb.fields)
	} else if cap > 4 {
		fbPool8.Put(fb.fields)
	} else {
		fbPool4.Put(fb.fields)
	}
}
