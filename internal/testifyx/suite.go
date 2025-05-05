package testifyx

import (
	"fmt"
	"testing"
)

type TestSuite struct {
	t          *testing.T
	name       string
	beforeEach func()
	afterEach  func()
}

type TestSuiteBench struct {
	b          *testing.B
	name       string
	beforeEach func()
	afterEach  func()
}

func Describe(t *testing.T, name string, fn func(*TestSuite)) {
	t.Helper()
	t.Logf("\n\n=== SUITE: %s ===", name)
	ts := &TestSuite{t: t, name: name}
	fn(ts)
}

func (ts *TestSuite) It(name string, fn func(*TC)) {
	ts.t.Run(fmt.Sprintf("%s: %s", ts.name, name), func(t *testing.T) {
		tc := &TC{t: t}
		if ts.beforeEach != nil {
			ts.beforeEach()
		}
		fn(tc)
		if ts.afterEach != nil {
			ts.afterEach()
		}
	})
}

func Benchmark(b *testing.B, name string, fn func(*TestSuiteBench)) {
	b.Helper()
	b.Logf("\n\n=== BENCHMARK SUITE: %s ===", name)
	ts := &TestSuiteBench{b: b, name: name}
	fn(ts)
}

func (ts *TestSuiteBench) Bench(name string, fn func(*TC)) {
	ts.b.Run(fmt.Sprintf("%s: %s", ts.name, name), func(b *testing.B) {
		tc := &TC{t: b}
		for i := 0; i < b.N; i++ {
			if ts.beforeEach != nil {
				ts.beforeEach()
			}
			fn(tc)
			if ts.afterEach != nil {
				ts.afterEach()
			}
		}
	})
}
