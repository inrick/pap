// Code generated by "stringer -type ProfileKind"; DO NOT EDIT.

package profiler

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[KindNone-0]
	_ = x[KindReadInputFile-1]
	_ = x[KindParsePairs-2]
	_ = x[KindParsePair-3]
	_ = x[KindParseNumber-4]
	_ = x[KindParseFloat-5]
	_ = x[KindCalculateDistances-6]
	_ = x[KindReadReferenceFile-7]
	_ = x[KindCompareReferenceFile-8]
	_ = x[KindTotalRuntime-9]
	_ = x[KindCount-10]
}

const _ProfileKind_name = "KindNoneKindReadInputFileKindParsePairsKindParsePairKindParseNumberKindParseFloatKindCalculateDistancesKindReadReferenceFileKindCompareReferenceFileKindTotalRuntimeKindCount"

var _ProfileKind_index = [...]uint8{0, 8, 25, 39, 52, 67, 81, 103, 124, 148, 164, 173}

func (i ProfileKind) String() string {
	if i < 0 || i >= ProfileKind(len(_ProfileKind_index)-1) {
		return "ProfileKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ProfileKind_name[_ProfileKind_index[i]:_ProfileKind_index[i+1]]
}