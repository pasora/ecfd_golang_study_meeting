package main

import "testing"

func Benchmark_copy(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy()
	}
}

func Benchmark_tee(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tee()
	}
}
