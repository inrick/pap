package main

import (
	"fmt"
	"part3/asm"
	"part3/goasm"
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
	{"ReadSuccessiveSizes_go_4KB", mk(goasm.ReadSuccessiveSizes_go, 1<<12-1)},
	{"ReadSuccessiveSizes_go_32KB", mk(goasm.ReadSuccessiveSizes_go, 1<<15-1)},
	{"ReadSuccessiveSizes_go_64KB", mk(goasm.ReadSuccessiveSizes_go, 1<<16-1)},
	{"ReadSuccessiveSizes_go_128KB", mk(goasm.ReadSuccessiveSizes_go, 1<<17-1)},
	{"ReadSuccessiveSizes_go_256KB", mk(goasm.ReadSuccessiveSizes_go, 1<<18-1)},
	{"ReadSuccessiveSizes_go_512KB", mk(goasm.ReadSuccessiveSizes_go, 1<<19-1)},
	{"ReadSuccessiveSizes_go_4MB", mk(goasm.ReadSuccessiveSizes_go, 1<<22-1)},
	{"ReadSuccessiveSizes_go_8MB", mk(goasm.ReadSuccessiveSizes_go, 1<<23-1)},
	{"ReadSuccessiveSizes_go_16MB", mk(goasm.ReadSuccessiveSizes_go, 1<<24-1)},
	{"ReadSuccessiveSizes_go_32MB", mk(goasm.ReadSuccessiveSizes_go, 1<<25-1)},
	{"ReadSuccessiveSizes_go_64MB", mk(goasm.ReadSuccessiveSizes_go, 1<<26-1)},
	{"ReadSuccessiveSizes_go_1GB", mk(goasm.ReadSuccessiveSizes_go, 1<<30-1)},
	{"ReadSuccessiveSizes_4KB", mk(asm.ReadSuccessiveSizes, 1<<12-1)},
	{"ReadSuccessiveSizes_32KB", mk(asm.ReadSuccessiveSizes, 1<<15-1)},
	{"ReadSuccessiveSizes_64KB", mk(asm.ReadSuccessiveSizes, 1<<16-1)},
	{"ReadSuccessiveSizes_128KB", mk(asm.ReadSuccessiveSizes, 1<<17-1)},
	{"ReadSuccessiveSizes_256KB", mk(asm.ReadSuccessiveSizes, 1<<18-1)},
	{"ReadSuccessiveSizes_512KB", mk(asm.ReadSuccessiveSizes, 1<<19-1)},
	{"ReadSuccessiveSizes_4MB", mk(asm.ReadSuccessiveSizes, 1<<22-1)},
	{"ReadSuccessiveSizes_8MB", mk(asm.ReadSuccessiveSizes, 1<<23-1)},
	{"ReadSuccessiveSizes_16MB", mk(asm.ReadSuccessiveSizes, 1<<24-1)},
	{"ReadSuccessiveSizes_32MB", mk(asm.ReadSuccessiveSizes, 1<<25-1)},
	{"ReadSuccessiveSizes_64MB", mk(asm.ReadSuccessiveSizes, 1<<26-1)},
	{"ReadSuccessiveSizes_1GB", mk(asm.ReadSuccessiveSizes, 1<<30-1)},
}

func main() {
	const nbytes = 1 << 30
	params := Params{count: nbytes, buf: make([]byte, nbytes)}
	for i := range params.buf {
		params.buf[i] = byte(i)
	}
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

func mk(fn func(uint64, []byte, uint64), mask uint64) func(*reptest.Tester, Params) {
	return func(rt *reptest.Tester, p Params) {
		for rt.IsTesting() {
			rt.BeginTime()
			fn(p.count, p.buf, mask)
			rt.EndTime()
			rt.CountBytes(p.count)
		}
	}
}

func ulen[E any](u []E) uint64 {
	return uint64(len(u))
}
