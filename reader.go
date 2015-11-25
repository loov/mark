package mark

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type reader struct {
	content string
	head    span
}

type span struct {
	line  int
	start int
	at    int
	stop  int
	end   int
}

func (rd *reader) nextLine() bool {
	rd.head.line++
	rd.head.start = rd.head.end
	rd.head.at = rd.head.end

	off := strings.IndexAny(rd.content[rd.head.start:], "\r\n")
	if off < 0 {
		return false
	}

	rd.head.stop = rd.head.start + off
	rd.head.end = rd.head.start + off

	// if there's a \rd\n or \n\rd, the head-break takes two runes
	if rd.head.stop+1 < len(rd.content) &&
		rd.content[rd.head.stop] != rd.content[rd.head.stop+1] {
		rd.head.end++
	}

	rd.head.end++
	return true
}

func (rd *reader) resetLine() { rd.head.at = rd.head.start }

// ignores 0-3 spaces
func (rd *reader) ignore3() {
	for p := 0; p < 3; p++ {
		if rd.head.at < rd.head.stop &&
			rd.content[rd.head.at] == ' ' {
			rd.head.at++
		} else {
			break
		}
	}
}

func (rd *reader) count(r rune) int {
	c := 0
	for off, x := range rd.content[rd.head.at:rd.head.stop] {
		if x != r {
			rd.head.at += off
			break
		}
		c++
	}
	return c
}

func (rd *reader) expectFn(valid func(r rune) bool) {
	r, s := utf8.DecodeRuneInString(rd.rest())
	if s <= 0 || !valid(r) {
		panic("invalid symbol")
	}
	rd.head.at += s
}

func (rd *reader) expect(r rune) {
	rd.expectFn(func(x rune) bool { return x == r })
}

func (rd *reader) ignoreFn(valid func(r rune) bool) {
	for {
		r, s := utf8.DecodeRuneInString(rd.rest())
		if s <= 0 || !valid(r) {
			return
		}
		rd.head.at += s
	}
}

func (rd *reader) ignore(r rune) {
	rd.ignoreFn(func(x rune) bool { return x == r })
}

func (rd *reader) ignoreTrailing(r rune) {
	if utf8.RuneLen(r) > 1 {
		panic("unimplemented for trailing large runes")
	}
	for rd.head.at < rd.head.stop && rune(rd.content[rd.head.stop-1]) == r {
		rd.head.stop--
	}
}

func (rd *reader) ignoreSpaceTrailing(r rune) {
	if utf8.RuneLen(r) > 1 {
		panic("unimplemented for trailing large runes")
	}

	stop := rd.head.stop
	for rd.head.at < stop && rune(rd.content[stop-1]) == r {
		stop--
	}
	if rd.head.at < stop && rune(rd.content[stop-1]) == ' ' {
		rd.head.stop = stop - 1
	}
}

type line string

func (line line) trim3() string {
	for p := 0; p < 3 && p < len(line); p++ {
		if line[p] != ' ' {
			return string(line[p:])
		}
	}
	return string(line)
}

func (line line) IsEmpty() bool { return line.trim3() == "" }
func (line line) StartsWith(prefix string) bool {
	return strings.HasPrefix(line.trim3(), prefix)
}

func (line line) StartsWithNumbering() bool {
	for i, r := range line.trim3() {
		if !unicode.IsNumber(r) {
			if r != '.' {
				return false
			}
			return 1 <= i
		}
	}
	return false
}

func (line line) StartsTitle() bool {
	for i, r := range line.trim3() {
		if r != '#' {
			if !unicode.IsSpace(r) {
				return false
			}
			return 1 <= i
		}
	}
	return false
}

// returns current line, excluding line-feeds
func (rd *reader) line() line {
	return line(rd.content[rd.head.start:rd.head.stop])
}

// returns unparsed part of current line, excluding line-feeds
func (rd *reader) rest() string {
	return string(rd.content[rd.head.at:rd.head.stop])
}
