package mark

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem interface {
	FileExists(path string) error
	ReadFile(path string) ([]byte, error)
}

type Dir string

func (dir Dir) FileExists(file string) error {
	full := filepath.Join(string(dir), filepath.FromSlash(file))
	_, err := os.Lstat(full)
	return err
}

func (dir Dir) ReadFile(file string) ([]byte, error) {
	full := filepath.Join(string(dir), filepath.FromSlash(file))
	return ioutil.ReadFile(full)
}

type VirtualDir map[string]string

func (dir VirtualDir) FileExists(file string) error {
	full := strings.TrimPrefix(file, "/")
	if _, ok := dir[full]; !ok {
		return os.ErrNotExist
	}
	return nil
}

func (dir VirtualDir) ReadFile(file string) ([]byte, error) {
	full := strings.TrimPrefix(file, "/")
	content, ok := dir[full]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}

func getPathScheme(url string) string {
	for i, c := range url {
		switch {
		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z':
		case '0' <= c && c <= '9' || c == '+' || c == '-' || c == '.':
			if i == 0 {
				return ""
			}
		case c == ':':
			return url[:i]
		default:
			return ""
		}
	}
	return ""
}

func isLocalPath(ref string) bool { return getPathScheme(ref) == "" }
