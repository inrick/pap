package loops

// #cgo LDFLAGS: -L. -lloops
// #include <stdint.h>
// extern void MovAllBytes(uint64_t count, uint8_t *data);
// extern void NopAllBytes(uint64_t count);
// extern void CmpAllBytes(uint64_t count);
// extern void DecAllBytes(uint64_t count);
import "C"
import "unsafe"

func MovAllBytes(bb []byte) {
	C.MovAllBytes(C.uint64_t(len(bb)), (*C.uint8_t)(unsafe.SliceData(bb)))
}

func NopAllBytes(bb []byte) {
	C.NopAllBytes(C.uint64_t(len(bb)))
}

func CmpAllBytes(bb []byte) {
	C.CmpAllBytes(C.uint64_t(len(bb)))
}

func DecAllBytes(bb []byte) {
	C.DecAllBytes(C.uint64_t(len(bb)))
}
