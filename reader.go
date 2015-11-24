package mark

import (
	"strings"
	"unicode"
)

type reader struct {
	content string
	head    span
}

type span struct {
	start int
	stop  int
	end   int
}

func (rd *reader) HasPrefix(x string) bool {
	return strings.HasPrefix(rd.content[rd.head.start:], x)
}

func (rd *reader) nextLine() bool {
	rd.head.start = rd.head.end

	off := strings.IndexAny(rd.content[rd.head.start:], "\rd\n")
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

	return true
}

func (rd *reader) ignore3spaces() {
	for p := 0; p < 3; p++ {
		if rd.head.start < rd.head.stop &&
			rd.content[rd.head.start] == ' ' {
			rd.head.start++
		} else {
			break
		}
	}
}

func (rd *reader) count(r rune) int {
	c := 0
	for off, x := range rd.content[rd.head.start:rd.head.stop] {
		if x != r {
			rd.head.start += off
			break
		}
		c++
	}
	return c
}

func (rd *reader) rest() string {
	return string(rd.currentLine())
}

func (rd *reader) expect(r rune) {
	if rd.head.start >= rd.head.stop ||
		rd.content[rd.head.start] != ' ' {
		panic("expected " + string(r))
	}
	rd.head.start++
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
			if r != ' ' {
				return false
			}
			return 1 <= i && i <= 6
		}
	}
	return false
}

// returns current line, excluding line-feeds
func (rd *reader) currentLine() line {
	return line(rd.content[rd.head.start:rd.head.stop])
}
