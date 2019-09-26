package archive

import (
	"io"
	"io/ioutil"
)

// FileList is a generic helper producing interface methods
// for seeking formats that keep track of all files.
type FileList struct {
	io.Reader
	io.Closer

	FList    []Header
	FPos     int
	CurFile  *Header
	SkipDirs bool
}

// Switch reader/closer for a new pair.
func (f *FileList) Switch(rc io.ReadCloser, err error) (*Header, error) {
	_ = f.Closer.Close()
	f.Closer = ioutil.NopCloser(nil)
	if err != nil {
		return nil, err
	}
	f.Reader, f.Closer = rc, rc
	return f.Current()
}

func (f *FileList) List() []Header {
	return f.FList
}

// Go to next file. Used by Next().
func (f *FileList) Advance() (eof bool) {
	// Something wished for all Next() to be nuked
	if f.FPos == -1 || f.FList == nil {
		return true
	}
	for {
		if f.FPos >= len(f.FList) {
			return true
		}
		if !f.SkipDirs || !f.FList[f.FPos].IsDir {
			break
		}
		f.FPos++
	}
	f.CurFile = &f.FList[f.FPos]
	f.FPos++
	return
}

// Return the entry we're currently at.
func (f *FileList) Current() (*Header, error) {
	return f.CurFile, nil
}

func (f *FileList) Seek(pos, whence int) error {
	fLen := len(f.FList)
	switch whence {
	case io.SeekCurrent:
		pos += f.FPos
	case io.SeekEnd:
		pos = fLen - pos
	}
	if pos < 0 {
		f.FPos = 0
	} else if pos >= fLen {
		f.FPos = fLen
	}
	return nil
}

