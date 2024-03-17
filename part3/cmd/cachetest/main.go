package main

import (
	"fmt"
	"io"
	"os"
	"part3/asm"
	"part3/goasm"
	"part3/internal"
	"part3/internal/reptest"
)

type Params struct {
	count uint64
	buf   []byte
}

type TestFn func(*reptest.Tester, Params) reptest.FinalTestResults

type TestFnSpec struct {
	Name  string
	Label string
	Func  TestFn
}

var TestFunctions []TestFnSpec

func main() {
	const bufsz = 1 << 30
	params := Params{count: bufsz, buf: make([]byte, bufsz)}
	for i := range params.buf {
		params.buf[i] = byte(i)
	}

	// Populate TestFunctions
	const startBytes, stopBytes = 1 << 10, 1 << 30
	// Written in Go assembly
	for n := uint64(startBytes); n <= stopBytes; n <<= 1 {
		strBytes := readableBytes(n)
		testFn := TestFnSpec{
			Name:  fmt.Sprintf("ReadSuccessiveSizes_go_%s", strBytes),
			Label: strBytes,
			Func:  mk(goasm.ReadSuccessiveSizes_go, n-1),
		}
		TestFunctions = append(TestFunctions, testFn)
	}
	// Compiled with nasm
	for n := uint64(startBytes); n <= stopBytes; n <<= 1 {
		strBytes := readableBytes(n)
		testFn := TestFnSpec{
			Name:  fmt.Sprintf("ReadSuccessiveSizes_%s", strBytes),
			Label: strBytes,
			Func:  mk(asm.ReadSuccessiveSizes, n-1),
		}
		TestFunctions = append(TestFunctions, testFn)
	}

	testers := make([]reptest.Tester, len(TestFunctions))
	freqReport := internal.EstimateCpuFrequency(10)

	var results []reptest.FinalTestResults
	for i, testFn := range TestFunctions {
		rt := &testers[i]
		fmt.Printf("\n--- %s ---\n", testFn.Name)
		rt.NewTestWave(bufsz, freqReport.EstFreq, 10)
		res := testFn.Func(rt, params)
		results = append(results, res)
	}

	// Print CSV friendly output
	fmt.Println()
	fmt.Println("--- Summary ---")
	fmt.Println()
	printCsvResults(os.Stdout, results)
}

func printCsvResults(w io.Writer, results []reptest.FinalTestResults) {
	fmt.Fprintln(w, "Function,Label,Min GB/s,Max GB/s,Avg GB/s")
	for i, res := range results {
		bandwidth := func(t uint64) float64 {
			tf := float64(t)
			secs := tf / float64(res.TimerFreq)
			bandwidth := float64(res.ByteCount) / ((1 << 30) * secs)
			return bandwidth
		}
		avgTime := res.TotalTime / res.TestCount
		fmt.Fprintf(
			w,
			"%s,%s,%f,%f,%f\n",
			TestFunctions[i].Name,
			TestFunctions[i].Label,
			bandwidth(res.MinTime),
			bandwidth(res.MaxTime),
			bandwidth(avgTime),
		)
	}
}

func mk(fn func(uint64, []byte, uint64), mask uint64) TestFn {
	return func(rt *reptest.Tester, p Params) reptest.FinalTestResults {
		for rt.IsTesting() {
			rt.BeginTime()
			fn(p.count, p.buf, mask)
			rt.EndTime()
			rt.CountBytes(p.count)
		}
		return rt.FinalTestResults()
	}
}

func readableBytes(n uint64) string {
	units := [...]string{"B", "KB", "MB", "GB"}
	i := 0
	for i+1 < len(units) && 1<<10 <= n {
		n >>= 10
		i++
	}
	return fmt.Sprintf("%d%s", n, units[i])
}
