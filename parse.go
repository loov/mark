package mark

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type parse struct {
	fs     FileSystem
	path   string // relative to fs root
	reader *reader
	*state

	parent *parse // can be nil
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
		lines []string
		class string
	}
}

func ParseFile(fs FileSystem, filename string) (Sequence, []error) {
	name := filepath.ToSlash(filename)
	data, err := fs.ReadFile(name)
	if err != nil {
		return nil, []error{err}
	}
	return ParseContent(fs, name, data)
}

func ParseContent(fs FileSystem, filename string, content []byte) (Sequence, []error) {
	parse := &parse{
		fs:     fs,
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
		case line.ContainsOnly('=') || line.ContainsOnly('-'):
			parse.setext()
		case line.StartsWith("    "):
			parse.code()
		case line.StartsWith("```"):
			parse.fenced()
		case line.StartsWith("{{"):
			parse.include()
		case line.StartsWith("{"):
			parse.modifier()
		default:
			parse.line()
		}
	}
}

// flushes pending paragraph
func (parse *parse) flushParagraph() {
	if len(parse.partial.lines) == 0 {
		parse.partial.class = ""
		return
	}

	para := parse.linesToParagraph(parse.partial.lines)
	seq := parse.currentSequence(lastlevel)
	if parse.partial.class != "" {
		seq.Append(&Modifier{
			Class:   parse.partial.class,
			Content: Sequence{para},
		})
	} else {
		seq.Append(para)
	}

	parse.partial.lines = nil
	parse.partial.class = ""
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

	child := &parse{
		fs:     parent.fs,
		path:   parent.path,
		state:  &state{},
		reader: &reader{},

		parent: parent,
	}
	*child.reader = *parent.reader

	child.reader.setNextLineStart(parent.reader.head.start)
	child.reader.prefixes = append(child.reader.prefixes, prefix{
		symbol: '>',
	})

	child.run()
	parent.reader.head = child.reader.head

	parent.errors = append(parent.errors, child.errors...)
	seq := parent.currentSequence(lastlevel)
	seq.Append(&Quote{
		Category: "",
		Title:    Paragraph{},
		Content:  child.sequence,
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
	// Use child-parser to parse a single list-item
	// push list-item to the list at the end of current sequence

	parent.reader.ignore(delim)
	parent.reader.ignore(' ')

	child := &parse{
		fs:     parent.fs,
		path:   parent.path,
		state:  &state{},
		reader: &reader{},

		parent: parent,
	}
	*child.reader = *parent.reader

	child.reader.setNextLineStart(parent.reader.head.start)
	child.reader.prefixes = append(child.reader.prefixes, prefix{
		symbol: delim,
	})

	child.run()
	parent.reader.head = child.reader.head

	parent.errors = append(parent.errors, child.errors...)

	list := &List{
		Ordered: false,
		Content: nil,
	}

	for _, item := range child.sequence {
		list.Content = append(list.Content, Sequence{item})
	}

	seq := parent.currentSequence(lastlevel)
	seq.Append(list)
}

func (parse *parse) numlist() {
	parse.flushParagraph()
	//TODO: implement
	panic("numlist not implemented")
}

func (parse *parse) setext() {
	if len(parse.partial.lines) != 1 {
		parse.line()
		return
	}
	reader := parse.reader
	section := &Section{}

	//TODO: check for lazy continuation

	reader.ignoreN(' ', 3)
	switch x := reader.peekRune(); x {
	case '=':
		section.Level = 1
	case '-':
		section.Level = 2
	default:
		panic("Invalid setext header symbol " + string(x))
	}

	parse.partial.lines[0] = strings.TrimSpace(parse.partial.lines[0])
	section.Title = *parse.linesToParagraph(parse.partial.lines)
	parse.partial.lines = nil

	seq := parse.currentSequence(section.Level)
	seq.Append(section)
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
	//TODO: figure out whether this is the best way?
	parse.flushParagraph()
	reader := parse.reader

	reader.ignoreN('{', 1)
	reader.ignore(' ')
	reader.ignoreN('.', 1)
	reader.ignoreTrailingN('}', 1)

	parse.partial.class = strings.TrimSpace(reader.rest())
}

func (parent *parse) hasPath(path string) bool {
	for ; parent.parent != nil; parent = parent.parent {
		if parent.path == path {
			return true
		}
	}
	return false
}

func (parser *parse) reltoabs(ref string) string {
	// is it an absolute or non-local
	if strings.HasPrefix(ref, "/") || !isLocalPath(ref) {
		return ref
	}

	return path.Clean(path.Join(path.Dir(parser.path), ref))
}

func (parser *parse) checkPathExists(p string) {
	if isLocalPath(p) {
		if parser.fs == nil {
			parser.check(fmt.Errorf("Cannot find file %s", p))
		} else if err := parser.fs.FileExists(p); err != nil {
			parser.check(fmt.Errorf("Cannot find file %s: %s", p, err))
		}
	}
}

func (parent *parse) include() {
	parentreader := parent.reader
	parent.flushParagraph()

	parentreader.ignoreN('{', 2)
	parentreader.ignoreTrailingN('}', 2)

	file := strings.TrimSpace(parentreader.rest())
	abs := parent.reltoabs(file)

	child := &parse{
		fs:     parent.fs,
		path:   abs,
		state:  &state{},
		reader: &reader{},

		parent: parent,
	}

	if parent.hasPath(abs) {
		parent.check(fmt.Errorf("Cannot recursively include %v", abs))
		return
	}

	content, err := child.fs.ReadFile(abs)
	if err != nil {
		parent.check(fmt.Errorf("Failed to read file %v: %v", abs, err))
		return
	}

	child.reader.content = string(content)
	child.run()

	seq := parent.currentSequence(lastlevel)
	for _, block := range child.sequence {
		if sec, ok := block.(*Section); ok {
			seq = parent.currentSequence(sec.Level)
		}
		seq.Append(block)
	}
	parent.errors = append(child.errors)
}

func (parse *parse) line() {
	reader := parse.reader

	reader.ignore(' ')
	parse.partial.lines = append(parse.partial.lines, reader.rest())
}

func (parse *parse) inline() *Paragraph {
	reader := parse.reader
	return parse.linesToParagraph([]string{reader.rest()})
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
