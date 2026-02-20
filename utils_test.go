/*
Copyright 2026 Olivier Mengué

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqlfunc_test

import "testing"

// TestingB is an extension of [testing.TB] that abstracts both [*testing.T] and [*testing.B],
// allowing to write benchmarks that can also be run as tests without modification.
type TestingB interface {
	testing.TB
	Run(name string, f func(tb TestingB)) bool

	Loop() bool    // from [testing.B]
	ReportAllocs() // from [testing.B]
	ResetTimer()   // from [testing.B]
	StartTimer()   // from [testing.B]
	StopTimer()    // from [testing.B]
}

// TestingTAsB is a helper that adapts [*testing.T] to the [TestingB] interface expected by some benchmarks.
// It allows to use the same benchmark code with both [*testing.B] and [*testing.T], enabling to run benchmarks
// in a non-looping mode for testing purposes.
func TestingTAsB(t *testing.T) TestingB {
	return &testingTAsB{T: t}
}

type testingTAsB struct {
	*testing.T
	loopDone bool
}

func (t *testingTAsB) Loop() bool {
	// Just run once
	defer func() { t.loopDone = true }()
	return !t.loopDone
}

func (t testingTAsB) ReportAllocs() {
	// No-op, as we can't report allocations from a *testing.T
}

func (t testingTAsB) ResetTimer() {
	// No-op, as there is no timer in a *testing.T
}

func (t testingTAsB) StartTimer() {
	// No-op, as there is no timer in a *testing.T
}

func (t testingTAsB) StopTimer() {
	// No-op, as there is no timer in a *testing.T
}

func (t testingTAsB) Run(name string, f func(tb TestingB)) bool {
	return t.T.Run(name, func(t *testing.T) {
		f(TestingTAsB(t))
	})
}

// TestingBAsB is a helper that adapts [*testing.B] to the [TestingB] interface expected by some benchmarks.
// It allows to use the same benchmark code with both [*testing.B] and [*testing.T], enabling to run benchmarks
// in a non-looping mode for testing purposes.
func TestingBAsB(b *testing.B) TestingB {
	return testingBAsB{B: b}
}

type testingBAsB struct {
	*testing.B
}

func (b testingBAsB) Run(name string, f func(tb TestingB)) bool {
	return b.B.Run(name, func(b *testing.B) {
		f(TestingBAsB(b))
	})
}

func benchTestingB(tb TestingB) {
	tb.Log("Running via TestingB interface")
	tb.Run("sub", func(tb TestingB) {
		tb.Log("Running subtest via TestingB interface")
		n := 0
		tb.ReportAllocs()
		tb.ResetTimer()
		for tb.Loop() {
			n++
		}
		if _, isT := tb.(interface{ Parallel() }); isT && n != 1 {
			tb.Errorf("Loop should run only once when running with *testing.T, but ran %d times", n)
		} else {
			tb.Logf("Loop ran %d times", n)
		}
	})
}

func TestTestingB(t *testing.T) {
	benchTestingB(TestingTAsB(t))
}

func BenchmarkTestingB(b *testing.B) {
	benchTestingB(TestingBAsB(b))
}
