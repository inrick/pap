package asm

// #cgo LDFLAGS: -L. -lloops
// #include <stdint.h>
// extern void MovAllBytes(uint64_t count, uint8_t *data);
// extern void Nop3x1AllBytes(uint64_t count);
// extern void CmpAllBytes(uint64_t count);
// extern void DecAllBytes(uint64_t count);
// extern void Nop1x3AllBytes(uint64_t count);
// extern void Nop1x9AllBytes(uint64_t count);
import "C"
import "unsafe"

func MovAllBytes(bb []byte) {
	C.MovAllBytes(C.uint64_t(len(bb)), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Nop3x1AllBytes(bb []byte) {
	C.Nop3x1AllBytes(C.uint64_t(len(bb)))
}

func CmpAllBytes(bb []byte) {
	C.CmpAllBytes(C.uint64_t(len(bb)))
}

func DecAllBytes(bb []byte) {
	C.DecAllBytes(C.uint64_t(len(bb)))
}

func Nop1x3AllBytes(bb []byte) {
	C.Nop1x3AllBytes(C.uint64_t(len(bb)))
}

func Nop1x9AllBytes(bb []byte) {
	C.Nop1x9AllBytes(C.uint64_t(len(bb)))
}
