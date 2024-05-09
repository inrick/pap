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

type TestFn func(uint64, []byte, uint64, uint64)

type TestFnSpec struct {
	Name        string
	Label       string
	StrideSize  uint64
	StrideLabel string
	Func        TestFn
}

var TestSpecs []TestFnSpec

func main() {
	flagOutput := flag.String("o", "", "output file")
	flagCode := flag.String("source", "go", "source function to use (go/nasm/both)")
	flag.Parse()

	var useGoAsm, useNasm bool
	switch *flagCode {
	case "go":
		useGoAsm = true
	case "nasm":
		useNasm = true
	case "both":
		useGoAsm = true
		useNasm = true
	default:
		fmt.Printf("Unknown option %q: valid alternatives are go/nasm/both\n", *flagCode)
		os.Exit(1)
	}

	const (
		cacheLineSize = 64
		chunkSize     = 256 * cacheLineSize
		total         = 64 * chunkSize
	)

	buf := make([]byte, 1<<20)
	for i := range buf {
		buf[i] = byte(i)
	}

	// Populate TestFunctions
	var strides []uint64
	for i := range uint64(128) {
		strides = append(strides, cacheLineSize*i)
	}

	sourceFns := []struct {
		fn    func(uint64, []byte, uint64, uint64)
		name  string
		label string
		use   bool
	}{
		{
			// Written in Go assembly
			goasm.ReadStrided_32x2_go,
			"ReadStrided_32x2_go",
			"go",
			useGoAsm,
		},
		{
			// Compiled with nasm
			asm.ReadStrided_32x2,
			"ReadStrided_32x2",
			"nasm",
			useNasm,
		},
	}

	for _, source := range sourceFns {
		if source.use {
			for _, n := range strides {
				spec := TestFnSpec{
					Name:       source.name,
					Label:      source.label,
					StrideSize: n,
					Func:       source.fn,
				}
				TestSpecs = append(TestSpecs, spec)
			}
		}
	}

	testers := make([]reptest.Tester, len(TestSpecs))
	freqReport := internal.EstimateCpuFrequency(10)

	var results []reptest.FinalTestResults
	for i, testSpec := range TestSpecs {
		rt := &testers[i]
		fmt.Printf("\n--- %s (stride %d bytes) ---\n", testSpec.Name, testSpec.StrideSize)
		rt.NewTestWave(total, freqReport.EstFreq, 10)
		for rt.IsTesting() {
			rt.BeginTime()
			testSpec.Func(total, buf, chunkSize, testSpec.StrideSize)
			rt.EndTime()
			rt.CountBytes(total)
		}
		res := rt.FinalTestResults()
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
	fmt.Fprintln(w, "Function,Label,Stride size,Max GB/s,Min GB/s,Avg GB/s")
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
			"%s,%s,%d,%f,%f,%f\n",
			TestSpecs[i].Name,
			TestSpecs[i].Label,
			TestSpecs[i].StrideSize,
			bandwidth(res.MinTime),
			bandwidth(res.MaxTime),
			bandwidth(avgTime),
		)
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
