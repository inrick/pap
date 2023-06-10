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
	"strconv"
	"time"
	"unsafe"
)

var (
	ErrTooFew      = errors.New("too few bytes read")
	ErrNoMoreInput = errors.New("no input left")
	ErrExpectedEof = errors.New("expected EOF")
)

func main() {
	log.SetFlags(0)
	var cpuProf, memProf string
	flag.StringVar(&cpuProf, "cpuprof", "", "cpu profile output file")
	flag.StringVar(&memProf, "memprof", "", "mem profile output file")
	flag.Parse()
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
	t0 := time.Now()
	inputFile := args[0]
	var comparisonFile string
	if len(args) == 2 {
		comparisonFile = args[1]
	}
	buf, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err)
	}
	pp, err := ParsePairs(buf)
	if err != nil {
		log.Fatalf("%s failed to parse: %v", inputFile, err)
	}
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
		for i, d0 := range dists {
			d1 := distsRef[i]
			// TODO: choose appropriate epsilon
			if 1e-9 < math.Abs(d0-d1) {
				log.Fatalf("difference detected: %f != %f", d0, d1)
			}
		}
		log.Print("result identical to reference file")
		log.Printf("Average=%f", avg)
	}
	t1 := time.Now()
	dt := t1.Sub(t0).Seconds()
	log.Printf("%s processed successfully in %.3f seconds", inputFile, dt)
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
	n, err := strconv.ParseFloat(string(tok), 64)
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

func Radians(deg float64) float64 {
	return (math.Pi / 180) * deg
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
