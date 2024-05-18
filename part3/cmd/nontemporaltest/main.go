package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"part3/asm"
	"part3/goasm"
	"part3/internal"
	"part3/internal/reptest"
	"path/filepath"
)

type TestFn func([]byte, []byte, uint64, uint64)

type TestFnSpec struct {
	Name      string
	Label     string
	ChunkSize uint64
	Func      TestFn
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
		inputBufSz  = 1 << 20
		outputBufSz = 1 << 30
	)

	inbuf := make([]byte, inputBufSz)
	for i := range inbuf {
		inbuf[i] = byte(i)
	}
	outbuf := make([]byte, outputBufSz)

	// Populate TestFunctions

	sourceFns := []struct {
		fn    func([]byte, []byte, uint64, uint64)
		name  string
		label string
		use   bool
	}{
		{
			// Written in Go assembly
			goasm.WriteTemporal_go,
			"WriteTemporal_go",
			"go",
			useGoAsm,
		},
		{
			// Written in Go assembly
			goasm.WriteNonTemporal_go,
			"WriteNonTemporal_go",
			"go NT",
			useGoAsm,
		},
		{
			// Compiled with nasm
			asm.WriteTemporal,
			"WriteTemporal",
			"nasm",
			useNasm,
		},
		{
			// Compiled with nasm
			asm.WriteNonTemporal,
			"WriteNonTemporal",
			"nasm NT",
			useNasm,
		},
	}

	for _, source := range sourceFns {
		if source.use {
			for n := uint64(1 << 10); n <= inputBufSz; n <<= 1 {
				spec := TestFnSpec{
					Name:      source.name,
					Label:     source.label,
					ChunkSize: n,
					Func:      source.fn,
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
		fmt.Printf("\n--- %s (%s) ---\n", testSpec.Name, readableBytes(testSpec.ChunkSize))
		rt.NewTestWave(outputBufSz, freqReport.EstFreq, 10)
		for rt.IsTesting() {
			rt.BeginTime()
			testSpec.Func(inbuf, outbuf, outputBufSz, inputBufSz)
			rt.EndTime()
			rt.CountBytes(outputBufSz)
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
	fmt.Fprintln(w, "Function,Label,Chunk size,Max GB/s,Min GB/s,Avg GB/s")
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
			TestSpecs[i].ChunkSize,
			bandwidth(res.MinTime),
			bandwidth(res.MaxTime),
			bandwidth(avgTime),
		)
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
