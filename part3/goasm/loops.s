// Code generated by command: go run goasm_gen.go -out loops.s -stubs stub.go. DO NOT EDIT.

#include "textflag.h"

// func Read_4x2_go(repeatCount uint64, bb []byte)
TEXT ·Read_4x2_go(SB), NOSPLIT, $0-32
	MOVQ    repeatCount+0(FP), AX
	MOVQ    bb_base+8(FP), CX
	PCALIGN $0x40

loop:
	MOVL (CX), BX
	MOVL 4(CX), BX
	ADDQ $0x08, DX
	CMPQ DX, AX
	JB   loop
	RET

// func Read_8x2_go(repeatCount uint64, bb []byte)
TEXT ·Read_8x2_go(SB), NOSPLIT, $0-32
	MOVQ    repeatCount+0(FP), AX
	MOVQ    bb_base+8(FP), CX
	PCALIGN $0x40

loop:
	MOVQ (CX), BX
	MOVQ 8(CX), BX
	ADDQ $0x10, DX
	CMPQ DX, AX
	JB   loop
	RET

// func Read_16x2_go(repeatCount uint64, bb []byte)
// Requires: AVX
TEXT ·Read_16x2_go(SB), NOSPLIT, $0-32
	MOVQ    repeatCount+0(FP), AX
	MOVQ    bb_base+8(FP), CX
	PCALIGN $0x40

loop:
	VMOVDQU (CX), X0
	VMOVDQU 16(CX), X1
	ADDQ    $0x20, DX
	CMPQ    DX, AX
	JB      loop
	RET

// func Read_32x2_go(repeatCount uint64, bb []byte)
// Requires: AVX
TEXT ·Read_32x2_go(SB), NOSPLIT, $0-32
	MOVQ    repeatCount+0(FP), AX
	MOVQ    bb_base+8(FP), CX
	PCALIGN $0x40

loop:
	VMOVDQU (CX), Y0
	VMOVDQU 32(CX), Y1
	ADDQ    $0x40, DX
	CMPQ    DX, AX
	JB      loop
	RET

// func ReadSuccessiveSizes_go(repeatCount uint64, bb []byte, offsetMask uint64)
// Requires: AVX
TEXT ·ReadSuccessiveSizes_go(SB), NOSPLIT, $0-40
	MOVQ    repeatCount+0(FP), AX
	MOVQ    bb_base+8(FP), CX
	MOVQ    offsetMask+32(FP), DX
	XORQ    DI, DI
	XORQ    BX, BX
	PCALIGN $0x40

loop:
	MOVQ    CX, SI
	ADDQ    BX, SI
	VMOVDQU (SI), Y0
	VMOVDQU 32(SI), Y0
	VMOVDQU 64(SI), Y0
	VMOVDQU 96(SI), Y0
	VMOVDQU 128(SI), Y0
	VMOVDQU 160(SI), Y0
	VMOVDQU 192(SI), Y0
	VMOVDQU 224(SI), Y0
	ADDQ    $0x00000100, DI
	MOVQ    DI, BX
	ANDQ    DX, BX
	CMPQ    DI, AX
	JB      loop
	RET

// func ReadSuccessiveSizesNonPow2_go(repeatCount uint64, bb []byte, chunkSize uint64)
// Requires: AVX
TEXT ·ReadSuccessiveSizesNonPow2_go(SB), NOSPLIT, $0-40
	MOVQ    repeatCount+0(FP), AX
	MOVQ    bb_base+8(FP), CX
	MOVQ    chunkSize+32(FP), DX
	XORQ    DI, DI
	PCALIGN $0x40

loop:
	XORQ SI, SI

inner:
	MOVQ    CX, BX
	ADDQ    SI, BX
	VMOVDQU (BX), Y0
	VMOVDQU 32(BX), Y0
	VMOVDQU 64(BX), Y0
	VMOVDQU 96(BX), Y0
	VMOVDQU 128(BX), Y0
	VMOVDQU 160(BX), Y0
	VMOVDQU 192(BX), Y0
	VMOVDQU 224(BX), Y0
	ADDQ    $0x00000100, SI
	CMPQ    SI, DX
	JB      inner
	ADDQ    SI, DI
	CMPQ    DI, AX
	JB      loop
	RET

