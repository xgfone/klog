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

import "testing"

func BenchmarkKlogNothingEncoder(b *testing.B) {
	logger := New("").WithEncoder(NothingEncoder())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Log(LvlInfo, "message")
		}
	})
}

func BenchmarkKlogTextEncoder(b *testing.B) {
	logger := New("").WithEncoder(TextEncoder(DiscardWriter()))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Log(LvlInfo, "message")
		}
	})
}

func BenchmarkKlogJSONEncoder(b *testing.B) {
	logger := New("").WithEncoder(JSONEncoder(DiscardWriter()))

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Log(LvlInfo, "message")
		}
	})
}

func BenchmarkKlogTextEncoder10CtxFields(b *testing.B) {
	fs := []Field{F("k1", "v1"), F("k2", "v2"), F("k3", "v3"), F("k4", "v4"), F("k5", "v5"),
		F("k6", "v6"), F("k7", "v7"), F("k8", "v8"), F("k9", "v9"), F("k10", "v10")}
	logger := New("").WithEncoder(TextEncoder(DiscardWriter())).WithCtx(fs...)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Log(LvlInfo, "message")
		}
	})
}

func BenchmarkKlogJSONEncoder10CtxFields(b *testing.B) {
	fs := []Field{F("k1", "v1"), F("k2", "v2"), F("k3", "v3"), F("k4", "v4"), F("k5", "v5"),
		F("k6", "v6"), F("k7", "v7"), F("k8", "v8"), F("k9", "v9"), F("k10", "v10")}
	logger := New("").WithEncoder(JSONEncoder(DiscardWriter())).WithCtx(fs...)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Log(LvlInfo, "message")
		}
	})
}
