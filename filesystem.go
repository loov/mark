package mark

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileSystem interface {
	ReadFile(path string) ([]byte, error)
}

type Dir string

func (dir Dir) ReadFile(file string) ([]byte, error) {
	full := filepath.Join(string(dir), filepath.FromSlash(file))
	return ioutil.ReadFile(full)
}

type VirtualDir map[string]string

func (dir VirtualDir) ReadFile(file string) ([]byte, error) {
	content, ok := dir[file]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}
