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
	TargetBytes uint64
	Label       string
	SizeLabel   string
	ChunkSize   uint64
	Func        TestFn
}

var TestFunctions []TestFnSpec

func main() {
	flagOutput := flag.String("o", "", "output file")
	flag.Parse()

	const (
		startBytes   = 1 << 10
		stopBytes    = 1 << 30
		stepSizeInit = 1 << 10
		bufsz        = stopBytes
	)
	params := Params{count: bufsz, buf: make([]byte, bufsz)}
	for i := range params.buf {
		params.buf[i] = byte(i)
	}

	// Populate TestFunctions

	// Written in Go assembly
	for n, stepSize := uint64(startBytes), uint64(stepSizeInit); n <= stopBytes; n += stepSize {
		strBytes := readableBytes(n)
		chunkSize := (n / 256) * 256
		testFn := TestFnSpec{
			Name:        "ReadSuccessiveSizesNonPow2_go",
			TargetBytes: (bufsz / chunkSize) * chunkSize,
			Label:       "go",
			SizeLabel:   strBytes,
			ChunkSize:   n,
			Func:        mk(goasm.ReadSuccessiveSizesNonPow2_go, n),
		}
		TestFunctions = append(TestFunctions, testFn)
		// Increase stepSize when we hit a power of 2, but never lower than
		// the initial step size.
		if (n-1)&n == 0 {
			stepSize = max(n>>1, stepSizeInit)
		}
	}

	// Compiled with nasm
	for n, stepSize := uint64(startBytes), uint64(stepSizeInit); n <= stopBytes; n += stepSize {
		strBytes := readableBytes(n)
		chunkSize := (n / 256) * 256
		testFn := TestFnSpec{
			Name:        "ReadSuccessiveSizesNonPow2",
			TargetBytes: (bufsz / chunkSize) * chunkSize,
			Label:       "nasm",
			SizeLabel:   strBytes,
			ChunkSize:   n,
			Func:        mk(asm.ReadSuccessiveSizesNonPow2, n),
		}
		TestFunctions = append(TestFunctions, testFn)
		// Increase stepSize when we hit a power of 2, but never lower than
		// the initial step size.
		if (n-1)&n == 0 {
			stepSize = max(n>>1, stepSizeInit)
		}
	}

	testers := make([]reptest.Tester, len(TestFunctions))
	freqReport := internal.EstimateCpuFrequency(10)

	var results []reptest.FinalTestResults
	for i, testFn := range TestFunctions {
		rt := &testers[i]
		fmt.Printf("\n--- %s %s ---\n", testFn.Name, testFn.SizeLabel)
		rt.NewTestWave(testFn.TargetBytes, freqReport.EstFreq, 10)
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
	fmt.Fprintln(w, "Function,Label,Size label,Chunk size,Max GB/s,Min GB/s,Avg GB/s")
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
			TestFunctions[i].SizeLabel,
			TestFunctions[i].ChunkSize,
			bandwidth(res.MinTime),
			bandwidth(res.MaxTime),
			bandwidth(avgTime),
		)
	}
}

func mk(fn func(uint64, []byte, uint64), chunkSize uint64) TestFn {
	return func(rt *reptest.Tester, p Params) reptest.FinalTestResults {
		// Divide and remultiply because it might not evenly divide.
		count := (p.count / chunkSize) * chunkSize
		for rt.IsTesting() {
			rt.BeginTime()
			fn(count, p.buf, chunkSize)
			rt.EndTime()
			rt.CountBytes(count)
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
