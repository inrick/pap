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
	"runtime"
	"unsafe"

	"part4/internal"
	"part4/mathalt"
	"part4/profiler"
)

var (
	ErrTooFew      = errors.New("too few bytes read")
	ErrNoMoreInput = errors.New("no input left")
	ErrExpectedEof = errors.New("expected EOF")
)

var (
	SinFn  func(float64) float64
	CosFn  func(float64) float64
	AsinFn func(float64) float64
	SqrtFn func(float64) float64
	Pi     float64
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	profiler.PrintReport()
	log.Println()
	mathalt.PrintReport()
}

func run() error {
	defer profiler.End(profiler.Begin(profiler.KindTotalRuntime))
	log.SetFlags(0)
	log.SetPrefix("[haversine] ")
	var printFreq bool
	var useReferenceMathFns bool
	flag.BoolVar(&printFreq, "freq", false, "print estimated CPU frequency")
	flag.BoolVar(&useReferenceMathFns, "refmath", false, "use reference math functions")
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

	if useReferenceMathFns {
		SinFn = math.Sin
		CosFn = math.Cos
		AsinFn = math.Asin
		SqrtFn = math.Sqrt
		Pi = math.Pi
	} else {
		SinFn = mathalt.SinAlt
		CosFn = mathalt.CosAlt
		AsinFn = mathalt.AsinAlt
		SqrtFn = mathalt.SqrtAlt
		Pi = mathalt.Pi
	}

	dists, avg := Distances(pp)
	if comparisonFile != "" {
		distsRef, err := ReadReferenceFile(comparisonFile)
		if err != nil {
			return err
		}
		t, err := CompareReferenceFile(dists, avg, distsRef)
		if err != nil {
			return err
		}
		if len(t.diffs) == 0 {
			log.Print("result identical to reference file")
		} else {
			for _, d := range t.diffs {
				log.Printf("difference detected in pair %d: %.16f != %.16f", d.idx, d.dist, d.distRef)
			}
			log.Printf("total diff=%.16f", t.total)
		}
	}
	log.Printf("average=%.16f", avg)
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
) (TotalDiff, error) {
	expectedBytes := uint64(8 * (len(dists) + len(distsRef)))
	defer profiler.End(profiler.BeginWithBandwidth(profiler.KindCompareReferenceFile, expectedBytes))
	if N0, N1 := len(dists), len(distsRef); N0 != N1 {
		return TotalDiff{}, fmt.Errorf("different length to comparison file: %d != %d", N0, N1)
	}
	var diffs []Diff
	var total float64
	for i, d0 := range dists {
		d1 := distsRef[i]
		// TODO: choose appropriate epsilon
		const eps = 1e-9
		if eps < mathalt.Abs(d0-d1) {
			diffs = append(diffs, Diff{i, d0, d1})
			total += d0 - d1
		}
	}
	return TotalDiff{diffs: diffs, total: total}, nil
}

type TotalDiff struct {
	diffs []Diff
	total float64
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
	return (Pi / 180) * deg
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

func Square(x float64) float64 { return x * x }

func Haversine(p Pair) float64 {
	const earthRadius = 6372.8
	dLat := Radians(p.Y1 - p.Y0)
	dLon := Radians(p.X1 - p.X0)
	lat0 := Radians(p.Y0)
	lat1 := Radians(p.Y1)
	a := Square(SinFn(dLat/2)) + CosFn(lat0)*CosFn(lat1)*Square(SinFn(dLon/2))
	c := 2 * AsinFn(SqrtFn(a))
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
	// Again, remember to offset an extra 8 bytes.
	copy(dists, unsafe.Slice((*float64)(unsafe.Pointer(&buf[8])), N))
	runtime.KeepAlive(buf)
	return dists, nil
}
