// Code generated by command: go run math_gen.go -out math.s -stubs stubs.go. DO NOT EDIT.

#include "textflag.h"

// func SqrtAlt(x float64) float64
// Requires: SSE2
TEXT ·SqrtAlt(SB), NOSPLIT, $0-16
	MOVSD  x+0(FP), X0
	SQRTSD X0, X0
	MOVSD  X0, ret+8(FP)
	RET
