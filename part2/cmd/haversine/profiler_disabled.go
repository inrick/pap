//go:build noprofiler

package main

type Profiler struct{}
type ProfileResult struct{}

func ProfilerBegin(kind ProfileKind) {}
func ProfilerEnd(kind ProfileKind)   {}
func PrintReport()                   {}
