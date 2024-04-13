package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"part3/asm"
	"part3/goasm"
	"part3/internal"
	"part3/internal/reptest"
	"path/filepath"
)

type Params struct {
	count uint64
	buf   []byte
}

type TestFn func(*reptest.Tester, Params) reptest.FinalTestResults

type TestFnSpec struct {
	Name        string
	Label       string
	OffsetSize  uint64
	OffsetLabel string
	Func        TestFn
}

var TestFunctions []TestFnSpec

func main() {
	flagOutput := flag.String("o", "", "output file")
	flag.Parse()

	const bufsz = 1 << 30
	params := Params{count: bufsz, buf: make([]byte, bufsz)}
	for i := range params.buf {
		params.buf[i] = byte(i)
	}

	// Populate TestFunctions
	const startBytes, stopBytes = 0, 128

	// Written in Go assembly
	for n := uint64(0); n <= stopBytes; n++ {
		testFn := TestFnSpec{
			Name:        "ReadSuccessiveSizesNonPow2_go",
			Label:       "go",
			OffsetSize:  n,
			OffsetLabel: readableBytes(n),
			Func:        mk(goasm.ReadSuccessiveSizesNonPow2_go, n),
		}
		TestFunctions = append(TestFunctions, testFn)
	}

	// Compiled with nasm
	for n := uint64(0); n <= stopBytes; n++ {
		testFn := TestFnSpec{
			Name:        "ReadSuccessiveSizesNonPow2",
			Label:       "nasm",
			OffsetSize:  n,
			OffsetLabel: readableBytes(n),
			Func:        mk(asm.ReadSuccessiveSizesNonPow2, n),
		}
		TestFunctions = append(TestFunctions, testFn)
	}

	testers := make([]reptest.Tester, len(TestFunctions))
	freqReport := internal.EstimateCpuFrequency(10)

	var results []reptest.FinalTestResults
	for i, testFn := range TestFunctions {
		rt := &testers[i]
		fmt.Printf("\n--- %s %s ---\n", testFn.Name, testFn.OffsetLabel)
		rt.NewTestWave(bufsz, freqReport.EstFreq, 10)
		res := testFn.Func(rt, params)
		results = append(results, res)
	}

	var w io.Writer
	switch *flagOutput {
	case "", "-":
		w = os.Stdout
		fmt.Println()
		fmt.Println("--- Summary ---")
		fmt.Println()
	default:
		dir := filepath.Dir(*flagOutput)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Could not create directory: %v\n", err)
		}
		f, err := os.Create(*flagOutput)
		if err != nil {
			fmt.Printf("Could not create output file: %v\n", err)
			fmt.Println("Writing to stdout instead.")
			fmt.Println()
			w = os.Stdout
		} else {
			defer f.Close()
			fmt.Println()
			fmt.Printf("Printing results to %s\n", *flagOutput)
			w = f
		}
	}

	printCsvResults(w, results)
}

func printCsvResults(w io.Writer, results []reptest.FinalTestResults) {
	fmt.Fprintln(w, "Function,Label,Offset label,Offset size,Max GB/s,Min GB/s,Avg GB/s")
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
			"%s,%s,%s,%d,%f,%f,%f\n",
			TestFunctions[i].Name,
			TestFunctions[i].Label,
			TestFunctions[i].OffsetLabel,
			TestFunctions[i].OffsetSize,
			bandwidth(res.MinTime),
			bandwidth(res.MaxTime),
			bandwidth(avgTime),
		)
	}
}

func mk(fn func(uint64, []byte, uint64), offset uint64) TestFn {
	return func(rt *reptest.Tester, p Params) reptest.FinalTestResults {
		for rt.IsTesting() {
			rt.BeginTime()
			fn(p.count, p.buf[offset:], 32<<10)
			rt.EndTime()
			rt.CountBytes(p.count)
		}
		return rt.FinalTestResults()
	}
}

func readableBytes(n uint64) string {
	units := [...]string{"B", "KB", "MB", "GB"}
	i := 0
	x := float64(n)
	for i+1 < len(units) && 1<<10 <= x {
		x /= 1024
		i++
	}
	decimals := 0
	if 1e-2 < math.Abs(x-math.Floor(x)) {
		decimals = 1
	}
	return fmt.Sprintf(fmt.Sprintf("%%.%df %%s", decimals), x, units[i])
}
