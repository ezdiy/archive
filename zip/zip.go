package zip

import (
	"archive/zip"
	"github.com/ezdiy/archive"
	"github.com/ezdiy/archive/util"
	"io"
	"io/ioutil"
)

type Reader struct {
	archive.FileList
	z *zip.Reader
}

func (r *Reader) Next() (*archive.Header, error) {
	if r.Advance() {
		return nil, io.EOF
	}
	return r.Switch(r.z.File[r.CurFile.Index].Open())
}

func Open(input *io.Reader, opt *archive.Options) (ret archive.Reader, e error) {
	if ok, err := util.CheckMagic(input, "PK"); err != nil || !ok {
		return nil, err
	}
	rat := util.MakeReaderAt(*input, &opt.Size)
	if rat == nil || opt.Size == 0 {
		return nil, nil
	}
	a := &Reader{}
	a.Closer = ioutil.NopCloser(nil)
	a.z, e = zip.NewReader(rat, opt.Size)
	if e != nil {
		return nil, e
	}
	for i, v := range a.z.File {
		a.FList = append(a.FList, archive.Header{
			Name:  v.Name,
			Size:  int64(v.UncompressedSize64),
			Time:  v.Modified,
			Index: i,
			IsDir: false,
		})
	}
	return a, e
}

func init() {
	archive.Formats = append(archive.Formats, Open)
}
