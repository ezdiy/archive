package zip

import (
	"github.com/ezdiy/archive"
	"io/ioutil"
	"log"
	"testing"
)

func TestZipOpen(t *testing.T) {
	a, err := archive.Open("e:/test.zip", nil)
	if err != nil {
		panic(err)
	}

	for {
		r, err := a.Next()
		if err != nil {
			log.Println(err)
			break
		}
		buf, _ := ioutil.ReadAll(a)
		log.Println(r.Name, r.Size, len(buf))
	}
	_ = a.Close()
}
