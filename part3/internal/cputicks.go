package internal

import (
	"log"
	"time"
)

func Rdtsc() uint64

func Rdtscp() uint64

func Rdtsc2() uint64

func Rdtscp2() uint64

func CpuidFreqMhz() uint64

type CpuFreqReport struct {
	Nanos   uint64
	Rdtsc0  uint64
	Rdtsc1  uint64
	EstFreq uint64
}

func EstimateCpuFrequency(millisToWait uint64) CpuFreqReport {
	t0 := time.Now()
	t1 := t0
	tsc0 := Rdtsc()
	toWait := 1e6 * millisToWait
	var dt uint64
	for dt < toWait {
		t1 = time.Now()
		dt = uint64(t1.Sub(t0).Nanoseconds())
	}
	tsc1 := Rdtsc()
	dtsc := tsc1 - tsc0
	cpuFreqEst := dtsc * 1000 / millisToWait
	return CpuFreqReport{
		Nanos:   dt,
		Rdtsc0:  tsc0,
		Rdtsc1:  tsc1,
		EstFreq: cpuFreqEst,
	}
}

func PrintCpuFrequency() {
	r := EstimateCpuFrequency(10)
	log.Printf("OS timer elapsed: %d nanoseconds (%.2f seconds)",
		r.Nanos, float64(r.Nanos)/1e9,
	)
	log.Printf(
		"CPU timer: %d -> %d = %d elapsed\n",
		r.Rdtsc0, r.Rdtsc1, r.Rdtsc1-r.Rdtsc0,
	)
	log.Printf(
		"CPU frequency estimate = %d (%.2f MHz)",
		r.EstFreq, float64(r.EstFreq)/1e6,
	)
	log.Printf("CPU frequency reported by cpuid: %d MHz", CpuidFreqMhz())
}

func TimestampToSec(ts uint64, estFreq uint64) float64 {
	return float64(ts) / float64(estFreq)
}
