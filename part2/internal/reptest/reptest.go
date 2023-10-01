package reptest

import (
	"fmt"
	"part2/internal"
)

type TestMode int32

const (
	TestModeUninitialized = iota
	TestModeTesting
	TestModeCompleted
	TestModeError
)

type Tester struct {
	state       TestMode
	targetBytes uint64
	timerFreq   uint64
	tryForTime  uint64
	startedAt   uint64
	printNewMin bool

	nOpenBlock  int
	nCloseBlock int
	timeAcc     uint64
	bytesAcc    uint64

	results TestResults
}

type TestResults struct {
	testCount uint64
	totalTime uint64
	minTime   uint64
	maxTime   uint64
}

func (rt *Tester) NewTestWave(targetBytes, timerFreq, secondsToTry uint64) {
	if rt.state == TestModeUninitialized {
		rt.state = TestModeTesting
		rt.targetBytes = targetBytes
		rt.timerFreq = timerFreq
		rt.printNewMin = true
		rt.results.minTime = 1<<63 - 1
	} else if rt.state == TestModeCompleted {
		rt.state = TestModeTesting
		if rt.targetBytes != targetBytes {
			rt.Error(fmt.Errorf(
				"targetBytes changed from %d to %d", rt.targetBytes, targetBytes,
			))
		}
		if rt.timerFreq != timerFreq {
			rt.Error(fmt.Errorf(
				"timerFreq changed from %d to %d", rt.timerFreq, timerFreq,
			))
		}
	}
	rt.tryForTime = secondsToTry * timerFreq
	rt.startedAt = internal.Rdtsc()
}

func (rt *Tester) BeginTime() {
	rt.nOpenBlock++
	rt.state = TestModeTesting
	rt.timeAcc -= internal.Rdtsc()
}

func (rt *Tester) EndTime() {
	rt.nCloseBlock++
	rt.timeAcc += internal.Rdtsc()
}

func (rt *Tester) CountBytes(n uint64) { rt.bytesAcc += n }

func (rt *Tester) Error(err error) {
	fmt.Printf("ERROR: %v", err)
	rt.state = TestModeError
}

func (rt *Tester) IsTesting() bool {
	if rt.state != TestModeTesting {
		return false
	}
	currentTime := internal.Rdtsc()
	if rt.nOpenBlock > 0 { // Tests without timing blocks are not counted
		if rt.nOpenBlock != rt.nCloseBlock {
			rt.Error(fmt.Errorf(
				"open block not equal to closed block count (%d != %d)",
				rt.nOpenBlock, rt.nCloseBlock,
			))
		}
		if rt.targetBytes != rt.bytesAcc {
			rt.Error(fmt.Errorf(
				"target bytes count not equal to accumulated bytes count (%d != %d)",
				rt.targetBytes, rt.bytesAcc,
			))
		}
		if rt.state != TestModeTesting {
			return false
		}
		elapsed := rt.timeAcc
		rt.results.testCount++
		rt.results.totalTime += elapsed
		rt.results.maxTime = max(rt.results.maxTime, elapsed)
		if rt.results.minTime > elapsed {
			rt.results.minTime = elapsed
			// If we hit a new time we reset the total clock to rerun for
			// "tryForTime" again, trying to find a new minimum.
			rt.startedAt = currentTime
			if rt.printNewMin {
				PrintTimeU("Min", rt.results.minTime, rt.timerFreq, rt.bytesAcc)
				fmt.Printf("        \r")
			}
		}
		rt.nOpenBlock = 0
		rt.nCloseBlock = 0
		rt.timeAcc = 0
		rt.bytesAcc = 0
	}
	if currentTime-rt.startedAt > rt.tryForTime {
		rt.state = TestModeCompleted
		PrintResults(rt.results, rt.timerFreq, rt.targetBytes)
	}
	return rt.state == TestModeTesting
}

func PrintTime(label string, cpuTime float64, timerFreq uint64, byteCount uint64) {
	fmt.Printf("%s: %.0f", label, cpuTime)
	if timerFreq == 0 {
		return
	}
	secs := cpuTime / float64(timerFreq)
	fmt.Printf(" (%f ms)", 1000*secs)
	if byteCount > 0 {
		const GB = 1 << 30
		bandwidth := float64(byteCount) / (GB * secs)
		fmt.Printf(" %f GB/s", bandwidth)
	}
}

func PrintTimeU(label string, cpuTime, timerFreq, byteCount uint64) {
	PrintTime(label, float64(cpuTime), timerFreq, byteCount)
}

func PrintResults(results TestResults, timerFreq uint64, byteCount uint64) {
	PrintTimeU("Min", results.minTime, timerFreq, byteCount)
	fmt.Println()
	PrintTimeU("Max", results.maxTime, timerFreq, byteCount)
	fmt.Println()
	if results.testCount > 0 {
		avgCpuTime := float64(results.totalTime) / float64(results.testCount)
		PrintTime("Avg", avgCpuTime, timerFreq, byteCount)
		fmt.Println()
	}
}
