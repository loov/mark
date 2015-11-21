package mark

import (
	"io/ioutil"
	"path/filepath"
)

type Dir interface {
	ReadFile(path string) ([]byte, error)
}

type fs string

func (dir fs) ReadFile(file string) ([]byte, error) {
	full := filepath.Join(string(dir), filepath.FromSlash(file))
	return ioutil.ReadFile(full)
}
