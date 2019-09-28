package util

import (
	"bufio"
	"io"
)


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

