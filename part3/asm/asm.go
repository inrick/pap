package asm

// #cgo LDFLAGS: -L. -lloops
// #include <stdint.h>
// extern void MovAllBytes(uint64_t count, uint8_t *data);
// extern void Nop3x1AllBytes(uint64_t count);
// extern void CmpAllBytes(uint64_t count);
// extern void DecAllBytes(uint64_t count);
// extern void Nop1x3AllBytes(uint64_t count);
// extern void Nop1x9AllBytes(uint64_t count);
// extern void Read_x1(uint64_t count, uint8_t *data);
// extern void Read_x2(uint64_t count, uint8_t *data);
// extern void Read_x3(uint64_t count, uint8_t *data);
// extern void Read_x4(uint64_t count, uint8_t *data);
// extern void Write_x1(uint64_t count, uint8_t *data);
// extern void Write_x2(uint64_t count, uint8_t *data);
// extern void Write_x3(uint64_t count, uint8_t *data);
// extern void Write_x4(uint64_t count, uint8_t *data);
// extern void Read_4x2(uint64_t count, uint8_t *data);
// extern void Read_8x2(uint64_t count, uint8_t *data);
// extern void Read_16x2(uint64_t count, uint8_t *data);
// extern void Read_32x2(uint64_t count, uint8_t *data);
// extern void ReadSuccessiveSizes(uint64_t count, uint8_t *data, uint64_t mask);
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

func Read_x1(repeatCount uint64, bb []byte) {
	C.Read_x1(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Read_x2(repeatCount uint64, bb []byte) {
	C.Read_x2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Read_x3(repeatCount uint64, bb []byte) {
	C.Read_x3(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Read_x4(repeatCount uint64, bb []byte) {
	C.Read_x4(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Write_x1(repeatCount uint64, bb []byte) {
	C.Write_x1(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Write_x2(repeatCount uint64, bb []byte) {
	C.Write_x2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Write_x3(repeatCount uint64, bb []byte) {
	C.Write_x3(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Write_x4(repeatCount uint64, bb []byte) {
	C.Write_x4(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Read_4x2(repeatCount uint64, bb []byte) {
	C.Read_4x2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Read_8x2(repeatCount uint64, bb []byte) {
	C.Read_8x2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Read_16x2(repeatCount uint64, bb []byte) {
	C.Read_16x2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func Read_32x2(repeatCount uint64, bb []byte) {
	C.Read_32x2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func ReadSuccessiveSizes(repeatCount uint64, bb []byte, mask uint64) {
	C.ReadSuccessiveSizes(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)), C.uint64_t(mask))
}
