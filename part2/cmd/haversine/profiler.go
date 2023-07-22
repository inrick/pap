//go:build !noprofiler

package main

import "log"

// This profiler is not concurrency safe. If/when the code needs that, we will
// have to consider a different approach. Perhaps a profiler per goroutine
// which is then merged in the end?
//
// It also seems that the smallest functions are not able to be profiled like
// this; introducing the profiler to, for example, IsSpace, causes the runtime
// of the entire program to shoot up significantly. Even profiling Expect
// causes a significant slowdown.

type Profiler struct {
	results [ProfileCount]ProfileResult
}

var prof Profiler

type ProfileResult struct {
	begin   uint64
	elapsed uint64
}

func ProfilerBegin(kind ProfileKind) {
	prof.results[kind].begin = Rdtsc()
}

func ProfilerEnd(kind ProfileKind) {
	rec := &prof.results[kind]
	rec.elapsed += Rdtsc() - rec.begin
}

func PrintReport() {
	freq := EstimateCpuFrequency(10)
	total := prof.results[ProfTotalRuntime]
	dt := total.elapsed
	tscToSec := func(c uint64) float64 {
		return float64(c) / float64(freq.EstFreq)
	}
	pct := func(n uint64) float64 {
		return 100 * float64(n) / float64(dt)
	}
	log.Print("Time report:")
	for kind := ProfUnused + 1; kind < ProfileCount; kind++ {
		r := &prof.results[kind]
		diff := r.elapsed
		log.Printf(
			"%2d. %-25s %11d cycles   %5.2f seconds   %6.2f %%",
			kind, kind.String(), diff, tscToSec(diff), pct(diff),
		)
	}
}
