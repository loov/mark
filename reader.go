package mark

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type reader struct {
	prefixes []prefix
	content  string
	head     span
}

type prefix struct {
	symbol rune
	strict bool
}

func (rd *reader) skipprefix(p prefix) bool {
	if !p.strict {
		rd.ignoreN(' ', 3)
	}
	if !rd.expect(p.symbol) {
		return false
	}
	if !p.strict {
		rd.ignore(' ')
	}
	return true
}

func (rd *reader) skipprefixes() bool {
	for _, prefix := range rd.prefixes {
		if !rd.skipprefix(prefix) {
			return false
		}
		rd.head.begin = rd.head.at
	}
	return true
}

type span struct {
	line  int // line number
	start int // line-start without prefix
	begin int // line-start skipped prefix
	at    int // current parsing position
	stop  int // first line ending delimiter
	end   int // next line start
}

func (rd *reader) nextLine() bool {
	if rd.head.end >= len(rd.content) {
		return false
	}
	previoushead := rd.head

	rd.head.line++
	rd.head.start = rd.head.end
	rd.head.at = rd.head.start
	rd.head.begin = rd.head.start

	rd.head.end = len(rd.content)
	rd.head.stop = len(rd.content)

	off := strings.IndexAny(rd.content[rd.head.start:], "\r\n")
	if off >= 0 {
		rd.head.stop = rd.head.start + off
		rd.head.end = rd.head.start + off
	}

	// if there's a \rd\n or \n\rd, the head-break takes two runes
	if rd.head.stop+1 < len(rd.content) {
		next := rd.content[rd.head.stop+1]
		if (next == '\r' || next == '\n') &&
			rd.content[rd.head.stop] != next {
			rd.head.end++
		}
	}
	if rd.head.end < len(rd.content) {
		rd.head.end++
	}

	if !rd.skipprefixes() {
		rd.head = previoushead
		return false
	}
	return true
}

func (rd *reader) resetLine() {
	rd.head.at = rd.head.begin
}

func (rd *reader) setNextLineStart(start int) {
	rd.head.start = start
	rd.head.begin = start
	rd.head.at = start
	rd.head.stop = start
	rd.head.end = start
}

// note: can be only done once
func (rd *reader) undoNextLine() {
	rd.head.line--
	rd.setNextLineStart(rd.head.start)
}

func (rd *reader) ignoreN(expect rune, max int) (count int) {
	if max < 0 {
		max = int(^uint(0) >> 1)
	}
	for p := 0; p < max; p++ {
		if rd.head.at >= rd.head.stop {
			return
		}

		r, sz := utf8.DecodeRuneInString(rd.content[rd.head.at:])
		if r != expect {
			return
		}
		rd.head.at += sz
		count++
	}
	return
}

func (rd *reader) ignore(expect rune) (count int) {
	return rd.ignoreN(expect, -1)
}

func (rd *reader) expect(r rune) bool { return rd.ignoreN(r, 1) == 1 }

func (rd *reader) ignoreTrailing(r rune) (count int) {
	if utf8.RuneLen(r) > 1 {
		panic("unimplemented for trailing large runes")
	}
	for rd.head.at < rd.head.stop && rune(rd.content[rd.head.stop-1]) == r {
		rd.head.stop--
		count++
	}
	return
}

func (rd *reader) ignoreTrailingN(r rune, n int) (count int) {
	//TODO: cleanup
	if utf8.RuneLen(r) > 1 {
		panic("unimplemented for trailing large runes")
	}
	for rd.head.at < rd.head.stop && rune(rd.content[rd.head.stop-1]) == r {
		rd.head.stop--
		count++
		if count >= n {
			return
		}
	}
	return
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
	for p := 0; p <= 3 && p < len(line); p++ {
		if line[p] != ' ' {
			return string(line[p:])
		}
	}
	return string(line)
}

func (line line) IsEmpty() bool { return line.trim3() == "" }

func (line line) StartsWith(prefix string) bool {
	if len(prefix) == 0 {
		return true
	}
	if prefix[0] == ' ' {
		return strings.HasPrefix(string(line), prefix)
	}
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

func (line line) ContainsOnly(r rune) bool {
	foundspace := false
	for i, x := range line.trim3() {
		if x == ' ' {
			if i == 0 {
				return false
			}
			foundspace = true
			continue
		} else if foundspace {
			// if we found a trailing-space and encounter a non-space
			return false
		}

		if x != r {
			return false
		}
	}
	return true
}

// returns current line, excluding line-feeds and prefixes
func (rd *reader) line() line {
	return line(rd.content[rd.head.begin:rd.head.stop])
}

// returns first rune in the unparsed string
func (rd *reader) peekRune() rune {
	r, _ := utf8.DecodeRuneInString(rd.rest())
	return r
}

// returns unparsed part of current line, excluding line-feeds
func (rd *reader) rest() string {
	return string(rd.content[rd.head.at:rd.head.stop])
}
