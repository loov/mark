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
	return fmt.Sprintf("%s line %d: %s", err.Path, err.At, err.Err)
}

func (c *context) check(err error) {
	if err != nil {
		c.errors = append(c.errors, &ParseError{c.path, c.rd.head.start, err})
	}
}

type state struct {
	sequence Sequence
	errors   []error
}

func ParseFile(filename string) (Sequence, []error) {
	dir := fs(filepath.Dir(filename))
	name := filepath.Base(filename)
	data, err := dir.ReadFile(name)
	if err != nil {
		return nil, []error{err}
	}
	return ParseContent(dir, filename, data)
}

func ParseContent(dir Dir, filename string, content []byte) (Sequence, []error) {
	c := &context{
		dir:   dir,
		path:  filename,
		state: &state{},
		rd:    &reader{},
	}
	c.rd.content = string(content)
	c.parse()
	return c.sequence, c.errors
}

func (c *context) parse() {
	for c.rd.nextLine() {
		line := c.rd.currentLine()
		switch {
		case line.IsEmpty():
		case line.StartsWith(">"):
			c.quote()
		case line.StartsWith("***") ||
			line.StartsWith("---") ||
			line.StartsWith("___"):
			c.separator()
		case line.StartsWith("*"):
			c.list()
		case line.StartsWith("-") || line.StartsWith("+"):
			c.list()
		case line.StartsWithNumbering():
			c.numlist()
		case line.StartsTitle():
			c.section()
		case line.StartsWith("    ") || line.StartsWith("\t"):
			c.code()
		case line.StartsWith("```"):
			c.fenced()
		case line.StartsWith("{"):
			c.modifier()
		case line.StartsWith("<{{"):
			c.include()
		default:
			c.paragraph()
		}
	}
}

func (c *context) quote() {

}

func (c *context) separator() {

}

func (c *context) list() {

}

func (c *context) numlist() {

}

func (c *context) section() {
	section := Section{}

	c.rd.ignore3spaces()
	section.Level = c.rd.count('#')
	c.rd.expect(' ')
	section.Title = c.rd.rest()

	c.state.sequence = append(c.state.sequence, section)
}

func (c *context) code() {

}

func (c *context) fenced() {

}

func (c *context) modifier() {

}

func (c *context) include() {

}

func (c *context) paragraph() {

}
