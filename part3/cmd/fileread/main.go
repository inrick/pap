package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"part3/internal"
	"part3/internal/reptest"
	"path/filepath"
)

type TestFn func(*os.File, []byte, uint64) (uint64, error)

type TestFnSpec struct {
	Name    string
	BufSize uint64
	Func    TestFn
}

var TestSpecs []TestFnSpec

func readInputFile(filepath string) (*os.File, uint64, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, 0, fmt.Errorf("could not open file %q: %w", filepath, err)
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, uint64(st.Size()), nil
}

func main() {
	flagInput := flag.String("i", "", "input file")
	flagOutput := flag.String("o", "", "output file")
	flag.Parse()

	f, fileSize, err := readInputFile(*flagInput)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, sz := range []uint64{1 << 10, 1 << 15, 1 << 18, 1 << 19, 1 << 20, 1 << 21, 1 << 22, 1 << 23, 1 << 25, 1 << 30} {
		TestSpecs = append(TestSpecs, TestFnSpec{
			Name:    "ReadFile",
			BufSize: sz,
			Func:    TestFn(ReadFileChunked),
		})
	}

	buf := make([]byte, 0, fileSize+1)

	testers := make([]reptest.Tester, len(TestSpecs))
	freqReport := internal.EstimateCpuFrequency(10)

	var results []reptest.FinalTestResults
	for i, testSpec := range TestSpecs {
		rt := &testers[i]
		fmt.Printf("\n--- %s (%s) ---\n", testSpec.Name, readableBytes(testSpec.BufSize))
		rt.NewTestWave(fileSize, freqReport.EstFreq, 10)
		for rt.IsTesting() {
			// Reset input status
			buf = buf[:0]
			f.Seek(0, io.SeekStart)

			rt.BeginTime()
			written, err := testSpec.Func(f, buf, testSpec.BufSize)
			rt.EndTime()
			if err != nil {
				log.Fatal(err)
			}
			rt.CountBytes(written)
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
	fmt.Fprintln(w, "Function,Buffer size,Max GB/s,Min GB/s,Avg GB/s")
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
			"%s,%d,%f,%f,%f\n",
			TestSpecs[i].Name,
			TestSpecs[i].BufSize,
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
	return fmt.Sprintf("%d %s", n, units[i])
}

func ReadFileChunked(f *os.File, buf []byte, chunksz uint64) (uint64, error) {
	var total uint64
	for {
		if cap(buf) == len(buf) {
			panic("no room left to read file")
		}
		n, err := f.Read(buf[len(buf):min(len(buf)+int(chunksz), cap(buf))])
		buf = buf[:len(buf)+n]
		total += uint64(n)
		if err == io.EOF {
			return total, nil
		} else if err != nil {
			return total, err
		} else if n == 0 {
			return total, errors.New("f.Read returned 0 bytes")
		}
	}
}
