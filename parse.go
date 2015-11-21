package mark

import (
	"fmt"
	"path/filepath"
)

type context struct {
	dir  Dir
	path string
	rd   *reader
	*state
}

type ParseError struct {
	Path string
	At   int
	Err  error
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("%s@%d: %s", err.Path, err.At, err.Err)
}

func (c *context) check(err error) {
	if err != nil {
		c.errors = append(c.errors, &ParseError{c.path, c.rd.head, err})
	}
}

type state struct {
	sequence Sequence
	errors   []error
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

func ParseContent(dir Dir, filename string, content []byte) (Sequence, []error) {
	c := &context{
		dir:   dir,
		path:  filename,
		state: &state{},
		rd:    &reader{string(content)},
	}
	c.parse()
	return c.sequence, c.errors
}

func (c *context) parse() {
	for c.rd.nextLine() {
		line := c.rd.currentLine()
		switch {
		case line.IsEmpty():
		case line.StartsWith(">"):
		case line.StartsWith("*"):
		case line.StartsWith("#"):
		case line.StartsWith("  ") || line.StartsWith("\t"):
		case line.StartsWith("```"):
		case line.StartsWith("{"):
		case line.StartsWith("<{{"):
		}
	}
}
