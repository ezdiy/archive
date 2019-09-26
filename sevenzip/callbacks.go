package sevenzip
//#include "7z.h"
import "C"
import (
	"math"
	"unsafe"
)

//export readCallback
func readCallback(pp unsafe.Pointer, pBuf unsafe.Pointer, size *int) int32 {
	p := (*Reader)(pp)
	buf := (*[math.MaxInt32]byte)(pBuf)
	got, err := p.RS.Read(buf[:*size])
	*size = got
	if err != nil {
		return -1
	}
	return 0
}

//export seekCallback
func seekCallback(pp unsafe.Pointer, pos *int64, origin int32) int32 {
	p := (*Reader)(pp)
	var err error
	*pos, err = p.RS.Seek(*pos, int(origin))
	if err != nil {
		return -1
	}
	return 0
}

