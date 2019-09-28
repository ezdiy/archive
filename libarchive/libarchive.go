package libarchive

/*
#cgo windows AND amd64 LDFLAGS: -lbcrypt
#cgo !windows OR !amd64 pkg-config: libarchive
#include <archive.h>
#include <archive_entry.h>
#include <locale.h>

extern int64_t seekFunc(struct archive *, void *, int64_t, int);
extern int64_t readFunc(struct archive *, void *, const void **);
extern int64_t skipFunc(struct archive *, void *, int64_t);

void setError(struct archive *a, const char *msg) {
	archive_set_error(a, -1, msg);
}

*/
import "C"
import (
	"errors"
	"fmt"
	"github.com/ezdiy/archive"
	"github.com/ezdiy/archive/util"
	"io"
	"math"
	"time"
	"unicode/utf16"
	"unsafe"
)

const ReadBufSize = 1 << 18

type Reader struct {
	ReadBuf [ReadBufSize]byte

	Arc      *C.struct_archive
	Seeker   io.Seeker
	Reader   io.Reader
	Entry    *C.struct_archive_entry
	Pos      int
	SkipDirs bool

	Size int64
}

const ifdir = 0o0040000

func (entry *C.struct_archive_entry) getName(index int) string {
	// Works for RAR (most of the time)
	if p := C.archive_entry_pathname_utf8(entry); p != nil {
		return C.GoString(p)
	}
	// Works for 7z/zip (most of the time)
	if p := C.archive_entry_pathname_w(entry); p != nil {
		w := (*[math.MaxInt32]uint16)(unsafe.Pointer(p))
		i := 0
		for w[i] != 0 {
			i++
		}
		return string(utf16.Decode(w[:i]))
	}
	// Last resort ASCII
	if p := C.archive_entry_pathname(entry); p != nil {
		return C.GoString(p)
	}
	// Want the gory details? Make a cup of tea, for an once upon a time...
	// ... github.com/mpv-player/mpv/commit/1e70e82baa9193f6f027338b0fab0f5078971fbe
	return fmt.Sprintf("(no name for file entry #%d due to broken libarchive locale handling)", index)
}

func (r *Reader) Next() (*archive.Header, error) {
	if r.Arc == nil {
		return nil, io.ErrClosedPipe
	}
skipDir:
	ret := C.archive_read_next_header(r.Arc, &r.Entry)
	if ret == 1 {
		return nil, io.EOF
	} else if ret != 0 {
		return nil, r.Arc.getErr()
	}
	// ok got entry, fill it in
	ent := &archive.Header{
		Name:  r.Entry.getName(r.Pos),
		Size:  int64(C.archive_entry_size(r.Entry)),
		Time:  time.Unix(int64(C.archive_entry_mtime(r.Entry)), 0),
		Index: r.Pos,
		IsDir: (C.archive_entry_filetype(r.Entry) & ifdir) != 0,
	}
	r.Pos++
	if r.SkipDirs && ent.IsDir {
		goto skipDir
	}
	return ent, nil
}

func (r *Reader) Close() error {
	if r.Arc == nil {
		return io.ErrClosedPipe
	}
	C.archive_read_free(r.Arc)
	r.Arc = nil
	return nil
}

func (r *Reader) Read(b []byte) (int, error) {
	if r.Arc == nil {
		return 0, io.ErrClosedPipe
	}
	got := C.archive_read_data(r.Arc, unsafe.Pointer(&b[0]), C.size_t(len(b)))
	if got == 0 {
		return 0, io.EOF
	}
	if got < 0 {
		return 0, r.Arc.getErr()
	}
	return int(got), nil
}

func (a *C.struct_archive) getErr() error {
	return errors.New(C.GoString(C.archive_error_string(a)))
}

func Open(input *io.Reader, opt *archive.Options) (reader archive.Reader, e error) {
	// Setup
	arc := C.archive_read_new()
	C.archive_read_support_filter_all(arc)
	C.archive_read_support_format_all(arc)
	r := &Reader{
		Arc:      arc,
		Reader:   *input,
		SkipDirs: opt.SkipDirs,
	}

	// Add seeker if we can have one
	if seeker := util.MakeReadSeeker(*input, &opt.Size); seeker != nil && opt.Size != 0 {
		r.Seeker = seeker
		C.archive_read_set_seek_callback(arc, (*C.archive_seek_callback)(unsafe.Pointer(C.seekFunc)))
	}
	r.Size = opt.Size

	// Set callbacks
	C.archive_read_set_skip_callback(arc, (*C.archive_skip_callback)(unsafe.Pointer(C.skipFunc)))
	C.archive_read_set_read_callback(arc, (*C.archive_read_callback)(unsafe.Pointer(C.readFunc)))
	C.archive_read_set_callback_data(arc, unsafe.Pointer(&r.ReadBuf[0]))

	ret := C.archive_read_open1(arc)
	if ret != 0 {
		if r.Seeker == nil {
			return nil, arc.getErr()
		}
		// if there's a seeker, other handler can pick up
		return nil, nil
	}

	return r, nil
}

func init() {
	archive.Formats = append(archive.Formats, Open)
}
