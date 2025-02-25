//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	genSqrt()
	genFma()
	Generate()
}

func genSqrt() {
	TEXT("SqrtAlt", NOSPLIT, "func(x float64) float64")
	x := Load(Param("x"), XMM())
	SQRTSD(x, x)
	Store(x, ReturnIndex(0))
	RET()
}

// This is a terrible idea due to all the calling overhead, just trying it out
// to see it's the right instruction.
func genFma() {
	TEXT("FMAAlt", NOSPLIT, "func(x, y, z float64) float64")
	x := Load(Param("x"), XMM())
	y := Load(Param("y"), XMM())
	z := Load(Param("z"), XMM())
	VFMADD231SD(x, y, z)
	Store(z, ReturnIndex(0))
	RET()
}
