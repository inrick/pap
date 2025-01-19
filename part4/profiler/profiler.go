//go:build !noprofiler

package profiler

import (
	"log"

	"part4/internal"
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
	results [KindCount]Result
}

var (
	prof          Profiler
	currProfScope ProfileKind
)

type Result struct {
	elapsedIncl uint64 // Including child blocks
	elapsedExcl uint64 // Excluding child blocks
	count       uint64
	bytesCount  uint64
}

type Block struct {
	kind   ProfileKind
	parent ProfileKind
	start  uint64
}

func Begin(kind ProfileKind) Block {
	bl := Block{
		kind:   kind,
		parent: currProfScope,
		start:  internal.Rdtsc(),
	}
	currProfScope = kind
	return bl
}

func BeginWithBandwidth(kind ProfileKind, bytesCount uint64) Block {
	prof.results[kind].bytesCount += bytesCount
	return Begin(kind)
}

func End(bl Block) {
	rec := &prof.results[bl.kind]
	elapsed := internal.Rdtsc() - bl.start
	rec.elapsedIncl += elapsed
	rec.elapsedExcl += elapsed
	prof.results[bl.parent].elapsedExcl -= elapsed
	rec.count++
	currProfScope = bl.parent
}

func PrintReport() {
	freq := internal.EstimateCpuFrequency(10)
	total := prof.results[KindTotalRuntime]
	dt := total.elapsedIncl
	tscToSec := func(c uint64) float64 {
		return float64(c) / float64(freq.EstFreq)
	}
	pct := func(n uint64) float64 {
		return 100 * float64(n) / float64(dt)
	}
	log.Print("Time report:")
	for kind := KindNone + 1; kind < KindCount; kind++ {
		r := &prof.results[kind]
		log.Printf(
			"%2d. %-25s [%9d] %11d cycles   %5.2f seconds   %6.2f %%   (excl. %5.2f seconds   %6.2f %%)",
			kind, kind.String(), r.count,
			r.elapsedIncl,
			tscToSec(r.elapsedIncl), pct(r.elapsedIncl),
			tscToSec(r.elapsedExcl), pct(r.elapsedExcl),
		)
		if r.bytesCount > 0 {
			seconds := tscToSec(r.elapsedIncl)
			bytesPerSec := float64(r.bytesCount) / seconds
			megabytes := float64(r.bytesCount) / (1 << 20)
			gbsPerSec := bytesPerSec / (1 << 30)
			log.Printf("    (processed %.3f MB at %.2f GB/s)", megabytes, gbsPerSec)
		}
	}
}
