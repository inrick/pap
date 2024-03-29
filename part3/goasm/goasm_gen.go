//go:build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/ir"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func main() {
	gen4x2()
	//gen4x2_alt()
	gen8x2()
	gen16x2()
	gen32x2()
	genCacheFinder()
	genCacheFinderNonPow2()
	//genCacheFinder_alt()
	Generate()
}

// Avo currently does not implement the PCALIGN pseudo instruction but we can
// emit it literally like this.
func pcAlign(align uint8) {
	Instruction(&ir.Instruction{
		Opcode:   "PCALIGN",
		Operands: []Op{U8(align)},
	})
}

func gen4x2() {
	TEXT("Read_4x2_go", NOSPLIT, "func(repeatCount uint64, bb []byte)")

	repeatCount := Load(Param("repeatCount"), GP64())
	ptr := Load(Param("bb").Base(), GP64())
	rep := GP64()
	target := GP32()
	pcAlign(64)

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
	pcAlign(64)

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
	pcAlign(64)

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
	pcAlign(64)

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
	pcAlign(64)

	Label("loop")
	VMOVDQU(Mem{Base: ptr}, Y0)
	VMOVDQU(Mem{Base: ptr, Disp: 32}, Y1)
	ADDQ(Imm(64), rep)
	CMPQ(rep, repeatCount)
	JB(LabelRef("loop"))

	RET()
}

func genCacheFinder() {
	TEXT(
		"ReadSuccessiveSizes_go",
		NOSPLIT,
		"func(repeatCount uint64, bb []byte, offsetMask uint64)",
	)

	repeatCount := Load(Param("repeatCount"), GP64())
	ptr := Load(Param("bb").Base(), GP64())
	mask := Load(Param("offsetMask"), GP64())
	offset := GP64()
	offsetPtr := GP64()
	count := GP64()
	XORQ(count, count)
	XORQ(offset, offset)
	pcAlign(64)

	Label("loop")
	MOVQ(ptr, offsetPtr)
	ADDQ(offset, offsetPtr)
	const unroll = 8
	for i := 0; i < unroll; i++ {
		VMOVDQU(Mem{Base: offsetPtr, Disp: 32 * i}, Y0)
	}
	ADDQ(U32(unroll*32), count)
	MOVQ(count, offset)
	ANDQ(mask, offset)
	CMPQ(count, repeatCount)
	JB(LabelRef("loop"))

	RET()
}

func genCacheFinder_alt() {
	TEXT(
		"ReadSuccessiveSizes_go",
		NOSPLIT,
		"func(repeatCount uint64, bb []byte, offsetMask uint64)",
	)

	repeatCount := Load(Param("repeatCount"), RDI)
	ptr := Load(Param("bb").Base(), RSI)
	mask := Load(Param("offsetMask"), RDX)
	XORQ(RAX, RAX)
	XORQ(R8, R8)
	pcAlign(64)

	Label("loop")
	MOVQ(ptr, R9)
	ADDQ(R8, R9)
	const unroll = 8
	for i := 0; i < unroll; i++ {
		VMOVDQU(Mem{Base: R9, Disp: 32 * i}, Y0)
	}
	ADDQ(U32(unroll*32), RAX)
	MOVQ(RAX, R8)
	ANDQ(mask, R8)
	CMPQ(RAX, repeatCount)
	JB(LabelRef("loop"))

	RET()
}

func genCacheFinderNonPow2() {
	TEXT(
		"ReadSuccessiveSizesNonPow2_go",
		NOSPLIT,
		"func(repeatCount uint64, bb []byte, chunkSize uint64)",
	)

	repeatCount := Load(Param("repeatCount"), GP64())
	basePtr := Load(Param("bb").Base(), GP64())
	chunkSize := Load(Param("chunkSize"), GP64())

	loopPtr := GP64()    // Current pointer in inner loop
	loopOffset := GP64() // Current offset in inner loop
	totalCount := GP64() // Total count processed

	XORQ(totalCount, totalCount)

	pcAlign(64)

	Label("loop")
	XORQ(loopOffset, loopOffset)

	Label("inner")
	MOVQ(basePtr, loopPtr)
	ADDQ(loopOffset, loopPtr)
	const unroll = 8
	for i := 0; i < unroll; i++ {
		VMOVDQU(Mem{Base: loopPtr, Disp: 32 * i}, Y0)
	}

	ADDQ(U32(unroll*32), loopOffset)
	CMPQ(loopOffset, chunkSize)
	JB(LabelRef("inner"))

	ADDQ(loopOffset, totalCount)
	CMPQ(totalCount, repeatCount)
	JB(LabelRef("loop"))

	RET()
}
