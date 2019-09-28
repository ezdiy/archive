package dir

import (
	"github.com/ezdiy/archive"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Reader struct {
	archive.FileList
	Name string
}

func (r *Reader) Next() (ret *archive.Header, err error) {
	if r.Advance() {
		return nil, io.EOF
	}
	return r.Switch(os.Open(filepath.Join(r.Name, r.CurFile.Name)))
}

func Open(input *io.Reader, opt *archive.Options) (ret archive.Reader, e error) {
	nn := *input
	if (*input) != nil || !opt.AllowDir {
		return nil, nil
	}
	r := Reader{}
	r.FileList.Closer = ioutil.NopCloser(nil)
	r.SkipDirs = opt.SkipDirs
	r.Name, e = filepath.Abs(opt.Name)
	e = filepath.Walk(r.Name, func(path string, info os.FileInfo, err error) (e error) {
		if info == nil {
			return
		}
		rPath, _ := filepath.Rel(r.Name, path)
		r.FList = append(r.FList, archive.Header{
			Name:  filepath.ToSlash(rPath),
			Size:  info.Size(),
			Time:  info.ModTime(),
			Index: len(r.FList),
			IsDir: info.IsDir(),
		})
		return nil
	})
	if e == nil {
		ret = &r
	}
	return
}

func init() {
	archive.Formats = append(archive.Formats, Open)
}
