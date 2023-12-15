package main

import (
	"fmt"
	"part3/asm"
	"part3/internal"
	"part3/internal/reptest"
)

type Params struct {
	bb []byte
}

var TestFunctions = []struct {
	Name string
	Func func(*reptest.Tester, Params)
}{
	//{"MovAllBytes", mk(asm.MovAllBytes)},
	//{"CmpAllBytes", mk(asm.CmpAllBytes)},
	//{"DecAllBytes", mk(asm.DecAllBytes)},
	{"Nop3x1AllBytes", mk(asm.Nop3x1AllBytes)},
	{"Nop1x3AllBytes", mk(asm.Nop1x3AllBytes)},
	{"Nop1x9AllBytes", mk(asm.Nop1x9AllBytes)},
}

func main() {
	const nbytes = 1 << 30
	params := Params{make([]byte, nbytes)}
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

func mk(fn func([]byte)) func(*reptest.Tester, Params) {
	return func(rt *reptest.Tester, p Params) {
		for rt.IsTesting() {
			rt.BeginTime()
			fn(p.bb)
			rt.EndTime()
			rt.CountBytes(ulen(p.bb))
		}
	}
}

func ulen[E any](u []E) uint64 {
	return uint64(len(u))
}
