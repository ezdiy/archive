package rar

import (
	"github.com/ezdiy/archive"
	"io/ioutil"
	"log"
	"testing"
)

func TestRarOpen(t *testing.T) {
	a, err := archive.Open("e:/loc.rar", nil)
	if err != nil {
		panic(err)
	}

	for {
		r, err := a.Next()
		if err != nil {
			log.Println(err)
			break
		}
		b, err := ioutil.ReadAll(a)
		if err != nil {
			log.Println(r.Name, r.Size, err)
		} else {
			log.Println(r.Name, r.Size, len(b))
		}
	}
	_ = a.Close()
}
