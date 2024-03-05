//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	gen4x2()
	//gen4x2_alt()
	gen8x2()
	gen16x2()
	gen32x2()
	Generate()
}

func gen4x2() {
	TEXT("Read_4x2_go", NOSPLIT, "func(repeatCount uint64, bb []byte)")

	repeatCount := Load(Param("repeatCount"), GP64())
	ptr := Load(Param("bb").Base(), GP64())
	rep := GP64()
	target := GP32()

	Label("loop")
	MOVL(Mem{Base: ptr}, target)
	MOVL(Mem{Base: ptr, Disp: 4}, target)
	ADDQ(Imm(8), rep)
	CMPQ(rep, repeatCount)
	JB(LabelRef("loop"))

	RET()
}

// Alternative implementation using specific pseudo registers.
func gen4x2_alt() {
	TEXT("Read_4x2_go", NOSPLIT, "func(repeatCount uint64, bb []byte)")

	repeatCount := Load(Param("repeatCount"), RDI)
	ptr := Load(Param("bb").Base(), RSI)
	XORQ(RAX, RAX)

	Label("loop")
	MOVL(Mem{Base: ptr}, R8L)
	MOVL(Mem{Base: ptr, Disp: 4}, R8L)
	ADDQ(Imm(8), RAX)
	CMPQ(RAX, repeatCount)
	JB(LabelRef("loop"))

	RET()
}

func gen8x2() {
	TEXT("Read_8x2_go", NOSPLIT, "func(repeatCount uint64, bb []byte)")

	repeatCount := Load(Param("repeatCount"), GP64())
	ptr := Load(Param("bb").Base(), GP64())
	rep := GP64()
	target := GP64()

	Label("loop")
	MOVQ(Mem{Base: ptr}, target)
	MOVQ(Mem{Base: ptr, Disp: 8}, target)
	ADDQ(Imm(16), rep)
	CMPQ(rep, repeatCount)
	JB(LabelRef("loop"))

	RET()
}

func gen16x2() {
	TEXT("Read_16x2_go", NOSPLIT, "func(repeatCount uint64, bb []byte)")

	repeatCount := Load(Param("repeatCount"), GP64())
	ptr := Load(Param("bb").Base(), GP64())
	rep := GP64()

	Label("loop")
	VMOVDQU(Mem{Base: ptr}, X0)
	VMOVDQU(Mem{Base: ptr, Disp: 16}, X1)
	ADDQ(Imm(32), rep)
	CMPQ(rep, repeatCount)
	JB(LabelRef("loop"))

	RET()
}

func gen32x2() {
	TEXT("Read_32x2_go", NOSPLIT, "func(repeatCount uint64, bb []byte)")

	repeatCount := Load(Param("repeatCount"), GP64())
	ptr := Load(Param("bb").Base(), GP64())
	rep := GP64()

	Label("loop")
	VMOVDQU(Mem{Base: ptr}, Y0)
	VMOVDQU(Mem{Base: ptr, Disp: 32}, Y1)
	ADDQ(Imm(64), rep)
	CMPQ(rep, repeatCount)
	JB(LabelRef("loop"))

	RET()
}
