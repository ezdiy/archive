package util

import (
	"io"
	"math"
)

func SeekGetSize(input io.Seeker) (ret int64) {
	ret, _ = input.Seek(0, io.SeekEnd)
	_, _ = input.Seek(0, io.SeekStart)
	return
}

func MakeReaderAt(input io.Reader, optSize *int64) (rat io.ReaderAt) {
	rat = input.(io.ReaderAt)
	rs := input.(io.ReadSeeker)
	if rs != nil && optSize != nil && *optSize == 0 {
		*optSize = SeekGetSize(rs)
	}
	if rat != nil {
		return
	}
	if rs != nil {
		rat = &ReaderAtWrapper{rs}
	}
	return
}

func MakeReadSeeker(input io.Reader, size *int64) (rs io.ReadSeeker) {
	rs = input.(io.ReadSeeker)
	if rs != nil {
		if *size == 0 {
			*size = SeekGetSize(rs)
		}
		return
	}
	rat := input.(io.ReaderAt)
	if rat != nil {
		n := *size
		if n == 0 {
			// HACK can't determine size, but can still seek
			// TODO: check if libarchive is really ok with this
			n = CantGetSizeHack
		}
		return io.NewSectionReader(rat, 0, n)
	}
	return
}

const CantGetSizeHack = math.MaxInt64 / 2

type ReaderAtWrapper struct {
	io.ReadSeeker
}

func (w *ReaderAtWrapper) ReadAt(p []byte, off int64) (n int, err error) {
	save, _ := w.Seek(0, io.SeekCurrent)
	if _, err = w.Seek(off, io.SeekStart); err != nil {
		return
	}
	n, err = w.Read(p)
	_, _ = w.Seek(save, io.SeekStart)
	return
}

