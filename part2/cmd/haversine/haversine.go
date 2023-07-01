package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"unsafe"
)

var (
	ErrTooFew      = errors.New("too few bytes read")
	ErrNoMoreInput = errors.New("no input left")
	ErrExpectedEof = errors.New("expected EOF")
)

func main() {
	var (
		timeStart      uint64
		timeReadInput  uint64
		timeParse      uint64
		timeComparison uint64
		timeEnd        uint64
	)
	timeStart = Rdtsc()
	log.SetFlags(0)
	log.SetPrefix("[haversine] ")
	var (
		cpuProf   string
		memProf   string
		printFreq bool
	)
	flag.StringVar(&cpuProf, "cpuprof", "", "cpu profile output file")
	flag.StringVar(&memProf, "memprof", "", "mem profile output file")
	flag.BoolVar(&printFreq, "freq", false, "print estimated CPU frequency")
	flag.Parse()
	if printFreq {
		PrintCpuFrequency()
		return
	}
	args := flag.Args()
	if nargs := len(args); nargs < 1 || 2 < nargs {
		log.Fatalf("usage: %s <file.json> [file.f64]", os.Args[0])
	}
	if cpuProf != "" {
		f, err := os.Create(cpuProf)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	inputFile := args[0]
	var comparisonFile string
	if len(args) == 2 {
		comparisonFile = args[1]
	}
	buf, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	timeReadInput = Rdtsc()
	pp, err := ParsePairs(buf)
	if err != nil {
		log.Fatalf("%s failed to parse: %v", inputFile, err)
	}
	timeParse = Rdtsc()
	if comparisonFile != "" {
		f, err := os.Open(comparisonFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		distsRef, err := ReadReference(f)
		if err != nil {
			log.Fatal(err)
		}
		if N0, N1 := len(pp), len(distsRef); N0 != N1 {
			log.Fatalf("different length to comparison file: %d != %d", N0, N1)
		}
		dists, avg := Distances(pp)
		allEqual := true
		for i, d0 := range dists {
			d1 := distsRef[i]
			// TODO: choose appropriate epsilon
			const eps = 1e-9
			if eps < math.Abs(d0-d1) {
				log.Printf("    > difference detected in pair %d: %f != %f", i, d0, d1)
				allEqual = false
			}
		}
		if allEqual {
			log.Print("result identical to reference file")
		}
		log.Printf("average=%f", avg)
		timeComparison = Rdtsc()
	}
	timeEnd = Rdtsc()
	dt := timeEnd - timeStart
	fr := EstimateCpuFrequency(10)
	tscToSec := func(c uint64) float64 {
		return float64(c) / float64(fr.EstFreq)
	}
	pct := func(n uint64) float64 {
		return 100 * float64(n) / float64(dt)
	}
	log.Printf("Time report:")
	log.Printf(
		"1. Reading input file: %d cycles / %.2f seconds (%.2f %%)",
		timeReadInput-timeStart,
		tscToSec(timeReadInput-timeStart),
		pct(timeReadInput-timeStart),
	)
	log.Printf(
		"2. Parsing input file: %d cycles / %.2f seconds (%.2f %%)",
		timeParse-timeReadInput,
		tscToSec(timeParse-timeReadInput),
		pct(timeParse-timeReadInput),
	)
	log.Printf(
		"3. Comparing with comparison file: %d cycles / %.2f seconds (%.2f %%)",
		timeComparison-timeParse,
		tscToSec(timeComparison-timeParse),
		pct(timeComparison-timeParse),
	)
	log.Printf(
		"4. %s processed successfully in a total of %d cycles / %.2f seconds",
		inputFile, dt, tscToSec(dt),
	)
	log.Printf(
		"(Estimated CPU frequency = %.2f MHz; Cpuid frequency reported = %d MHz)",
		float64(fr.EstFreq)/1e6, CpuidFreqMhz(),
	)
	if memProf != "" {
		f, err := os.Create(memProf)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
	}
}

func Distances(pp []Pair) ([]float64, float64) {
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
	a := Square(math.Sin(dLat/2)) + math.Cos(lat0)*math.Cos(lat1)*Square(math.Sin(dLon/2))
	c := 2 * math.Asin(math.Sqrt(a))
	return earthRadius * c
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
