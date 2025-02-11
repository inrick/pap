package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"unsafe"

	"part2/internal"
	"part2/profiler"
)

var (
	ErrTooFew      = errors.New("too few bytes read")
	ErrNoMoreInput = errors.New("no input left")
	ErrExpectedEof = errors.New("expected EOF")
)

// Store some statistics on function range
type StatEntry struct{ InMin, InMax float64 }
type Statistics struct{ Sin, Cos, Asin, Sqrt StatEntry }

// Initialized in init()
var stats Statistics

func init() {
	stats.Sin.InMin = math.Inf(1)
	stats.Cos.InMin = math.Inf(1)
	stats.Asin.InMin = math.Inf(1)
	stats.Sqrt.InMin = math.Inf(1)
	stats.Sin.InMax = math.Inf(-1)
	stats.Cos.InMax = math.Inf(-1)
	stats.Asin.InMax = math.Inf(-1)
	stats.Sqrt.InMax = math.Inf(-1)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	profiler.PrintReport()
	log.Println()
	stats.PrintReport()
}

func (s Statistics) PrintReport() {
	log.Println("Function statistics report, input min/max:")
	log.Printf(" Sin: % 2.7f  | % 2.7f", s.Sin.InMin, s.Sin.InMax)
	log.Printf(" Cos: % 2.7f  | % 2.7f", s.Cos.InMin, s.Cos.InMax)
	log.Printf("Asin: % 2.7f  | % 2.7f", s.Asin.InMin, s.Asin.InMax)
	log.Printf("Sqrt: % 2.7f  | % 2.7f", s.Sqrt.InMin, s.Sqrt.InMax)
}

func run() error {
	defer profiler.End(profiler.Begin(profiler.KindTotalRuntime))
	log.SetFlags(0)
	log.SetPrefix("[haversine] ")
	var printFreq bool
	flag.BoolVar(&printFreq, "freq", false, "print estimated CPU frequency")
	flag.Parse()
	if printFreq {
		internal.PrintCpuFrequency()
		return nil
	}
	args := flag.Args()
	if nargs := len(args); nargs < 1 || 2 < nargs {
		log.Fatalf("usage: %s <file.json> [file.f64]", os.Args[0])
	}
	inputFile := args[0]
	var comparisonFile string
	if len(args) == 2 {
		comparisonFile = args[1]
	}
	buf, err := ReadInputFile(inputFile)
	pp, err := ParsePairs(buf)
	if err != nil {
		return fmt.Errorf("%s failed to parse: %v", inputFile, err)
	}
	dists, avg := Distances(pp)
	if comparisonFile != "" {
		distsRef, err := ReadReferenceFile(comparisonFile)
		if err != nil {
			return err
		}
		diffs, err := CompareReferenceFile(dists, avg, distsRef)
		if err != nil {
			return err
		}
		if len(diffs) == 0 {
			log.Print("result identical to reference file")
		} else {
			for _, d := range diffs {
				log.Printf("difference detected in pair %d: %f != %f", d.idx, d.dist, d.distRef)
			}
		}
	}
	log.Printf("average=%f", avg)
	return nil
}

func ReadInputFile(file string) ([]byte, error) {
	stat, err := os.Stat(file)
	if err != nil {
		return nil, err
	}
	size := uint64(stat.Size())
	defer profiler.End(profiler.BeginWithBandwidth(profiler.KindReadInputFile, size))
	return os.ReadFile(file)
}

func CompareReferenceFile(
	dists []float64, avg float64, distsRef []float64,
) ([]Diff, error) {
	expectedBytes := uint64(8 * (len(dists) + len(distsRef)))
	defer profiler.End(profiler.BeginWithBandwidth(profiler.KindCompareReferenceFile, expectedBytes))
	if N0, N1 := len(dists), len(distsRef); N0 != N1 {
		return nil, fmt.Errorf("different length to comparison file: %d != %d", N0, N1)
	}
	var diffs []Diff
	for i, d0 := range dists {
		d1 := distsRef[i]
		// TODO: choose appropriate epsilon
		const eps = 1e-9
		if eps < math.Abs(d0-d1) {
			diffs = append(diffs, Diff{i, d0, d1})
		}
	}
	return diffs, nil
}

type Diff struct {
	idx     int
	dist    float64
	distRef float64
}

