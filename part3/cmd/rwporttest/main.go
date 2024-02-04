package main

import (
	"fmt"
	"part3/asm"
	"part3/internal"
	"part3/internal/reptest"
)

type Params struct {
	count uint64
	buf   []byte
}

var TestFunctions = []struct {
	Name string
	Func func(*reptest.Tester, Params)
}{
	{"Read_x1", mk(asm.Read_x1)},
	{"Read_x2", mk(asm.Read_x2)},
	{"Read_x3", mk(asm.Read_x3)},
	{"Read_x4", mk(asm.Read_x4)},
	{"Write_x1", mk(asm.Write_x1)},
	{"Write_x2", mk(asm.Write_x2)},
	{"Write_x3", mk(asm.Write_x3)},
	{"Write_x4", mk(asm.Write_x4)},
}

func main() {
	const nbytes = 1 << 30
	params := Params{1 << 30, make([]byte, 4096)}
	testers := make([]reptest.Tester, len(TestFunctions))
	freqReport := internal.EstimateCpuFrequency(10)
	for {
		for i, testFn := range TestFunctions {
			rt := &testers[i]
			fmt.Printf("\n--- %s ---\n", testFn.Name)
			rt.NewTestWave(nbytes, freqReport.EstFreq, 10)
			testFn.Func(rt, params)
		}
	}
}

func mk(fn func(uint64, []byte)) func(*reptest.Tester, Params) {
	return func(rt *reptest.Tester, p Params) {
		for rt.IsTesting() {
			rt.BeginTime()
			fn(p.count, p.buf)
			rt.EndTime()
			rt.CountBytes(p.count)
		}
	}
}

func ulen[E any](u []E) uint64 {
	return uint64(len(u))
}
