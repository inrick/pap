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
// extern void ReadSuccessiveSizesNonPow2(uint64_t count, uint8_t *data, uint64_t chunk_size);
// extern void ReadStrided_32x2(uint64_t count, uint8_t *data, uint64_t chunk_size, uint64_t stride);
// extern void WriteTemporal(uint8_t *input, uint8_t *output, uint64_t read_size, uint64_t inner_read_size);
// extern void WriteNonTemporal(uint8_t *input, uint8_t *output, uint64_t read_size, uint64_t inner_read_size);
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

func ReadSuccessiveSizesNonPow2(repeatCount uint64, bb []byte, chunkSize uint64) {
	C.ReadSuccessiveSizesNonPow2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)), C.uint64_t(chunkSize))
}

func ReadStrided_32x2(repeatCount uint64, bb []byte, chunkSize uint64, stride uint64) {
	C.ReadStrided_32x2(C.uint64_t(repeatCount), (*C.uint8_t)(unsafe.SliceData(bb)), C.uint64_t(chunkSize), C.uint64_t(stride))
}

func WriteTemporal(input, output []byte, readSize, innerReadSize uint64) {
	C.WriteTemporal((*C.uint8_t)(unsafe.SliceData(input)), (*C.uint8_t)(unsafe.SliceData(output)), C.uint64_t(readSize), C.uint64_t(innerReadSize))
}

func WriteNonTemporal(input, output []byte, readSize, innerReadSize uint64) {
	C.WriteNonTemporal((*C.uint8_t)(unsafe.SliceData(input)), (*C.uint8_t)(unsafe.SliceData(output)), C.uint64_t(readSize), C.uint64_t(innerReadSize))
}
