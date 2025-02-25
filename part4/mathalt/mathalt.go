package mathalt

import (
	"fmt"
	"math"
)

const (
	pi = 3.1415926535897

	mathAssert = true
)

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func assertRange(x, lo, hi float64) {
	if mathAssert {
		if x < lo || x > hi {
			panic(fmt.Errorf("invalid value for x for approximation: %.16f", x))
		}
	}
}

func SinQ(x float64) float64 {
	assertRange(x, -pi, pi)
	f := func(x float64) float64 {
		return 4 * x * (pi - x) / (pi * pi)
	}
	switch {
	case pi/2 <= x && x <= pi:
		return f(pi/2 - x)
	case 0 <= x && x < pi/2:
		return f(x)
	case -pi/2 <= x && x < 0:
		return -f(-x)
	case -pi <= x && x < -pi/2:
		return -f(pi/2 + x)
	}
	return x
}

func factorial(n int) float64 {
	a := float64(1)
	m := float64(n)
	for m > 0 {
		a *= m
		m--
	}
	return a
}

func SinTaylorN(n int) func(float64) float64 {
	return func(x float64) float64 {
		return SinTaylor(x, n)
	}
}

func SinTaylorHornerN(n int) func(float64) float64 {
	return func(x float64) float64 {
		return SinTaylorHorner(x, n)
	}
}

func SinTaylorHornerFMAN(n int) func(float64) float64 {
	return func(x float64) float64 {
		return SinTaylorHornerFMA(x, n)
	}
}

func SinTaylorHornerFMAAltN(n int) func(float64) float64 {
	return func(x float64) float64 {
		return SinTaylorHornerFMAAlt(x, n)
	}
}

// Return coefficient for nth term
func sinTaylorCoeff(n int) float64 {
	sign := float64(1 - 2*(n&1))
	return sign / factorial(2*n+1)
}

func SinTaylor(x float64, n int) float64 {
	y := float64(0)
	x2 := x * x
	for i := range n {
		y += x * sinTaylorCoeff(i)
		x *= x2
	}
	return y
}

func SinTaylorHorner(x float64, n int) float64 {
	x2 := x * x
	y := float64(0)
	for i := n; i > 0; i-- {
		y = y*x2 + sinTaylorCoeff(i-1)
	}
	y *= x
	return y
}

func SinTaylorHornerFMA(x float64, n int) float64 {
	x2 := x * x
	y := float64(0)
	for i := n; i > 0; i-- {
		y = math.FMA(y, x2, sinTaylorCoeff(i-1))
	}
	y *= x
	return y
}

func SinTaylorHornerFMAAlt(x float64, n int) float64 {
	x2 := x * x
	y := float64(0)
	for i := n; i > 0; i-- {
		y = FMAAlt(y, x2, sinTaylorCoeff(i-1))
	}
	y *= x
	return y
}

// SinAlt approximates sin(x) for x \in [-pi, pi] by two parabolas.
func SinAlt(x float64) float64 {
	assertRange(x, -pi, pi)

	ax := abs(x)
	y := 4 * ax * (pi - ax) / (pi * pi)
	if x < 0 {
		return -y
	}
	return y
}

// CosAlt approximates cos(x) for x \in [-pi/2, pi]
func CosAlt(x float64) float64 {
	assertRange(x, -pi/2-1e8, pi)

	switch {
	case x <= pi/2:
		return -(2*x + pi) * (2*x - pi) / (pi * pi)
	case x > pi/2:
		return (2*x - pi) * (2*x - 3*pi) / (pi * pi)
	}
	return x
}

func AsinAlt(x float64) float64 {
	return x
}
