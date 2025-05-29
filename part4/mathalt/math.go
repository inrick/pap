package mathalt

//go:generate go run ./math_gen.go -out math.s -stubs stubs.go

import (
	"log"
	"math"
)

const collectStats = true

// Store some statistics on function range
type StatEntry struct{ InMin, InMax float64 }
type Statistics struct{ Sin, Cos, Asin, Sqrt StatEntry }

func collectStat(e *StatEntry, x float64) {
	if collectStats {
		e.InMin = min(x, e.InMin)
		e.InMax = max(x, e.InMax)
	}
}

// Initialized in init()
var stats Statistics

func init() {
	if collectStats {
		for _, e := range []*StatEntry{
			&stats.Sin, &stats.Cos, &stats.Asin, &stats.Sqrt,
		} {
			e.InMin = math.Inf(1)
			e.InMax = math.Inf(-1)
		}
	}
}

func Abs(x float64) float64 {
	return math.Float64frombits(math.Float64bits(x) &^ (1 << 63))
}

// Wrappers around math functions that also record max/min of input.
func Sin(x float64) float64 {
	collectStat(&stats.Sin, x)
	return math.Sin(x)
}

func Cos(x float64) float64 {
	collectStat(&stats.Cos, x)
	return math.Cos(x)
}

func Asin(x float64) float64 {
	collectStat(&stats.Asin, x)
	return math.Asin(x)
}

func Sqrt(x float64) float64 {
	collectStat(&stats.Sqrt, x)
	return math.Sqrt(x)
}

func PrintReport() {
	log.Println("Function statistics report, input min/max:")
	log.Printf(" Sin: % 2.7f  | % 2.7f", stats.Sin.InMin, stats.Sin.InMax)
	log.Printf(" Cos: % 2.7f  | % 2.7f", stats.Cos.InMin, stats.Cos.InMax)
	log.Printf("Asin: % 2.7f  | % 2.7f", stats.Asin.InMin, stats.Asin.InMax)
	log.Printf("Sqrt: % 2.7f  | % 2.7f", stats.Sqrt.InMin, stats.Sqrt.InMax)
}
