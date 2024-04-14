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
	ChunkSize   uint64
	ChunkLabel  string
	OffsetSize  uint64
	OffsetLabel string
	Func        TestFn
}

var TestFunctions []TestFnSpec

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

	const bufsz = 1 << 30
	// Make actual buffer slightly bigger so that we can read outside of it.
	params := Params{count: bufsz, buf: make([]byte, bufsz+1<<20)}
	for i := range params.buf {
		params.buf[i] = byte(i)
	}

	// Populate TestFunctions
	chunkSizes := []uint64{8 << 10, 128 << 10, 2 << 20, bufsz}
	alignments := []uint64{0, 1, 2, 15, 16, 17, 31, 32, 33, 47, 48, 49, 63, 64, 65}

	sourceFns := []struct {
		fn    func(uint64, []byte, uint64)
		name  string
		label string
		use   bool
	}{
		{
			// Written in Go assembly
			goasm.ReadSuccessiveSizesNonPow2_go,
			"ReadSuccessiveSizesNonPow2_go",
			"go",
			useGoAsm,
		},
		{
			// Compiled with nasm
			asm.ReadSuccessiveSizesNonPow2,
			"ReadSuccessiveSizesNonPow2",
			"nasm",
			useNasm,
		},
	}

	for _, source := range sourceFns {
		if source.use {
			for _, chunksz := range chunkSizes {
				for _, n := range alignments {
					testFn := TestFnSpec{
						Name:        source.name,
						Label:       source.label,
						ChunkSize:   chunksz,
						ChunkLabel:  readableBytes(chunksz),
						OffsetSize:  n,
						OffsetLabel: readableBytes(n),
						Func:        mk(source.fn, chunksz, n),
					}
					TestFunctions = append(TestFunctions, testFn)
				}
			}
		}
	}

	testers := make([]reptest.Tester, len(TestFunctions))
	freqReport := internal.EstimateCpuFrequency(10)

	var results []reptest.FinalTestResults
	for i, testFn := range TestFunctions {
		rt := &testers[i]
		fmt.Printf("\n--- %s (%s, %s) ---\n", testFn.Name, testFn.ChunkLabel, testFn.OffsetLabel)
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
	fmt.Fprintln(w, "Function,Label,Chunk label,Chunk size,Offset label,Offset size,Max GB/s,Min GB/s,Avg GB/s")
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
			"%s,%s,%s,%d,%s,%d,%f,%f,%f\n",
			TestFunctions[i].Name,
			TestFunctions[i].Label,
			TestFunctions[i].ChunkLabel,
			TestFunctions[i].ChunkSize,
			TestFunctions[i].OffsetLabel,
			TestFunctions[i].OffsetSize,
			bandwidth(res.MinTime),
			bandwidth(res.MaxTime),
			bandwidth(avgTime),
		)
	}
}

func mk(fn func(uint64, []byte, uint64), chunksz, offset uint64) TestFn {
	return func(rt *reptest.Tester, p Params) reptest.FinalTestResults {
		for rt.IsTesting() {
			rt.BeginTime()
			fn(p.count, p.buf[offset:], chunksz)
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
