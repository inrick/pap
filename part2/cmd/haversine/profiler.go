//go:build !noprofiler

package main

import (
	"log"

	"part2/internal"
)

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

var (
	prof          Profiler
	currProfScope ProfileKind
)

type ProfileResult struct {
	elapsedIncl uint64 // Including child blocks
	elapsedExcl uint64 // Excluding child blocks
	count       uint64
}

type ProfileBlock struct {
	kind   ProfileKind
	parent ProfileKind
	start  uint64
}

func ProfilerBegin(kind ProfileKind) ProfileBlock {
	bl := ProfileBlock{
		kind:   kind,
		parent: currProfScope,
		start:  internal.Rdtsc(),
	}
	currProfScope = kind
	return bl
}

func ProfilerEnd(bl ProfileBlock) {
	rec := &prof.results[bl.kind]
	elapsed := internal.Rdtsc() - bl.start
	rec.elapsedIncl += elapsed
	rec.elapsedExcl += elapsed
	prof.results[bl.parent].elapsedExcl -= elapsed
	rec.count++
	currProfScope = bl.parent
}

func PrintProfilerReport() {
	freq := internal.EstimateCpuFrequency(10)
	total := prof.results[ProfTotalRuntime]
	dt := total.elapsedIncl
	tscToSec := func(c uint64) float64 {
		return float64(c) / float64(freq.EstFreq)
	}
	pct := func(n uint64) float64 {
		return 100 * float64(n) / float64(dt)
	}
	log.Print("Time report:")
	for kind := ProfNone + 1; kind < ProfileCount; kind++ {
		r := &prof.results[kind]
		diff := r.elapsedIncl
		log.Printf(
			"%2d. %-25s [%9d] %11d cycles   %5.2f seconds   %6.2f %%   (excl. %5.2f seconds   %6.2f %%)",
			kind, kind.String(), r.count,
			diff, tscToSec(diff), pct(diff),
			tscToSec(r.elapsedExcl), pct(r.elapsedExcl),
		)
	}
}