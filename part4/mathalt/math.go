package mathalt

import (
	"log"
	"math"
)

const collectStats = true

// Store some statistics on function range
type StatEntry struct{ InMin, InMax float64 }
type Statistics struct{ Sin, Cos, Asin, Sqrt StatEntry }

// Initialized in init()
var stats Statistics

func init() {
	if collectStats {
		stats.Sin.InMin = math.Inf(1)
		stats.Cos.InMin = math.Inf(1)
		stats.Asin.InMin = math.Inf(1)
		stats.Sqrt.InMin = math.Inf(1)
		stats.Sin.InMax = math.Inf(-1)
		stats.Cos.InMax = math.Inf(-1)
		stats.Asin.InMax = math.Inf(-1)
		stats.Sqrt.InMax = math.Inf(-1)
	}
}

// Wrappers around math functions that also record max/min of input.
func Sin(x float64) float64 {
	if collectStats {
		stats.Sin.InMin = min(x, stats.Sin.InMin)
		stats.Sin.InMax = max(x, stats.Sin.InMax)
	}
	return math.Sin(x)
}

func Cos(x float64) float64 {
	if collectStats {
		stats.Cos.InMin = min(x, stats.Cos.InMin)
		stats.Cos.InMax = max(x, stats.Cos.InMax)
	}
	return math.Cos(x)
}

func Asin(x float64) float64 {
	if collectStats {
		stats.Asin.InMin = min(x, stats.Asin.InMin)
		stats.Asin.InMax = max(x, stats.Asin.InMax)
	}
	return math.Asin(x)
}

func Sqrt(x float64) float64 {
	if collectStats {
		stats.Sqrt.InMin = min(x, stats.Sqrt.InMin)
		stats.Sqrt.InMax = max(x, stats.Sqrt.InMax)
	}
	return math.Sqrt(x)
}

func PrintReport() {
	log.Println("Function statistics report, input min/max:")
	log.Printf(" Sin: % 2.7f  | % 2.7f", stats.Sin.InMin, stats.Sin.InMax)
	log.Printf(" Cos: % 2.7f  | % 2.7f", stats.Cos.InMin, stats.Cos.InMax)
	log.Printf("Asin: % 2.7f  | % 2.7f", stats.Asin.InMin, stats.Asin.InMax)
	log.Printf("Sqrt: % 2.7f  | % 2.7f", stats.Sqrt.InMin, stats.Sqrt.InMax)
}
