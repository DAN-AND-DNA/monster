package base

import "testing"

type Base interface {
	F()
}

type A struct {
}

func (this *A) F() {
	i := 7
	i++
}

func ff(b Base) {
	b.F()
}

func Benchmark_hh(b *testing.B) {
	b.ReportAllocs()

	a := &A{}
	for i := 0; i < b.N; i++ {
		ff(a)
	}
}
