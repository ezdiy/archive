package rar

//#include <stdint.h>
import "C"
import (
	"io"
	"math"
	"unsafe"
)

//export readFunc
func readFunc(opaque, buffer unsafe.Pointer, n int) int {
	ar := (*Reader)(opaque)
	got, err := ar.Rs.Read((*[math.MaxInt32]byte)(buffer)[:n])
	if err != nil {
		return -1
	}
	return got
}

//export seekFunc
func seekFunc(opaque unsafe.Pointer, offset int64) int {
	ar := (*Reader)(opaque)
	_, _ = ar.Rs.Seek(offset, io.SeekStart)
	return 0
}

//export extractFunc
func extractFunc(opaque unsafe.Pointer, buffer *unsafe.Pointer, size *C.size_t, usize C.size_t, ret *int) bool {
	ar := (*Reader)(opaque)
	ar.ReadResp <- int(usize)
	ar.ReadPending = false
	req := <-ar.ReadReq
	if req == nil {
		return false
	}
	// for next read
	ar.ReadPending = true
	*buffer = unsafe.Pointer(&req[0])
	*size = C.size_t(len(req))
	return true
}
