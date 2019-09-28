//+build cgo

// High level bindings for https://github.com/ebiggers/libdeflate
package deflate

/*
#cgo CFLAGS: -Ilibdeflate -Ilibdeflate/lib -Ilibdeflate/common -O3 -fomit-frame-pointer
#include "lib_common.h"
#include "libdeflate/libdeflate.h"
#include "aligned_malloc.c"
#include "deflate_compress.c"
#undef BITBUF_NBITS
#include "deflate_decompress.c"
#include "zlib_compress.c"
#include "zlib_decompress.c"
#define dispatch dispatch__
#include "adler32.c"
#include "x86/cpu_features.c"
*/
import "C"
import "unsafe"

// Compress input[] at level 1-12, append() the result to dst[].
func Compress(dst, input []byte, level int, zlib bool) []byte {
	comp := C.libdeflate_alloc_compressor(C.int(level))
	inLen := C.size_t(len(input))
	outLen := int(C.libdeflate_zlib_compress_bound(comp, inLen))
	if dst == nil {
		dst = make([]byte, outLen+len(dst))[:0]
	} else if outLen > cap(dst)-len(dst) {
		newDst := make([]byte, outLen+len(dst))
		dst = newDst[:copy(newDst, dst)]
	}
	lendst := len(dst)
	dst = dst[:lendst+outLen]
	if zlib {
		outLen = int(C.libdeflate_zlib_compress(comp, unsafe.Pointer(&input[0]), inLen, unsafe.Pointer(&dst[lendst]), C.size_t(outLen)))
	} else {
		outLen = int(C.libdeflate_deflate_compress(comp, unsafe.Pointer(&input[0]), inLen, unsafe.Pointer(&dst[lendst]), C.size_t(outLen)))
	}
	C.libdeflate_free_compressor(comp)
	return dst[:lendst+outLen]
}

// Decompress zlib/deflate stream in input[], append() the result to dst[]
func Decompress(dst, input []byte, zlib bool) []byte {
	decomp := C.libdeflate_alloc_decompressor()
	inLen := C.size_t(len(input))
	if dst == nil {
		dst = make([]byte, 4096)[:0]
	}
	var result uint32
	for {
		var outLen C.size_t
		outCap := C.size_t(cap(dst) - len(dst))
		lenDst := len(dst)
		dst = dst[:cap(dst)]
		if zlib {
			result = C.libdeflate_zlib_decompress(decomp, unsafe.Pointer(&input[0]), inLen, unsafe.Pointer(&dst[lenDst]), outCap, &outLen)
		} else {
			result = C.libdeflate_deflate_decompress(decomp, unsafe.Pointer(&input[0]), inLen, unsafe.Pointer(&dst[lenDst]), outCap, &outLen)
		}
		if result == C.LIBDEFLATE_INSUFFICIENT_SPACE {
			newDst := make([]byte, int(outLen)+lenDst)
			dst = newDst[:copy(newDst, dst)]
			continue
		} else if result != C.LIBDEFLATE_SUCCESS {
			return nil
		}
		return dst[:lenDst+int(outLen)]
	}
}
