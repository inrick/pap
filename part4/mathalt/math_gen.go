//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
)

func main() {
	genSqrt()
	Generate()
}

func genSqrt() {
	TEXT("SqrtAlt", NOSPLIT, "func(x float64) float64")
	x := Load(Param("x"), XMM())
	SQRTSD(x, x)
	Store(x, ReturnIndex(0))
	RET()
}
