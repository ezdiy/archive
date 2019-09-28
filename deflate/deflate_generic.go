//+build !cgo

package deflate

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"io"
)

type appendWriter struct {
	buf []byte
}

func (a *appendWriter) Write(b []byte) (int, error) {
	a.buf = append(a.buf, b...)
	return len(b), nil
}

// Compress input[] at level 1-12, append() the result to dst[].
func Compress(dst, input []byte, level int, z bool) []byte {
	if level > 9 {
		level = 9
	}
	aw := &appendWriter{dst}
	var w io.WriteCloser
	if z {
		w, _ = zlib.NewWriterLevel(aw, level)
	} else {
		w, _ = flate.NewWriter(aw, level)
	}
	_, _ = w.Write(input)
	_ = w.Close()
	return aw.buf
}

// Decompress zlib/deflate bytes in input[], append() the result to dst[]
func Decompress(dst, input []byte, z bool) []byte {
	ri := bytes.NewReader(input)
	var r io.Reader
	if z {
		r, _ = zlib.NewReader(ri)
	} else {
		r = flate.NewReader(ri)
	}
	buf := bytes.NewBuffer(dst)
	_, _ = buf.ReadFrom(r)
	return buf.Bytes()
}
