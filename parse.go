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
	parse.run()
	return parse.sequence, parse.errors
}

const lastlevel = 1 << 10

func (parse *parse) currentSequence(level int) *Sequence {
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

func (parse *parse) run() {
	reader := parse.reader
	for reader.nextLine() {
		line := reader.line()
		switch {
		case line.IsEmpty():
			parse.closeParagraph()

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
			parse.line()
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
	reader := parse.reader
	section := &Section{}

	reader.ignore3()
	section.Level = reader.count('#')
	if !order(1, section.Level, 6) {
		parse.check(errors.New("Expected heading, but contained too many #"))
		reader.resetLine()
		parse.line()
		return
	}

	if !reader.expect(' ') {
		parse.check(errors.New("Expected space after leading #"))
		reader.resetLine()
		parse.line()
		return
	}
	reader.ignore(' ')

	reader.ignoreTrailing(' ')
	reader.ignoreSpaceTrailing('#')
	reader.ignoreTrailing(' ')

	section.Title = *parse.inline()

	seq := parse.currentSequence(section.Level)
	seq.Append(section)
}

func (parse *parse) code() {

}

func (parse *parse) fenced() {

}

func (parse *parse) modifier() {

}

func (parse *parse) include() {

}

func (parse *parse) line() {
	reader := parse.reader

	reader.ignore(' ')
	reader.ignoreTrailing(' ')

	line := parse.inline()

	seq := parse.currentSequence(lastlevel)
	if len(*seq) == 0 {
		seq.Append(line)
	} else {
		if para, ok := (*seq)[len(*seq)-1].(*Paragraph); ok && !para.closed {
			para.AppendLine(line)
		} else {
			seq.Append(line)
		}
	}
}

func (parse *parse) closeParagraph() {
	seq := parse.currentSequence(lastlevel)
	if len(*seq) == 0 {
		return
	}
	if para, ok := (*seq)[len(*seq)-1].(*Paragraph); ok {
		para.Close()
	}
}

func (parse *parse) inline() *Paragraph {
	reader := parse.reader
	items := &Paragraph{}
	items.Append(Text(reader.rest()))
	return items
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
