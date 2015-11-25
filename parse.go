package mark

import (
	"errors"
	"fmt"
	"path/filepath"
)

type parse struct {
	dir    Dir
	path   string
	reader *reader
	*state
}

type ParseError struct {
	Path string
	Line int
	Err  error
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("%s:%d: %s", err.Path, err.Line, err.Err)
}

func (parse *parse) check(err error) {
	if err != nil {
		parse.errors = append(parse.errors, &ParseError{parse.path, parse.reader.head.line, err})
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
	parse := &parse{
		dir:    dir,
		path:   filename,
		state:  &state{},
		reader: &reader{},
	}
	parse.reader.content = string(content)
	parse.line()
	return parse.sequence, parse.errors
}

const last = 1 << 10

func (parse *parse) context(level int) *Sequence {
	seq := &parse.sequence
	for {
		if len(*seq) == 0 {
			return seq
		}

		if sec, ok := (*seq)[len(*seq)-1].(*Section); ok {
			if sec.Level >= level {
				return seq
			} else {
				seq = &sec.Content
			}
		} else {
			return seq
		}
	}
	panic("unreachable")
}

func (parse *parse) line() {
	for parse.reader.nextLine() {
		line := parse.reader.line()
		switch {
		case line.IsEmpty():
		case line.StartsWith(">"):
			parse.quote()
		case line.StartsWith("***") ||
			line.StartsWith("---") ||
			line.StartsWith("___"):
			parse.separator()
		case line.StartsWith("*"):
			parse.list()
		case line.StartsWith("-") || line.StartsWith("+"):
			parse.list()
		case line.StartsWithNumbering():
			parse.numlist()
		case line.StartsTitle():
			parse.section()
		case line.StartsWith("    ") || line.StartsWith("\t"):
			parse.code()
		case line.StartsWith("```"):
			parse.fenced()
		case line.StartsWith("{"):
			parse.modifier()
		case line.StartsWith("<{{"):
			parse.include()
		default:
			parse.paragraph()
		}
	}
}

func (parse *parse) quote() {

}

func (parse *parse) separator() {

}

func (parse *parse) list() {

}

func (parse *parse) numlist() {

}

func (parse *parse) section() {
	section := &Section{}

	parse.reader.ignore3()
	section.Level = parse.reader.count('#')
	if !order(1, section.Level, 6) {
		parse.check(errors.New("Expected heading, but contained too many #"))
		parse.reader.resetLine()
		parse.paragraph()
		return
	}
	parse.reader.expect(' ')
	parse.reader.ignore(' ')

	parse.reader.ignoreTrailing(' ')
	parse.reader.ignoreSpaceTrailing('#')
	parse.reader.ignoreTrailing(' ')

	section.Title = parse.reader.rest()

	context := parse.context(section.Level)
	*context = append(*context, section)
}

func (parse *parse) code() {

}

func (parse *parse) fenced() {

}

func (parse *parse) modifier() {

}

func (parse *parse) include() {

}

func (parse *parse) paragraph() {

}

func order(xs ...int) bool {
	if len(xs) == 0 {
		return true
	}
	p := xs[0]
	for _, x := range xs[1:] {
		if p > x {
			return false
		}
		p = x
	}
	return true
}
