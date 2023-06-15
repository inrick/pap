package main

import (
	"errors"
	"math"
	"os"
	"path"
	"testing"
)

func sliceEq[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func TestParsePairs(t *testing.T) {
	for _, test := range []struct {
		input    string
		expected []Pair
	}{
		{`{"pairs":[]}`, []Pair{}},
		{`{"pairs":[{"x0":0,"y0":0,"x1":0,"y1":0}]}`, []Pair{{0, 0, 0, 0}}},
		{`{"pairs":[{"x0":0,"y0":0,"x1":0,"y1":0},{"x0":1.2345,"y0":0,"x1":-987.654321,"y1":0}]}`, []Pair{{0, 0, 0, 0}, {1.2345, 0, -987.654321, 0}}},
	} {
		pair, err := ParsePairs([]byte(test.input))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !sliceEq(test.expected, pair) {
			t.Errorf("want %v, got %v", test.expected, pair)
		}
	}
}

func TestHaversineCalc(t *testing.T) {
	for i, test := range []struct {
		input       []byte
		expectedAvg float64
	}{
		{must(os.ReadFile(path.Join("testdata", "test1.json"))), 1307.029720},
		{must(os.ReadFile(path.Join("testdata", "test10.json"))), 4086.975125},
		{must(os.ReadFile(path.Join("testdata", "test100.json"))), 3215.237987},
	} {
		pairs := must(ParsePairs(test.input))
		_, avg := Distances(pairs)
		const eps = 1e-6
		if diff := math.Abs(test.expectedAvg - avg); diff > eps {
			t.Errorf(
				"difference too big in test case %d; %.16f > %.16f, %f != %f",
				i, diff, eps, test.expectedAvg, avg,
			)
		}
	}
}

func TestParseFloat(t *testing.T) {
	for i, test := range []struct {
		input    string
		expected float64
	}{
		{"1.", 1},
		{".1", .1},
		{"1.0", 1},
		{"1.123456789012345", 1.123456789012345},
		{"-1", -1},
		{"-1.123456789012345", -1.123456789012345},
		{"198273123.1231231", 198273123.1231231},
		{"3.333333333333334", 3.333333333333334},
		{".0000000123456789", .0000000123456789},
		// Test cases below exemplify lost precision when parsed as float64
		{"9007199254740992.99999999", 9007199254740992},
		{"-9007199254740992.99999999", -9007199254740992},
		{"-9007199254740993", -9007199254740992},
		{".0000000123456789999", .0000000123456789},
	} {
		x, err := ParseFloat([]byte(test.input))
		if err != nil {
			t.Errorf("could not parse float %s: %v", test.input, err)
			continue
		}
		const eps = 1e-16
		if diff := math.Abs(test.expected - x); diff > eps {
			t.Errorf(
				"difference too big in test case %d; %.16f > %.16f, %f != %f",
				i, diff, eps, test.expected, x,
			)
		}
	}
}

func TestParseFloatFail(t *testing.T) {
	for i, test := range []struct {
		input string
		err   error
	}{
		{"", ErrParseFloatEmpty},
		{"123.456.789", ErrParseFloatDecimal},
		{"-5-3", ErrParseFloatUnknown},
		{"123abc", ErrParseFloatUnknown},
	} {
		_, err := ParseFloat([]byte(test.input))
		if !errors.Is(err, test.err) {
			t.Errorf("test case %d: got error \"%v\", want \"%v\"", i, err, test.err)
		}
	}
}
