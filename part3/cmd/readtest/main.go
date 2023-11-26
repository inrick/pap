package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"part3/internal"
	"part3/internal/reptest"
)

var TestFunctions = []struct {
	Name string
	Func func(*reptest.Tester, ReadFnParams)
}{
	{"(*os.File).Read (preallocated)", ReadFilePreallocTest},
	{"io.ReadAll (reallocates, begins with small buf)", IoReadAllTest},
	{"os.ReadFile (reallocates, allocates full size buf)", OsReadFileTest},
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s [filename]", os.Args[0])
	}
	fileName := os.Args[1]
	finfo, err := os.Stat(fileName)
	if err != nil {
		log.Fatal(err)
	}
	fileSize := uint64(finfo.Size())
	params := ReadFnParams{
		// Need an extra byte at the end to leave room for the final call to Read().
		Buf:      make([]byte, 0, fileSize+1),
		FileName: os.Args[1],
	}
	testers := make([]reptest.Tester, len(TestFunctions))
	freqReport := internal.EstimateCpuFrequency(10)
	for {
		for i, testFn := range TestFunctions {
			rt := &testers[i]
			fmt.Printf("\n--- %s ---\n", testFn.Name)
			rt.NewTestWave(fileSize, freqReport.EstFreq, 10)
			testFn.Func(rt, params)
		}
	}
}

type ReadFnParams struct {
	Buf      []byte
	FileName string
}

func ReadFileUsingBuf(buf []byte, f *os.File) ([]byte, uint64, error) {
	if len(buf) != 0 {
		panic("preallocated buf already used")
	}
	var total uint64
	for {
		if cap(buf) == len(buf) {
			panic("no room left to read file")
		}
		n, err := f.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		total += uint64(n)
		if err == io.EOF {
			return buf, total, nil
		} else if err != nil {
			return buf, total, err
		} else if n == 0 {
			return buf, total, errors.New("f.Read returned 0 bytes")
		}
	}
}

func IoReadAllTest(rt *reptest.Tester, params ReadFnParams) {
	for rt.IsTesting() {
		f, err := os.Open(params.FileName)
		if err != nil {
			rt.Error(err)
		} else {
			rt.BeginTime()
			buf, err := io.ReadAll(f)
			rt.EndTime()
			if err != nil {
				rt.Error(err)
			} else {
				rt.CountBytes(ulen(buf))
			}
			f.Close()
		}
	}
}

func OsReadFileTest(rt *reptest.Tester, params ReadFnParams) {
	for rt.IsTesting() {
		rt.BeginTime()
		buf, err := os.ReadFile(params.FileName)
		rt.EndTime()
		if err != nil {
			rt.Error(err)
		} else {
			rt.CountBytes(ulen(buf))
		}
	}
}

func ReadFilePreallocTest(rt *reptest.Tester, params ReadFnParams) {
	for rt.IsTesting() {
		// Reset preallocated buffer
		params.Buf = params.Buf[:0]
		f, err := os.Open(params.FileName)
		if err != nil {
			rt.Error(err)
			continue
		}
		rt.BeginTime()
		var n uint64
		params.Buf, n, err = ReadFileUsingBuf(params.Buf, f)
		rt.EndTime()
		rt.CountBytes(n)
		if err != nil {
			rt.Error(err)
		}
		f.Close()
	}
}

func ulen[T ~[]V, V any](t T) uint64 {
	return uint64(len(t))
}
