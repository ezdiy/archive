package libarchive

import (
	"github.com/ezdiy/archive"
	"io"
	"io/ioutil"
	"log"
	"testing"
)

func TestLibarchOpen(t *testing.T) {
	a, err := archive.Open("e:/illusion_part1.7z", nil)
	if err != nil {
		panic(err)
	}
	var bigbuf [1024 * 1024]byte
	for {
		r, err := a.Next()
		if err != nil {
			log.Println(err)
			break
		}
		n, _ := io.CopyBuffer(ioutil.Discard, a, bigbuf[:])
		log.Println(r.Name, r.Size, n)
	}
	_ = a.Close()
}
