package mathalt

import (
	"fmt"
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

func factorial(n int) int {
	m := n
	for n > 1 {
		n--
		m *= n
	}
	return m
}

func pow(x float64, p int) float64 {
	y := float64(1)
	for p > 0 {
		y *= x
		p--
	}
	return y
}

func SinTaylorN(n int) func(float64) float64 {
	return func(x float64) float64 {
		return SinTaylor(x, n)
	}
}

func SinTaylor(x float64, n int) float64 {
	y := float64(0)
	sign := float64(-1)
	for i := range n {
		sign *= -1
		y += sign * pow(x, 2*i+1) / float64(factorial(2*i+1))
	}
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
