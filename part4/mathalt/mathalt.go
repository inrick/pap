package mathalt

import (
	"fmt"
	"math"
)

const (
	pi = 3.1415926535897

	mathAssert = true
)

// Generated using PrintSinTaylorCoeffs
var sinTaylorCoeffsArray = []float64{
	0x1p+00,
	-0x1.5555555555555p-03,
	0x1.1111111111111p-07,
	-0x1.a01a01a01a01ap-13,
	0x1.71de3a556c734p-19,
	-0x1.ae64567f544e4p-26,
	0x1.6124613a86d09p-33,
	-0x1.ae7f3e733b81fp-41,
	0x1.952c77030ad4ap-49,
	-0x1.2f49b46814157p-57,
	0x1.71b8ef6dcf572p-66,
	-0x1.761b413163819p-75,
	0x1.3f3ccdd165faap-84,
	-0x1.d1ab1c2dccea3p-94,
	0x1.259f98b4358aep-103,
	-0x1.434d2e783f5bcp-113,
	0x1.3981254dd0d5p-123,
	-0x1.0dc59c716d91fp-133,
	0x1.9ec8d1c94e85bp-144,
	-0x1.1e99449a4bacdp-154,
	0x1.65e61c39d0243p-165,
	-0x1.95db45257e511p-176,
	0x1.a3cb872220647p-187,
	-0x1.8da8e0a127eb7p-198,
	0x1.5a42f0dfeb083p-209,
	-0x1.161872bf7b825p-220,
	0x1.9d4f1058674e2p-232,
	-0x1.1d008faac5c55p-243,
	0x1.6db793c887b95p-255,
	-0x1.b5bfc17fa97d4p-267,
	0x1.e9e56d649f76bp-279,
	-0x1.00dcf6a320e1dp-290,
}

// Copied from listing_0184_sine_coefficients.inl
var SineRadiansC_MFTWP = [][]float64{
	{},
	{},
	{0x1.fc4eac57b4a27p-1, -0x1.2b704cf682899p-3},
	{0x1.fff1d21fa9dedp-1, -0x1.53e2e5c7dd831p-3, 0x1.f2438d36d9dbbp-8},
	{0x1.ffffe07d31fe8p-1, -0x1.554f800fc5ea1p-3, 0x1.105d44e6222ap-7, -0x1.83b9725dff6e8p-13},
	{0x1.ffffffd25a681p-1, -0x1.555547ef5150bp-3, 0x1.110e7b396c557p-7, -0x1.9f6445023f795p-13, 0x1.5d38b56aee7f1p-19},
	{0x1.ffffffffd17d1p-1, -0x1.55555541759fap-3, 0x1.11110b74adb14p-7, -0x1.a017a8fe15033p-13, 0x1.716ba4fe56f6ep-19, -0x1.9a0e192a4e2cbp-26},
	{0x1.ffffffffffdcep-1, -0x1.5555555540b9bp-3, 0x1.111111090f0bcp-7, -0x1.a019fce979937p-13, 0x1.71dce5ace58d2p-19, -0x1.ae00fd733fe8dp-26, 0x1.52ace959bd023p-33},
	{0x1.fffffffffffffp-1, -0x1.5555555555469p-3, 0x1.111111110941dp-7, -0x1.a01a0199e0eb3p-13, 0x1.71de37e62aacap-19, -0x1.ae634d22bb47cp-26, 0x1.60e59ae00e00cp-33, -0x1.9ef5d594b342p-41},
	{0x1p0, -0x1.5555555555555p-3, 0x1.11111111110c9p-7, -0x1.a01a01a014eb6p-13, 0x1.71de3a52aab96p-19, -0x1.ae6454d960ac4p-26, 0x1.6123ce513b09fp-33, -0x1.ae43dc9bf8ba7p-41, 0x1.883c1c5deffbep-49},
	{0x1p0, -0x1.5555555555555p-3, 0x1.11111111110dcp-7, -0x1.a01a01a016ef6p-13, 0x1.71de3a53fa85cp-19, -0x1.ae6455b871494p-26, 0x1.612421756f93fp-33, -0x1.ae671378c3d43p-41, 0x1.90277dafc8ab9p-49, -0x1.78262e1f2709cp-58},
	{0x1p0, -0x1.5555555555555p-3, 0x1.11111111110dp-7, -0x1.a01a01a01559ap-13, 0x1.71de3a52ad36dp-19, -0x1.ae64549aa7ca9p-26, 0x1.612392f66fdcdp-33, -0x1.ae11556cad6c4p-41, 0x1.71744c339ad03p-49, 0x1.52947c90f8199p-55, -0x1.ff1898c107cfap-59},
}

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

func SinTaylorFunc(fn func(float64, int) float64, n int) func(float64) float64 {
	return func(x float64) float64 {
		return fn(x, n)
	}
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

func SinTaylorPre(x float64, n int) float64 {
	x2 := x * x
	y := float64(0)
	for i := n; i > 0; i-- {
		y = math.FMA(y, x2, sinTaylorCoeffsArray[i-1])
	}
	y *= x
	return y
}

func SinMFTWP(x float64, n int) float64 {
	x2 := x * x
	y := float64(0)
	for i := n; i > 0; i-- {
		y = math.FMA(y, x2, SineRadiansC_MFTWP[n][i-1])
	}
	y *= x
	return y
}

func SinMFTWP_Manual9(x float64) float64 {
	x2 := x * x
	y := float64(0)

	y = math.FMA(y, x2, 0x1.883c1c5deffbep-49)
	y = math.FMA(y, x2, -0x1.ae43dc9bf8ba7p-41)
	y = math.FMA(y, x2, 0x1.6123ce513b09fp-33)
	y = math.FMA(y, x2, -0x1.ae6454d960ac4p-26)
	y = math.FMA(y, x2, 0x1.71de3a52aab96p-19)
	y = math.FMA(y, x2, -0x1.a01a01a014eb6p-13)
	y = math.FMA(y, x2, 0x1.11111111110c9p-7)
	y = math.FMA(y, x2, -0x1.5555555555555p-3)
	y = math.FMA(y, x2, 0x1p0)
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

func PrintSinTaylorCoeffs(n int) {
	fmt.Printf("var sinTaylorCoeffsArray = []float64{\n")
	for i := range n {
		fmt.Printf("	%x,\n", sinTaylorCoeff(i))
	}
	fmt.Printf("}\n")
}
