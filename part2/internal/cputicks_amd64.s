// See the cputicks function in go/src/runtime/asm_386.s to see how the Go
// runtime handles it. Unfortunately, it seems we cannot inline these
// instructions. There will always be a function call overhead in between.

#include "textflag.h"

// func Rdtsc() uint64
TEXT ·Rdtsc(SB),NOSPLIT,$0-8
	RDTSC
	MOVL	AX, ret_lo+0(FP)
	MOVL	DX, ret_hi+4(FP)
	RET

// func Rdtscp() uint64
TEXT ·Rdtscp(SB),NOSPLIT,$0-8
	RDTSCP
	MOVL	AX, ret_lo+0(FP)
	MOVL	DX, ret_hi+4(FP)
	RET

// func Rdtsc2() uint64
TEXT ·Rdtsc2(SB),NOSPLIT,$0-8
	RDTSC
	SHLQ	$32, DX
	ORQ	DX, AX
	MOVQ	AX, ret+0(FP)
	RET

// func Rdtscp2() uint64
TEXT ·Rdtscp2(SB),NOSPLIT,$0-8
	RDTSCP
	SHLQ	$32, DX
	ORQ	DX, AX
	MOVQ	AX, ret+0(FP)
	RET

// func CpuidFreqMhz() uint64
TEXT ·CpuidFreqMhz(SB),NOSPLIT,$0-8
	MOVQ	$0x16, AX
	XORQ	CX, CX
	CPUID
	MOVQ	AX, ret+0(FP)
	RET