// func ReadStrided_32x2_go(repeatCount uint64, bb []byte, chunkSize uint64, stride uint64)
// Requires: AVX
TEXT ·ReadStrided_32x2_go(SB), NOSPLIT, $0-48
	MOVQ    repeatCount+0(FP), AX
	MOVQ    bb_base+8(FP), CX
	MOVQ    chunkSize+32(FP), DX
	MOVQ    stride+40(FP), BX
	PCALIGN $0x40

loop:
	MOVQ DX, DI
	MOVQ CX, SI

inner:
	VMOVDQU (SI), Y0
	VMOVDQU 32(SI), Y0
	ADDQ    BX, SI
	SUBQ    $0x40, DI
	JNZ     inner
	SUBQ    DX, AX
	JNZ     loop
	RET

// func WriteTemporal_go(input []byte, output []byte, readSize uint64, innerReadSize uint64)
// Requires: AVX
TEXT ·WriteTemporal_go(SB), NOSPLIT, $0-64
	MOVQ    input_base+0(FP), AX
	MOVQ    output_base+24(FP), CX
	MOVQ    readSize+48(FP), DX
	MOVQ    innerReadSize+56(FP), BX
	MOVQ    AX, SI
	ADDQ    BX, SI
	MOVQ    CX, BX
	ADDQ    DX, BX
	PCALIGN $0x40

loop:
	MOVQ AX, DX

inner:
	VMOVDQU (DX), Y0
	VMOVDQU Y0, (CX)
	VMOVDQU 32(DX), Y0
	VMOVDQU Y0, 32(CX)
	VMOVDQU 64(DX), Y0
	VMOVDQU Y0, 64(CX)
	VMOVDQU 96(DX), Y0
	VMOVDQU Y0, 96(CX)
	VMOVDQU 128(DX), Y0
	VMOVDQU Y0, 128(CX)
	VMOVDQU 160(DX), Y0
	VMOVDQU Y0, 160(CX)
	VMOVDQU 192(DX), Y0
	VMOVDQU Y0, 192(CX)
	VMOVDQU 224(DX), Y0
	VMOVDQU Y0, 224(CX)
	ADDQ    $0x00000100, DX
	ADDQ    $0x00000100, CX
	CMPQ    DX, SI
	JB      inner
	CMPQ    CX, BX
	JB      loop
	RET

// func WriteNonTemporal_go(input []byte, output []byte, readSize uint64, innerReadSize uint64)
// Requires: AVX
TEXT ·WriteNonTemporal_go(SB), NOSPLIT, $0-64
	MOVQ    input_base+0(FP), AX
	MOVQ    output_base+24(FP), CX
	MOVQ    readSize+48(FP), DX
	MOVQ    innerReadSize+56(FP), BX
	MOVQ    AX, SI
	ADDQ    BX, SI
	MOVQ    CX, BX
	ADDQ    DX, BX
	PCALIGN $0x40

loop:
	MOVQ AX, DX

inner:
	VMOVDQU  (DX), Y0
	VMOVNTDQ Y0, (CX)
	VMOVDQU  32(DX), Y0
	VMOVNTDQ Y0, 32(CX)
	VMOVDQU  64(DX), Y0
	VMOVNTDQ Y0, 64(CX)
	VMOVDQU  96(DX), Y0
	VMOVNTDQ Y0, 96(CX)
	VMOVDQU  128(DX), Y0
	VMOVNTDQ Y0, 128(CX)
	VMOVDQU  160(DX), Y0
	VMOVNTDQ Y0, 160(CX)
	VMOVDQU  192(DX), Y0
	VMOVNTDQ Y0, 192(CX)
	VMOVDQU  224(DX), Y0
	VMOVNTDQ Y0, 224(CX)
	ADDQ     $0x00000100, DX
	ADDQ     $0x00000100, CX
	CMPQ     DX, SI
	JB       inner
	CMPQ     CX, BX
	JB       loop
	RET
