package sevenzip

/*
#cgo CFLAGS: -I7z/C -O0
#include "7z.h"
extern SRes readCallback(const ISeekInStream *p, void *buf, size_t *size);
extern SRes seekCallback(const ISeekInStream *p, Int64 *pos, ESzSeek origin);
#include "Delta.c"
#include "7zBuf.c"
#include "7zArcIn.c"
#include "CpuArch.c"
#include "7zCrc.c"
#include "7zCrcOpt.c"
#include "7zDec.c"
#include "7zStream.c"
#include "Bra.c"
#include "Bra86.c"
#include "BraIA64.c"
#include "LzmaDec.c"
#include "Lzma2Dec.c"
#include "Alloc.c"
#undef kTopValue
#undef kBitModelTotal
#include "Bcj2.c"
 */
import "C"
import (
	"github.com/ezdiy/archive"
	"github.com/ezdiy/archive/util"
	"github.com/getlantern/errors"
	"io"
	"math"
	"time"
	"unicode/utf16"
	"unsafe"
)

const lookSize = 131072

type Reader struct {
	ISeek   C.ISeekInStream //must be first
	ILook   C.CLookToRead2
	LookBuf [lookSize]byte
	SZData  C.CSzArEx
	RS      io.ReadSeeker
	archive.FileList

	BlockIndex    C.UInt32
	OutBuffer     *C.Byte
	OutBufferSize C.size_t

	Cache []byte
}

func (r *Reader) Next() (*archive.Header, error) {
	if r.Advance() {
		return nil, io.EOF
	}
	r.Cache = nil
	return r.Current()
}

func (r *Reader) Read(b []byte) (n int, err error) {
	if r.RS == nil {
		return 0, io.ErrClosedPipe
	}
	var offset C.size_t
	var proc C.size_t
	if r.Cache == nil {
		ret := C.SzArEx_Extract(&r.SZData, &r.ILook.vt, C.UInt32(r.CurFile.Index),
			&r.BlockIndex, &r.OutBuffer, &r.OutBufferSize, &offset, &proc, &C.g_Alloc, &C.g_Alloc)
		if ret == 0 && proc == 0 { // empty file
			return 0, io.EOF
		}
		if ret != 0 {
			return 0, errors.New("data error") // TODO: translate errors
		}
		r.Cache = (*[math.MaxInt32]byte)(unsafe.Pointer(r.OutBuffer))[offset:][:r.CurFile.Size]
	}
	n = copy(b, r.Cache)
	if n == 0 {
		return 0, io.EOF
	}
	r.Cache = r.Cache[n:]
	return
}

func (r *Reader) Close() error {
	r.FList = nil
	r.FPos = -1
	if r.OutBuffer != nil {
		C.free(unsafe.Pointer(r.OutBuffer))
	}
	C.SzArEx_Free(&r.SZData, &C.g_Alloc)
	r.RS = nil
	r.Cache = nil
	return nil
}

func Open(input *io.Reader, opt *archive.Options) (reader archive.Reader, e error) {
	if ok, err := util.CheckMagic(input, "7z"); err != nil || !ok {
		return nil, err
	}
	r := &Reader{}
	r.SkipDirs = opt.SkipDirs
	r.RS = util.MakeReadSeeker(*input, &opt.Size)
	if r.RS == nil || opt.Size == 0 {
		return nil, nil
	}

	r.ISeek.Read = (*[0]byte)(C.readCallback)
	r.ISeek.Seek = (*[0]byte)(C.seekCallback)
	C.LookToRead2_CreateVTable(&r.ILook, 0)
	r.ILook.buf = (*C.uchar)(&r.LookBuf[0])
	r.ILook.bufSize = lookSize
	r.ILook.realStream = &r.ISeek
	s := &r.SZData
	C.SzArEx_Init(&r.SZData)

	res := C.SzArEx_Open(&r.SZData, &r.ILook.vt, &C.g_Alloc, &C.g_Alloc)
	if res != 0 {
		return nil, io.EOF // TODO translate errors
	}
	nameOffsets := (*[math.MaxInt32]int)(unsafe.Pointer(s.FileNameOffsets))
	names := (*[math.MaxInt32]uint16)(unsafe.Pointer(s.FileNames))
	offsets := (*[math.MaxInt32]uint64)(unsafe.Pointer(s.UnpackPositions))
	isDirs := (*[math.MaxInt32]byte)(unsafe.Pointer(s.IsDirs))
	tv := unsafe.Pointer(s.MTime.Vals)
	if tv == nil {
		tv = unsafe.Pointer(s.MTime.Vals)
	}
	mtime := (*[math.MaxInt32]uint64)(tv)
	for i := 0; i < int(s.NumFiles); i++ {
		ent := archive.Header{
			Name:string(utf16.Decode(names[nameOffsets[i]: nameOffsets[i+1]-1])),
			IsDir:(isDirs[i/8]&byte(1<<uint(7-(i&7))))!=0,
			Size:int64(offsets[i+1]- offsets[i]),
			Index:i,
		}
		if mtime != nil {
			ent.Time = time.Unix(int64(mtime[i]/uint64(10000000))-11644473600, 0)
		}
		r.FList = append(r.FList, ent)
	}
	return r, nil
}

func init() {
	C.CrcGenerateTable()
	archive.Formats = append(archive.Formats, Open)
}
