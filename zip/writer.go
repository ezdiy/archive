package zip

/*
import "archive/zip"

import (
	"archive/zip"
	"github.com/ezdiy/archive/deflate"
	"io"
	"os"
	"time"
)

// Zip archive writer.
type Writer struct {
	zw *zip.Writer
	of io.WriteCloser

	buf []byte
	freeBuf []byte

	Level int			// Compression level, 0-12 (libdeflate)
	MinCompress int		// % by which the file must be shrunk in order to keep compression.
}

type writeOnClose struct {
	a *Writer
	w io.Writer
	buf []byte
}

func (w *writeOnClose) Write(b[]byte) (int,error) {
	return len(b), nil
}

func (w *writeOnClose) Close() error {
	wr, err := w.w.Write(w.buf)
	if wr != len(w.buf) {
		panic("short write")
	}
	// buffer gets recycled
	w.a.freeBuf = w.buf[:0]
	return err
}

// Write zip file to given output stream ownership of will be taken.
func NewCloseWriter(output io.WriteCloser) (a *Writer) {
	a = &Writer{of:output}
	a.zw = zip.NewWriter(a.of)
	a.zw.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) {
		return &writeOnClose{a,w, a.buf}, nil
	})
	return
}

type nopCloser struct {
	io.Writer
}
func (n *nopCloser) Close() error {
	return nil
}

// Write zip file to given output stream.
// The provided stream is kept unclosed after closing Writer.
func NewWriter(output io.Writer) *Writer {
	return NewCloseWriter(&nopCloser{output})
}

// Crate new zip file.
func CreateFile(file string) (*Writer, error) {
	of, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return NewCloseWriter(of), nil
}

// Flush the compressed data and close (depending on open type) the finished archive.
func (a *Writer) Close() {
	_ = a.zw.Close()
	_ = a.of.Close()
}

// Quick-add a file from buffer, timestamp is current time, default compression is used.
func (a *Writer) Add(name string, buffer []byte) (cs int) {
	return a.AddLevel(name,buffer,time.Now(),-1)
}

// Add a file name from buffer with timestamp t. Use compression level (0 = store method).
// Returns compressed size.
func (a *Writer) AddLevel(name string, buffer[]byte, t time.Time, level int) (cs int) {
	if level == -1 {
		level = a.Level
	}
	cs = len(buffer)
	if level > 0 {
		// Compress the data first
		a.buf = deflate.Compress(a.freeBuf, buffer, level, false)
		a.freeBuf = nil
		// Figure out if we should keep it compressed
		threshold := len(buffer) - (len(buffer) * a.MinCompress / 100)
		cs = len(a.buf)
		if cs >= len(buffer) || cs > threshold {
			level = 0
			cs = len(buffer)
		}
	}
	method := zip.Store
	if level > 0 {
		method = zip.Deflate
	}
	h := &zip.FileHeader{
		Name:name,
		Method:method,
		Modified:t,
	}
	fo, err := a.zw.CreateHeader(h)
	if err != nil {
		return -1
	}
	if level == 0 && a.buf != nil {
		a.freeBuf = a.buf[:0]
	}
	_, err = fo.Write(buffer)
	if err != nil {
		return -1
	}
	return
}

// Add a file name with timestamp, return its write stream.
func (a *Writer) AddWriter(name string, t time.Time, comp bool) (io.Writer, error) {
	method := zip.Deflate
	if !comp {
		method = zip.Store
	}
	h := &zip.FileHeader{
		Name:name,
		Method:method,
		Modified:t,
	}
	return a.zw.CreateHeader(h)
}

// Add a file, contents read from given source stream.
func (a *Writer) AddStream(name string, src io.Reader, t time.Time, comp bool) (int64, error) {
	fo, err := a.AddWriter(name, t, comp)
	if err != nil {
		return 0, err
	}
	return io.Copy(fo, src)
}
*/
