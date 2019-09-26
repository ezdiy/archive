package archive

/*
import (
	"io"
	"time"
)

type Header struct {
	Name	string				// Path, including subfolders within the archive.
	Size 	int64				// Size of the file, in bytes. If not known upfront, then -1.
	Time 	time.Time			// Timestamp, typically ModTime.
	Index 	int					// Index within the archive. For Seek(), if it works.
	IsDir	bool				// If this is, in fact, a directory.
}

// The reader keeps the state of the archive stream.
// It is directly provided by the backend format.
type Reader interface {
	io.Reader
	Next() (*Header, error)
}

type Options struct {

}

// A format is an individual driver to handle a file type.
// There can be multiple readers.
//
// A reader will typically probe start of the stream. If the object
// turns out to be non-rewindable (interface cast to Peek, Seeker or ReaderAt fails),
// the failed probe must replace input pointer with buffered reader where they
// prepend the probed bytes, along with a Peek() method.
type Format func(input *io.Reader, opt *Options) (Reader, error)


var Formats = []Format

// Create a new reader for auto-detected archive r.
func NewReader(r io.Reader) (Reader, error) {
	return nil, nil
}

// Create a new reader for auto-detected archive r.
// Reader pointer can be replaced due to failed probe.
func NewReaderOpt(r *io.Reader, opt *Options) (res Reader, err error) {
	for _, f := range Formats {
		if res, err = f(r, opt); res != nil || err != nil {
			break
		}
	}
	return
}





/*







// You can pass this as size argument for directory archive handlers.
const IsDirectory = -1234

// A file entry we're currently reading from the archive. Beware that
// there can be only one per archive at a time. If you Next(),
// and attempt to use stale File of old Next(), it will behave
// erratically or throw errors.
//
// The reason for this is that the underlying stream of all File can be
// one and the same (of the archive itself).
type File interface {
	io.ReadCloser
	FileInfo
}

// Information about one archived file.
type FileInfo interface {
	Name() string // Base name
	Path() string // Within archive
	Size() int64
	ModTime() time.Time
	IsDir() bool
}

var NotSeekable = errors.New("archive not seekable")

// Reader is the context of opened archive.
// Only Next() method is guaranteed to work.
// The other three may not work at all, or be very slow.
type Reader interface {
	io.Closer
	Next(skip int) (File, error)	// Return the current File, and skips n. io.EOF on archive end.
}

type SeekReader interface {
	List() []FileInfo 		// Returns info about all files in the archive.
	Seek(i int) error 		// Seek to file at index.
	Tell() int		  		// Tell where in archive we are.
}

type Options struct {
	io.Closer
	Size 		int64
	Path 		string
	MemLimitKb	int
	AllowDir	bool
	WantSeek	bool
}

var DefaultOptions = Options{
	AllowDir: true,
	Closer: ioutil.NopCloser(nil),
	WantSeek: false,
	MemLimitKb: 512 * 1024,
}

type Format func(input io.Reader, opt *Options) (Reader, error)

var LibFormats, Formats []Format

func Register(f Format) {
	Formats = append(Formats, f)
}

func LibRegister(f Format) {
	LibFormats = append(LibFormats, f)
}

func tryFormatsTab(tab []Format, input io.Reader, opt *Options) (r Reader, e error) {
	for _, f := range tab {
		r, e = f(input, opt)
		if r != nil || e != nil {
			return
		}
	}
	return
}

func tryFormats(input io.Reader, opt *Options) (r Reader, e error) {
	if r, e = tryFormatsTab(Formats, input, opt); r != nil || e != nil {
		return
	}
	return tryFormatsTab(LibFormats, input, opt)
}

// Open an input stream.
func Open(input io.Reader)(r Reader, e error) {
	o := DefaultOptions
	return OpenOpt(input, &o)
}

// Open an input stream with options.
func OpenOpt(input io.Reader, opt *Options) (r Reader, e error) {
	o := *opt
	return tryFormats(input, &o)
}

// Explicit ReadSeeker interface (size will be seek-determined)
func OpenSeekOpt(input io.ReadSeeker, opt *Options) (r Reader, e error) {
	o := *opt
	o.WantSeek = true
	return tryFormats(input.(io.Reader), &o)
}

func SeekGetSize(input io.Seeker) (ret int64) {
	ret, _ = input.Seek(0, io.SeekEnd)
	_,_ = input.Seek(0,io.SeekStart)
	return
}

// Explicitly ReaderAt interface (size pre-determined)
func OpenRatOpt(input io.ReaderAt, size int64, opt *Options) (r Reader, e error) {
	o := *opt
	o.WantSeek = true
	o.Size = size
	return tryFormats(input.(io.Reader), &o)
}

// Convert ReaderSeeker to ReaderAt. Note that this will make it racy.
type ReaderAtWrapper struct {
	io.ReadSeeker
}
func (w *ReaderAtWrapper) ReadAt(p []byte, off int64) (n int, err error) {
	save, _ := w.Seek(0, io.SeekCurrent)
	if _, err = w.Seek(off, io.SeekStart); err != nil {
		return
	}
	n, err = w.Read(p)
	_,_ = w.Seek(save, io.SeekStart)
	return
}

// Convert interface to ReaderAt, unless it is one already.
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

const CantGetSizeHack = math.MaxInt64/2

// Convert interface to ReadSeeker, if it isn't one already.
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

// Open an archive file (or directory) at a given path.
func OpenFile(path string) (Reader, error) {
	return OpenFileOpt(path, &DefaultOptions)
}

func OpenFileOpt(path string, opt *Options) (Reader,error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	o := *opt
	o.Path = path
	if o.AllowDir && st.IsDir() {
		return OpenOpt(nil, &o)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	o.Closer = f
	o.Size = st.Size()
	return OpenSeekOpt(f, &o)
}

// Generic helpers for Format handlers.
type GenericFileInfo struct {
	FSize	int64
	FTime 	time.Time
	FName 	string
	FIsDir	bool
	Index 	int
}

type GenericFileList struct {
	FList	[]FileInfo
	Pos 	int
}

func (f *GenericFileList) List() []FileInfo {
	return f.FList
}

func (f *GenericFileList) Next(skip int) (fi FileInfo) {
	for {
		if f.Pos >= len(f.FList) {
			return nil
		}
		if !f.FList[f.Pos].IsDir() {
			break
		}
		f.Pos++
	}
	fi = f.FList[f.Pos]
	f.Pos += skip
	return
}

func (a *GenericFileList) Tell() int {
	return a.Pos
}

func (a *GenericFileList) Seek(i int) error {
	if i >= len(a.FList) {
		return io.EOF
	}
	a.Pos = i
	return nil
}

func (f *GenericFileInfo) Name() string {
	return filepath.Base(f.FName)
}

func (f *GenericFileInfo) Path() string {
	return f.FName
}

func (f *GenericFileInfo) ModTime() time.Time {
	return f.FTime
}
func (f *GenericFileInfo) Size() int64 {
	return f.FSize
}

func (f *GenericFileInfo) IsDir() bool {
	return f.FIsDir
}

func CheckMagic(input io.Reader, magic string) bool {
	rs := MakeReaderAt(input, nil)
	if rs == nil {
		return false
	}
	buf := make([]byte, len(magic))
	if n, e := rs.ReadAt(buf, 0); e != nil || len(magic) != n {
		return false
	}
	return true
}

*/