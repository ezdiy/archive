package rar

import (
	"errors"
	"github.com/ezdiy/archive"
	"github.com/ezdiy/archive/util"
	"io"
	"math"
	"time"
	"unsafe"
)

/*
#cgo CFLAGS: -O6 -fomit-frame-pointer
// Keep the read buffer kinda big, to lower the cgo cost of re-entry
#define DMC_UNRAR_BS_BUFFER_SIZE 131072
#include "dmc_unrar/dmc_unrar.c"
extern dmc_unrar_read_func readFunc;
extern dmc_unrar_seek_func seekFunc;
extern dmc_unrar_extract_callback_func extractFunc;
*/
import "C"

type Reader struct {
	p [0]byte
	archive.FileList

	Rs          io.ReadSeeker
	Arc         *C.dmc_unrar_archive
	ReadPos     int
	ReadReq     chan []byte
	ReadResp    chan int
	ReadPending bool
}

func (r *Reader) Next() (*archive.Header, error) {
	if r.Advance() {
		return nil, io.EOF
	}
	r.ReadPos = r.CurFile.Index
	r.ReadReq <- nil
	return r.Current()
}

func (r *Reader) Read(b []byte) (int, error) {
	if r.FPos == -1 || r.FList == nil {
		return 0, io.EOF
	}
	r.ReadReq <- b
	got := <-r.ReadResp
	if got == -1 {
		return 0, io.EOF
	}
	if got < 0 {
		return 0, errors.New(C.GoString(C.dmc_unrar_strerror(C.dmc_unrar_return(-got))))
	}
	return got, nil
}

func (r *Reader) Close() error {
	r.FPos = -1
	r.FList = nil
	r.ReadReq <- nil
	ok := <-r.ReadResp
	if ok != math.MinInt32 {
		panic("unable to terminate read worker")
	}
	C.dmc_unrar_archive_close(r.Arc)
	C.free(unsafe.Pointer(r.Arc))
	r.Arc = nil
	return nil
}

func (r *Reader) readWorker() {
	for r.FPos != -1 {
		req := <-r.ReadReq
		if req == nil {
			continue
		}
		var got C.size_t
		r.ReadPending = true
		err := C.dmc_unrar_extract_file_with_callback(r.Arc, C.size_t(r.ReadPos), unsafe.Pointer(&req[0]), C.size_t(len(req)), &got, false,
			unsafe.Pointer(&r.p), C.dmc_unrar_extract_callback_func(unsafe.Pointer(&C.extractFunc)))
		if err != 0 {
			if !r.ReadPending {
				panic("readPending invariant")
			}
			r.ReadResp <- -int(err)
		} else if r.ReadPending {
			// there's a pending reader, but it never got anything
			r.ReadResp <- -1
		}
	}
	// notify Close() we've exited.
	r.ReadResp <- math.MinInt32
}

func Open(input *io.Reader, opt *archive.Options) (ret archive.Reader, e error) {
	if ok, err := util.CheckMagic(input, "Rar!"); err != nil || !ok {
		return nil, err
	}
	a := &Reader{}
	a.SkipDirs = opt.SkipDirs
	a.Rs = util.MakeReadSeeker(*input, &opt.Size)
	if a.Rs == nil || opt.Size == 0 {
		return nil, nil
	}

	a.Arc = (*C.dmc_unrar_archive)(C.malloc(C.sizeof_dmc_unrar_archive))
	C.dmc_unrar_archive_init(a.Arc)

	a.Arc.io.func_read = C.dmc_unrar_read_func(unsafe.Pointer(&C.readFunc))
	a.Arc.io.func_seek = C.dmc_unrar_seek_func(unsafe.Pointer(&C.seekFunc))
	a.Arc.io.opaque = unsafe.Pointer(a)

	C.dmc_unrar_archive_open(a.Arc, C.uint64_t(opt.Size))
	s := a.Arc.internal_state
	nf := int(s.file_count)
	for i := 0; i < nf; i++ {
		fi := (*[math.MaxInt32]C.dmc_unrar_file_block)(unsafe.Pointer(s.files))
		var nameBuf [1024]byte
		got := int(C.dmc_unrar_get_filename(a.Arc, C.size_t(i), (*C.char)(unsafe.Pointer(&nameBuf[0])), C.size_t(len(nameBuf))))
		name := string(nameBuf[:got]) // can be still empty on error?
		isDir := bool(C.dmc_unrar_file_is_directory(a.Arc, C.size_t(i)))
		a.FList = append(a.FList, archive.Header{
			Size:  int64(fi[i].file.uncompressed_size),
			Time:  time.Unix(int64(fi[i].file.unix_time), 0),
			Name:  name,
			IsDir: isDir,
			Index: i,
		})
	}
	a.ReadReq = make(chan []byte)
	a.ReadResp = make(chan int)
	go a.readWorker()
	return a, e

}

func init() {
	archive.Formats = append(archive.Formats, Open)
}
