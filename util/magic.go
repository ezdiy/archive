package util

import (
	"bufio"
	"io"
	"io/ioutil"
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

type Peeker interface {
	Peek(n int) ([]byte, error)
}

func CheckMagic(input *io.Reader, magic string) (bool, error) {
	rs := MakeReaderAt(*input, nil)
	var buf []byte
	var err error
	if rs == nil {
		// Can't seek. Check or create bufio.
		peeker := (*input).(Peeker)
		if peeker == nil {
			bio := bufio.NewReader(*input)
			peeker = bio
			*input = bio
		}
		buf, err = peeker.Peek(len(magic))
		if err != nil {
			return false, err
		}
	} else {
		buf = make([]byte, len(magic))
		// can seek
		n, err := rs.ReadAt(buf, 0)
		if err != nil || n < len(buf) {
			return false, err
		}
	}
	return string(buf) == magic, nil
}

type RCDelegate struct {
	io.Reader
	io.Closer
}

func (rc *RCDelegate) Update() {
	rc.Closer.Close()
	rc.Closer = ioutil.NopCloser(nil)

}
