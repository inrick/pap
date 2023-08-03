//go:build noprofiler

package main

type Profiler struct{}
type ProfileResult struct{}
type ProfileBlock struct{}

func ProfilerBegin(kind ProfileKind) (bl ProfileBlock) { return }
func ProfilerEnd(bl ProfileBlock)                      {}
func PrintProfilerReport()                             {}
