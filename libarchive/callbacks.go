package libarchive

/*
#include <archive.h>
extern void setError(struct archive *a, const char *msg);
 */
import "C"
import (
	"github.com/ezdiy/archive/util"
	"io"
	"unsafe"
)


//export seekFunc
func seekFunc(arc *C.struct_archive, p unsafe.Pointer, off int64, whence int32) int64 {
	a := (*Reader)(p)
	if whence == io.SeekEnd && a.Size == util.CantGetSizeHack {
		C.setError(arc, C.CString("Seeking to end, but don't know end."))
	}
	g, e := a.Seeker.Seek(off, int(whence))
	if e != nil {
		C.setError(arc, C.CString(e.Error()))
	}
	return g
}


//export skipFunc
func skipFunc(arc *C.struct_archive, p unsafe.Pointer, off int64) int64 {
	a := (*Reader)(p)
	if a.Seeker != nil {
		_, _ = a.Seeker.Seek(off, io.SeekCurrent) // TODO error
		return off
	}
	if off > ReadBufSize {
		off = ReadBufSize
	}
	got, _ := a.Reader.Read(a.ReadBuf[:off])
	return int64(got)
}

//export readFunc
func readFunc(arc *C.struct_archive, p unsafe.Pointer, buf **byte) int64 {
	a := (*Reader)(p)
	got, err := a.Reader.Read(a.ReadBuf[:])
	if got == 0 && err != nil {
		C.setError(arc, C.CString(err.Error()))
	}
	*buf = &a.ReadBuf[0]
	return int64(got)
}


