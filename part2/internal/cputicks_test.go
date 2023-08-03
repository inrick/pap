package internal

import (
	"testing"
)

func BenchmarkRdtsc1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Rdtsc()
	}
}

func BenchmarkRdtsc2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Rdtsc2()
	}
}

func BenchmarkRdtscp1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Rdtscp()
	}
}

func BenchmarkRdtscp2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Rdtscp2()
	}
}
