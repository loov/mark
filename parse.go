package mark

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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

	partial struct {
		paragraph []string
	}
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
	defer parse.flushParagraph()

	reader := parse.reader
	for reader.nextLine() {
		line := reader.line()
		switch {
		case line.IsEmpty():
			parse.flushParagraph()
		case line.StartsWith(">"):
			parse.quote()
		case line.StartsWith("*** ") || line.StartsWith("--- ") || line.StartsWith("___ "):
			// TODO: handle empty item
			parse.separator()
		case line.StartsWith("* ") || line.StartsWith("- ") || line.StartsWith("+ "):
			// TODO: handle empty item
			parse.list()
		case line.StartsWithNumbering():
			parse.numlist()
		case line.StartsTitle():
			parse.section()
		case line.StartsWith("    "):
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

// flushes pending paragraph
func (parse *parse) flushParagraph() {
	if len(parse.partial.paragraph) == 0 {
		return
	}

	para := tokenizeParagraph(parse.partial.paragraph)
	seq := parse.currentSequence(lastlevel)
	seq.Append(para)

	parse.partial.paragraph = nil
}

func (parse *parse) separator() {
	reader := parse.reader
	parse.flushParagraph()

	delim := reader.peekRune()
	reader.ignore(delim)
	reader.ignore(' ')

	reader.ignoreTrailing(delim)
	reader.ignoreTrailing(' ')

	separator := &Separator{}
	separator.Title = *parse.inline()

	seq := parse.currentSequence(lastlevel)
	seq.Append(separator)
}

func (parent *parse) quote() {
	parent.flushParagraph()

	//TODO: implement lazyness
	// http://spec.commonmark.org/0.22/#block-quotes

	parent.reader.ignore(' ')
	if !parent.reader.expect('>') {
		panic("sanity check: " + parent.reader.rest())
	}

	sub := &parse{
		dir:    parent.dir,
		path:   parent.path,
		state:  &state{},
		reader: &reader{},
	}
	*sub.reader = *parent.reader

	sub.reader.setNextLineStart(parent.reader.head.start)
	sub.reader.prefixes = append(sub.reader.prefixes, prefix{
		symbol: '>',
	})

	sub.run()
	parent.reader.head = sub.reader.head

	parent.errors = append(parent.errors, sub.errors...)
	seq := parent.currentSequence(lastlevel)
	seq.Append(&Quote{
		Category: "",
		Title:    Paragraph{},
		Content:  sub.sequence,
	})
}

func (parent *parse) list() {
	parent.flushParagraph()

	parent.reader.ignore(' ')
	delim := parent.reader.peekRune()
	if !(delim == '-' || delim == '+' || delim == '*') {
		panic("sanity check: " + parent.reader.rest())
	}

	//TODO: fix
	// Use sub-parser to parse a single list-item
	// push list-item to the list at the end of current sequence

	parent.reader.ignore(delim)
	parent.reader.ignore(' ')

	sub := &parse{
		dir:    parent.dir,
		path:   parent.path,
		state:  &state{},
		reader: &reader{},
	}
	*sub.reader = *parent.reader

	sub.reader.setNextLineStart(parent.reader.head.start)
	sub.reader.prefixes = append(sub.reader.prefixes, prefix{
		symbol: delim,
	})

	sub.run()
	parent.reader.head = sub.reader.head

	parent.errors = append(parent.errors, sub.errors...)

	list := &List{
		Ordered: false,
		Content: nil,
	}

	for _, item := range sub.sequence {
		list.Content = append(list.Content, Sequence{item})
	}

	seq := parent.currentSequence(lastlevel)
	seq.Append(list)
}

func (parse *parse) numlist() {
	parse.flushParagraph()
	panic("numlist not implemented")
}

func (parse *parse) section() {
	reader := parse.reader
	section := &Section{}

	reader.ignoreN(' ', 3)
	section.Level = reader.ignore('#')
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

	parse.flushParagraph()
	seq := parse.currentSequence(section.Level)
	seq.Append(section)
}

func (parse *parse) code() {
	reader := parse.reader
	parse.flushParagraph()

	code := &Code{}

	line := reader.line()
	if !line.StartsWith("    ") {
		panic("sanity check")
	}
	code.Lines = append(code.Lines, string(line[4:]))

	undo := false
	for reader.nextLine() {
		line := reader.line()
		if line.StartsWith("    ") {
			code.Lines = append(code.Lines, string(line[4:]))
			continue
		}
		if line.trim3() == "" {
			code.Lines = append(code.Lines, "")
			continue
		}
		undo = true
		break
	}
	if undo {
		reader.undoNextLine()
	}

	seq := parse.currentSequence(lastlevel)
	seq.Append(code)
}

func (parse *parse) fenced() {
	reader := parse.reader

	reader.ignore(' ')

	fencesize := reader.ignore('`')
	fence := strings.Repeat("`", fencesize)

	reader.ignore(' ')
	reader.ignoreTrailing(' ')

	code := &Code{}
	code.Language = reader.rest()

	foundend := false
	for reader.nextLine() {
		line := reader.line()
		if line.StartsWith(fence) {
			foundend = true
			break
		}
		code.Lines = append(code.Lines, string(line))
	}

	if !foundend {
		parse.check(errors.New("Did not find ending code fence"))
	}

	parse.flushParagraph()
	seq := parse.currentSequence(lastlevel)
	seq.Append(code)
}

func (parse *parse) modifier() {
	parse.flushParagraph()
	panic("modifier not implemented")
}

func (parse *parse) include() {
	parse.flushParagraph()
	panic("include not implemented")
}

func (parse *parse) line() {
	reader := parse.reader

	reader.ignore(' ')
	parse.partial.paragraph =
		append(parse.partial.paragraph, reader.rest())
}

func (parse *parse) inline() *Paragraph {
	reader := parse.reader
	return tokenizeParagraph([]string{reader.rest()})
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
