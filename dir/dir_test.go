package dir

import (
	"github.com/ezdiy/archive"
	"io/ioutil"
	"log"
	"testing"
)

func TestOpen(t *testing.T) {
	a, err := archive.Open("e:/fullbook",nil)
	if err != nil {
		panic(err)
	}
	if a == nil {
		panic("no entry?")
	}
	for {
		r, err := a.Next()
		if err != nil {
			log.Println(err)
			break
		}
		buf, _ := ioutil.ReadAll(a)
		log.Println(r.Index, r.Name, r.Size, len(buf))
	}
	_ = a.Close()
}


