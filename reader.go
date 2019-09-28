package archive

import (
	"errors"
	"io"
	"os"
	"time"
)

// Header is returned per each file encountered in the stream, or in a List() if supported.
type Header struct {
	Name  string    // Path, including sub folders within the archive.
	Size  int64     // Size of the file, in bytes. If not known upfront, then -1.
	Time  time.Time // Timestamp, typically ModTime.
	Index int       // Index within the archive. For Seek(), if it works.
	IsDir bool      // If this is, in fact, a directory.
}

// The reader keeps the state of the archive stream.
// The interface is directly provided by the backend format.
// If you close a Reader, io.ErrClosedPipe will be thrown
// by all attempts to Read(), Next() or Seek() on it.
// Note that Close() never closes the actual input stream.
// It is here to clean up cgo memory used by decoders.
type Reader interface {
	io.ReadCloser
	Next() (*Header, error)
}

// If a cast to (reader).(SeekReader) succeeds, the underlying format
// and the input stream support seeking (gleaned by internal casts made on it).
type SeekReader interface {
	List() []Header               // Returns all headers, or nil if not supported.
	Seek(i int, whence int) error // Seek to index. whence same meaning as in io.Seeker.
}

// *Opt configuration.
type Options struct {
	Name        string // Used as file path by dir format, otherwise informational.
	Size        int64  // Size of input stream. Providing helps stream become seekable.
	MemLimitKiB int    // Memory bound for solid blocks. When overstepped, seeking is disabled.
	AllowDir    bool   // Allow opening of directories as "archive". Path in Name.
	WantSeek    bool   // Seeking is preferred, otherwise prefers non-seeking handler.
	SkipDirs    bool   // Next() will auto-skip over IsDir entries.
}

// Default options, if no other supplied.
var DefaultOptions = Options{
	MemLimitKiB: 512 * 1024,
	AllowDir:    true,
	WantSeek:    false,
	SkipDirs:    true,
}

// A format is an individual driver to handle a file type.
// There can be multiple handlers for same archive format.
// Whichever "wins" depends on Formats[] order, as well as Options
// and interface of the underlying input (if it can seek or not).
//
// A format will typically probe start of the stream. If the object
// turns out to be non-rewindable (interface cast to Peek, Seeker or
// ReaderAt fails), a failed probe must replace input pointer with
// buffered reader where they prepend the probed bytes, along with a
// Peek() interface method so only a single wrapper will be made.
type Format func(input *io.Reader, opt *Options) (Reader, error)

// List of formats. You may want to clear it and re-register
// formats in custom order.
var Formats []Format

// Create a new reader for auto-detected archive on r.
func NewReader(r io.Reader) (Reader, error) {
	return NewReaderOpt(r, nil)
}

// Create a new reader for auto-detected archive on r with options.
func NewReaderOpt(r io.Reader, opt *Options) (res Reader, err error) {
	o := checkOpt(opt)
	for _, f := range Formats {
		if res, err = f(&r, &o); res != nil || err != nil {
			break
		}
	}
	return
}

var ErrNoDirsAllowed = errors.New("no directories allowed")
var ErrFormatUnknown = errors.New("unknown format")

// Open an archive from a file path.
func Open(path string, opt *Options) (Reader, error) {
	st, err := os.Stat(path)
	o := checkOpt(opt)
	if err != nil {
		return nil, err
	}
	var f *os.File
	if st.IsDir() {
		// Will try dir formats
		if !o.AllowDir {
			return nil, ErrNoDirsAllowed
		}
		o.Name = path
	} else {
		// Otherwise file formats
		if f, err = os.Open(path); err != nil {
			return nil, err
		}
		o.Name = path
		o.Size = st.Size()
	}

	// CAVEAT: interface nil workaround, dir format wants it
	var ior io.Reader
	if f != nil {
		ior = f
	}

	r, err := NewReaderOpt(ior, &o)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrFormatUnknown
	}

	// Wrap the result in outer closer, so that Close() call will
	// nuke our file as well.
	return &fileCloser{r, f}, nil
}

type fileCloser struct {
	Reader
	f *os.File
}

// close the Open()ed file, as well the reader format.
func (fc *fileCloser) Close() error {
	err := fc.Reader.Close() // usually a nopcloser, but sometimes state cleanup.
	err2 := fc.f.Close()
	if err != nil {
		return err
	}
	return err2
}

func checkOpt(o *Options) Options {
	if o == nil {
		return DefaultOptions
	}
	return *o
}
