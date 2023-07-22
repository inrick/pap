package main

// ProfileKind names the different profiler slots. It's placed in its own file
// so that both profiler.go and profiler_disabled.go can refer to the same
// type. Go only provides build tags at the file granularity level.

type ProfileKind int32

//go:generate stringer -type ProfileKind
const (
	ProfUnused ProfileKind = iota
	ProfReadInputFile
	ProfParsePairs
	ProfParsePair
	ProfParseNumber
	ProfParseFloat
	ProfCalculateDistances
	ProfReadReferenceFile
	ProfCompareReferenceFile
	ProfTotalRuntime
	ProfileCount
)