func Distances(pp []Pair) ([]float64, float64) {
	expectedBytes := uint64(len(pp)) * uint64(unsafe.Sizeof(pp[0]))
	defer profiler.End(profiler.BeginWithBandwidth(profiler.KindCalculateDistances, expectedBytes))
	dists := make([]float64, len(pp))
	N := float64(len(pp))
	var avg float64
	for i, p := range pp {
		d := Haversine(p)
		dists[i] = d
		avg += d / N
	}
	return dists, avg
}

// NOTE: shortcuts have been taken in the parser below. For instance when
// parsing identifiers and numbers, we are not following the grammar but the
// approach works in the common case. Potentially worth tightening up.

type PairsParser struct {
	buf []byte
	pos int
	err error
}

type Pair struct {
	X0, Y0, X1, Y1 float64
}

func ParsePairs(buf []byte) ([]Pair, error) {
	defer profiler.End(profiler.Begin(profiler.KindParsePairs))
	p := PairsParser{buf: buf}
	return p.Parse()
}

func (p *PairsParser) Parse() ([]Pair, error) {
	p.Expect('{')
	p.ExpectBytes([]byte(`"pairs"`))
	p.Expect(':')
	p.Expect('[')
	p.SkipSpace()
	var pp []Pair
	for p.Peek() != ']' {
		pair := p.ParsePair()
		pp = append(pp, pair)
		if p.Peek() != ',' || p.err != nil {
			break
		}
		p.Expect(',')
		p.SkipSpace()
	}
	p.Expect(']')
	p.Expect('}')
	p.ExpectEof()
	return pp, p.err
}

func (p *PairsParser) ParsePair() Pair {
	defer profiler.End(profiler.Begin(profiler.KindParsePair))
	var pair Pair
	if p.err != nil {
		return pair
	}
	p.Expect('{')
	for {
		field := p.Ident()
		p.Expect(':')
		n := p.Number()
		switch {
		case bytes.Equal(field, []byte("x0")):
			pair.X0 = n
		case bytes.Equal(field, []byte("y0")):
			pair.Y0 = n
		case bytes.Equal(field, []byte("x1")):
			pair.X1 = n
		case bytes.Equal(field, []byte("y1")):
			pair.Y1 = n
		default:
			p.err = fmt.Errorf("unknown field %s", string(field))
		}
		if p.Peek() != ',' || p.err != nil {
			break
		}
		p.Expect(',')
	}
	p.Expect('}')
	return pair
}

func (p *PairsParser) SkipSpace() {
	i := p.pos
	for ; i < len(p.buf) && IsSpace(p.buf[i]); i++ {
	}
	p.pos = i
}

func (p *PairsParser) Ident() []byte {
	p.Expect('"')
	start := p.pos
	p.SkipUntil('"')
	end := p.pos
	p.Expect('"')
	return p.buf[start:end]
}

func (p *PairsParser) Number() float64 {
	defer profiler.End(profiler.Begin(profiler.KindParseNumber))
	start := p.pos
	for p.pos < len(p.buf) && IsNumChar(p.buf[p.pos]) {
		p.pos++
	}
	end := p.pos
	tok := p.buf[start:end]
	n, err := ParseFloat(tok)
	if err != nil {
		p.err = err
	}
	return n
}

func (p *PairsParser) SkipUntil(c byte) {
	i := p.pos
	for ; i < len(p.buf) && p.buf[i] != c; i++ {
	}
	p.pos = i
}

func (p *PairsParser) Peek() byte {
	if len(p.buf) <= p.pos {
		return 0
	}
	return p.buf[p.pos]
}

func (p *PairsParser) ExpectEof() {
	p.SkipSpace()
	if p.err != nil {
		return
	}
	if len(p.buf) != p.pos {
		p.err = fmt.Errorf("%w but %d bytes left", ErrExpectedEof, len(p.buf)-p.pos)
	}
}

func (p *PairsParser) Expect(c byte) {
	if p.err != nil {
		return
	}
	p.SkipSpace()
	if found := p.buf[p.pos]; found != c {
		p.err = fmt.Errorf("expected character %c found %c", c, found)
	}
	p.pos++
}

func (p *PairsParser) ExpectBytes(b []byte) {
	if p.err != nil {
		return
	}
	p.SkipSpace()
	if len(p.buf) <= len(b)+p.pos {
		p.err = ErrNoMoreInput
		return
	}
	for i, c := range b {
		if c != p.buf[p.pos+i] {
			p.err = fmt.Errorf(
				"expected %s, found %s",
				string(b),
				string(p.buf[p.pos:p.pos+len(b)]),
			)
			return
		}
	}
	p.pos += len(b)
}

func IsSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\v' || c == '\n' || c == '\r'
}

func IsNumChar(c byte) bool {
	return ('0' <= c && c <= '9') || c == '.' || c == '-'
}

func IsDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func Radians(deg float64) float64 {
	return (math.Pi / 180) * deg
}

var (
	ErrParseFloatEmpty   = fmt.Errorf("ParseFloat: given empty string")
	ErrParseFloatDecimal = fmt.Errorf("ParseFloat: more than one decimal point")
	ErrParseFloatUnknown = fmt.Errorf("ParseFloat: unknown character")
)

func ParseFloat(s []byte) (float64, error) {
	defer profiler.End(profiler.Begin(profiler.KindParseFloat))
	if len(s) == 0 {
		return 0, ErrParseFloatEmpty
	}
	neg := s[0] == '-'
	decimal := false
	var err error
	n, i := float64(0), 0
	if neg {
		i = 1
	}
	for ; i < len(s) && err == nil && !decimal; i++ {
		switch {
		case s[i] == '.':
			decimal = true
		case IsDigit(s[i]):
			n = n*10 + float64(s[i]-'0')
		default:
			err = fmt.Errorf("%w %c", ErrParseFloatUnknown, s[i])
		}
	}
	// Parse everything after '.'
	q := float64(0)
	pow := float64(10)
	for ; i < len(s) && err == nil; i++ {
		switch {
		case s[i] == '.':
			err = ErrParseFloatDecimal
		case IsDigit(s[i]):
			q += float64(s[i]-'0') / pow
			pow *= 10
		}
	}
	n += q
	if neg {
		n = -n
	}
	return n, err
}

// Wrappers around math functions that also record max/min of input.
func Sin(x float64) float64 {
	stats.Sin.InMin = min(x, stats.Sin.InMin)
	stats.Sin.InMax = max(x, stats.Sin.InMax)
	return math.Sin(x)
}

func Cos(x float64) float64 {
	stats.Cos.InMin = min(x, stats.Cos.InMin)
	stats.Cos.InMax = max(x, stats.Cos.InMax)
	return math.Cos(x)
}

func Asin(x float64) float64 {
	stats.Asin.InMin = min(x, stats.Asin.InMin)
	stats.Asin.InMax = max(x, stats.Asin.InMax)
	return math.Asin(x)
}

func Sqrt(x float64) float64 {
	stats.Sqrt.InMin = min(x, stats.Sqrt.InMin)
	stats.Sqrt.InMax = max(x, stats.Sqrt.InMax)
	return math.Sqrt(x)
}

func Square(x float64) float64 { return x * x }

func Haversine(p Pair) float64 {
	const earthRadius = 6372.8
	dLat := Radians(p.Y1 - p.Y0)
	dLon := Radians(p.X1 - p.X0)
	lat0 := Radians(p.Y0)
	lat1 := Radians(p.Y1)
	a := Square(Sin(dLat/2)) + Cos(lat0)*Cos(lat1)*Square(Sin(dLon/2))
	c := 2 * Asin(Sqrt(a))
	return earthRadius * c
}

func ReadReferenceFile(refFile string) ([]float64, error) {
	stat, err := os.Stat(refFile)
	if err != nil {
		return nil, err
	}
	size := uint64(stat.Size())
	defer profiler.End(profiler.BeginWithBandwidth(profiler.KindReadReferenceFile, size))
	f, err := os.Open(refFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	distsRef, err := ReadReference(f)
	if err != nil {
		return nil, err
	}
	return distsRef, nil
}

// Read reference file format which is a length (int64) followed by that number
// of float64, describing the haversine distance of each pair of points.
func ReadReference(r io.Reader) ([]float64, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	// We need at least 8 bytes to read the length
	if len(buf) < 8 {
		return nil, ErrTooFew
	}
	N := *(*int64)(unsafe.Pointer(&buf[0]))
	// Make sure we have read the expected amount of data, otherwise reaching
	// into the underlying array below will be dangerous. The 8 in the beginning
	// is added because of the 8 bytes read above containing the length.
	if 8+8*N != int64(len(buf)) {
		return nil, ErrTooFew
	}
	dists := make([]float64, N)
	for i := range dists {
		// Again, remember to offset an extra 8 bytes.
		dists[i] = *(*float64)(unsafe.Pointer(&buf[8+8*i]))
	}
	return dists, nil
}
