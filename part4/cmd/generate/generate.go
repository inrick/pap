package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"
	"unsafe"
)

const (
	OutputFile  = "output.json"
	OutputDists = "output.f64"
)

var ErrTooFew = errors.New("too few bytes written")

func usage() {
	log.Fatalf("Usage: %s [flags] <uniform/cluster> <seed> <number of entries>", os.Args[0])
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("[generate] ")
	var (
		discard   bool
		pretty    bool
		outputDir string
	)
	flag.BoolVar(&discard, "discard", false, "discard generated output, do not create a file")
	flag.BoolVar(&pretty, "pretty", false, "pretty print output JSON file")
	flag.StringVar(&outputDir, "dir", ".", "output directory")
	flag.Parse()
	cfg := NewConfig(flag.Args())
	t0 := time.Now()
	var (
		pp    []Pair
		dists []float64
	)
	switch cfg.mode {
	case OutputUniform:
		pp, dists = Uniform(cfg)
	case OutputCluster:
		pp, dists = Cluster(cfg)
	}
	avg := Average(dists)
	t1 := time.Now()
	log.Printf("Generating points took %.3f seconds.", t1.Sub(t0).Seconds())
	if !discard {
		t0 = time.Now()
		f, err := os.Create(path.Join(outputDir, OutputFile))
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		defer w.Flush()
		enc := json.NewEncoder(w)
		if pretty {
			// Pretty printing takes a lot more time.
			enc.SetIndent("", "  ")
		}
		if err := enc.Encode(&Output{pp}); err != nil {
			log.Fatal(err)
		}
		t1 = time.Now()
		log.Printf("Wrote %q successfully, took %.3f seconds.", OutputFile, t1.Sub(t0).Seconds())
		t0 = time.Now()
		fReference, err := os.Create(path.Join(outputDir, OutputDists))
		if err != nil {
			log.Fatal(err)
		}
		defer fReference.Close()
		if err := WriteReference(fReference, dists); err != nil {
			log.Fatal(err)
		}
		t1 = time.Now()
		log.Printf("Wrote %q successfully, took %.3f seconds.", OutputDists, t1.Sub(t0).Seconds())
	}
	log.Printf("Average=%f", avg)
}

type Pair struct {
	X0 float64 `json:"x0"`
	Y0 float64 `json:"y0"`
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
}

type Output struct {
	Pairs []Pair `json:"pairs"`
}

type OutputMode int32

const (
	OutputUniform OutputMode = iota
	OutputCluster
)

type Config struct {
	mode    OutputMode
	seed    int64
	entries int
}

func NewConfig(args []string) Config {
	if len(args) != 3 {
		usage()
	}
	var mode OutputMode
	switch args[0] {
	case "uniform":
		mode = OutputUniform
	case "cluster":
		mode = OutputCluster
	default:
		usage()
	}
	seed, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		usage()
	}
	entries, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		usage()
	}
	if 1<<30 < entries {
		log.Fatalf("%d entries seems excessive, is it really intentional?", entries)
	}
	return Config{mode, seed, int(entries)}
}

func Uniform(cfg Config) ([]Pair, []float64) {
	rnd := rand.New(rand.NewSource(cfg.seed))
	pp := make([]Pair, cfg.entries)
	var dists []float64
	for i := 0; i < cfg.entries; i++ {
		pp[i].X0 = rnd.Float64()*360 - 180
		pp[i].Y0 = rnd.Float64()*180 - 90
		pp[i].X1 = rnd.Float64()*360 - 180
		pp[i].Y1 = rnd.Float64()*180 - 90
		dists = append(dists, Haversine(pp[i]))
	}
	return pp, dists
}

func Cluster(cfg Config) ([]Pair, []float64) {
	rnd := rand.New(rand.NewSource(cfg.seed))
	pp := make([]Pair, cfg.entries)
	clusters := rnd.Intn((100 + int(cfg.seed)) % 1024)
	var dists []float64
	pos := 0
	for i := 0; i < clusters; i++ {
		xmin := rnd.Float64()*360 - 180
		xmax := rnd.Float64()*360 - 180
		if xmin > xmax {
			xmax, xmin = xmin, xmax
		}
		ymin := rnd.Float64()*180 - 90
		ymax := rnd.Float64()*180 - 90
		if ymin > ymax {
			ymax, ymin = ymin, ymax
		}
		N := (len(pp) - pos) / (clusters - i)
		for j := 0; j < N; j++ {
			pi := pos + j
			pp[pi].X0 = rnd.Float64()*xmax - xmin
			pp[pi].Y0 = rnd.Float64()*ymax - ymin
			pp[pi].X1 = rnd.Float64()*xmax - xmin
			pp[pi].Y1 = rnd.Float64()*ymax - ymin
			dists = append(dists, Haversine(pp[pi]))
		}
		pos += N
	}
	return pp, dists
}

func Radians(deg float64) float64 {
	return (math.Pi / 180) * deg
}

func Square(x float64) float64 { return x * x }

func Average(xx []float64) float64 {
	var avg float64
	for _, x := range xx {
		avg += x
	}
	return avg / float64(len(xx))
}

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

// Writes output in binary format for later comparison. Writes data in byte
// order used by the machine it's running on, which in all likelihood is little
// endian.
//
// Format: 64 byte integer `n` specifying the length of the array; followed by
// `n` number of float64:s.
func WriteReference(w0 io.Writer, dists []float64) (err error) {
	w := bufio.NewWriter(w0)
	defer w.Flush()
	N := uint64(len(dists))
	buf := *(*[8]byte)(unsafe.Pointer(&N))
	if n, _ := w.Write(buf[:]); n != 8 {
		return ErrTooFew
	}
	for _, d := range dists {
		buf := *(*[8]byte)(unsafe.Pointer(&d))
		if n, _ := w.Write(buf[:]); n != 8 {
			return ErrTooFew
		}
	}
	return nil
}
