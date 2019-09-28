package sevenzip

import (
	"github.com/ezdiy/archive"
	"io/ioutil"
	"log"
	"testing"
)

func Test7z(t *testing.T) {
	f, _ := archive.Open("e:/fullbook.7z", nil)
	for {
		e, err := f.Next()
		if e == nil {
			log.Println(err)
			break
		}
		bb, _ := ioutil.ReadAll(f)
		log.Println(e.Name, e.Size, len(bb), e.Time)
	}
	f.Close()
}
