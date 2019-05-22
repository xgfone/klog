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
	logger := New(DiscardWriter()).WithEncoder(NothingEncoder())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().K("key", "vlaue").Msg("message")
		}
	})
}

func BenchmarkKlogTextEncoder(b *testing.B) {
	logger := New(DiscardWriter())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().K("key", "vlaue").Msg("message")
		}
	})
}

func BenchmarkKlogJSONEncoder(b *testing.B) {
	logger := New(DiscardWriter()).WithEncoder(JSONEncoder())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().K("key", "vlaue").Msg("message")
		}
	})
}

func BenchmarkKlogStdJSONEncoder(b *testing.B) {
	logger := New(DiscardWriter()).WithEncoder(StdJSONEncoder())

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info().K("key", "vlaue").Msg("message")
		}
	})
}
