package mark

import "path/filepath"

type context struct {
	dir   Dir
	path  string
	state *state
}

type state struct {
	Sequence Sequence
}

func ParseFile(filename string) (Sequence, error) {
	dir := fs(filepath.Dir(filename))
	name := filepath.Base(filename)
	data, err := dir.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ParseIn(dir, data)
}

func ParseContent(dir Dir, filename string, content []byte) (Sequence, error) {
	state := &state{}
	context{dir, filename, state}.Parse(content)
}
