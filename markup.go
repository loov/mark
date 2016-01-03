package mark

import "strings"

// Algorithm for marking up
// 0.   tokenize
// 1.   resolve any code spans;
// 2.   resolve any links, call `markup` starting from 3,
//      find starting [ and ending ]( and final );
// 3.   N := maximum number of markup tokens in line;
// 4.0. for ; N > 0; N -- {
// 4.1.    s, t := first markup sequence with at least length N;
// 4.2.    e := next matching sequence with at least length N and token t;
// 4.3.    recurse `markup` on s..e, where unused tokens are inside of the span;
// 4.4.    if tokens weren't used when recursing, keep them outside of the span;
// 4.5. }
// 5.   convert all the items to their appropriate types,
//      concatenating when possible.

type token struct {
	delim rune
	level int
	text  string
	elem  Inline
}

func markupDelimiter(r rune) bool {
	switch r {
	case '`', '[', ']', '(', ')', '*', '_':
		return true
	}
	return false
}

func delimhistogram(tokens []token, isdelim func(r rune) bool) (hist []int) {
	for _, t := range tokens {
		if isdelim(t.delim) {
			if t.level+1 >= len(hist) {
				hist = append(hist, make([]int, t.level-len(hist)+1)...)
			}
			hist[t.level]++
		}
	}
	return
}

func replaceCodeSpans(tokens []token) []token {
	hist := delimhistogram(tokens, func(r rune) bool { return r == '`' })
	for level := len(hist) - 1; level >= 1; level-- {
		count := hist[level]
		var s, e int
		s = -1
		for ; count >= 2; count -= 2 {
			s, e = findpair(s+1, tokens, level, '`', '`')
			if s < 0 || e < 0 {
				break
			}

			text := ""
			for _, t := range tokens[s+1 : e] {
				text += t.String()
			}
			tokens[s] = token{elem: CodeSpan(text)}
			tokens = append(tokens[:s+1], tokens[e+1:]...)
		}
	}

	return tokens
}

func findpair(start int, tokens []token, level int, begin, end rune) (s, e int) {
	s, e = -1, -1
	if start > len(tokens) {
		return
	}

	// find begin
	for i, t := range tokens[start:] {
		if t.delim == begin && t.level == level {
			s = i
			start = start + i + 1
			break
		}
	}

	if s < 0 { // didn't find begin
		return -1, -1
	}

	// find end
	for i, t := range tokens[start:] {
		if t.delim == end && t.level == level {
			// found end
			return s, start + i
		}
	}

	// didn't find end
	return -1, -1
}

/*
func findpairmin(tokens []token, mincount int) (s, e int) {
	var first [96]int
	first['*'] = -1
	first['_'] = -1

	for i, t := range tokens {
		if t.level >= mincount {
			if t.delim == '*' || t.delim == '_' {
				if first[t.delim] >= 0 {
					return first[t.delim], i
				}
				first[t.delim] = i
			}
		}
	}
	return -1, -1
}

func finddelimpair(tokens []token, delim rune) (s, e int) {
	s, e = -1, -1
	for i, t := range tokens {
		if t.delim == delim {
			if s < 0 {
				s = i
			} else if e < 0 {
				e = i
				return
			}
		}
	}
	return -1, -1
}


func clone(tokens []token) []token {
	x := make([]token, len(tokens))
	copy(x, tokens)
	return x
}

func merge(tokens []token, count int) []Inline {
	if len(tokens) == 0 || count == 0 {
		elems := []Inline{}
		for _, t := range tokens {
			if !t.isempty() {
				elems = append(elems, t.Inline())
			}
		}
		return elems
	}

	s, e := findpairmin(tokens, count)
	if s < 0 || e < 0 {
		return merge(tokens, count-1)
	}

	min := tokens[s].count
	if min > tokens[e].count {
		min = tokens[e].count
	}
	tokens[s].count -= min
	tokens[e].count -= min

	between := clone(tokens[s+1 : e])
	center := merge(between, count-1)

	var t token
	if min >= 2 {
		t.elem = Bold(center)
	} else {
		t.elem = Emphasis(center)
	}

	x := append(tokens[:s+1], t)
	x = append(x, tokens[e:]...)

	return merge(x, count)
}
*/

func tokenize(lines []string) (tokens []token) {
	pushdelim := func(r rune) {
		n := len(tokens) - 1
		canadd := n >= 0 && tokens[n].elem == nil
		if canadd && tokens[n].delim == r {
			tokens[n].level++
		} else {
			tokens = append(tokens, token{delim: r, level: 1})
		}
	}

	pushrune := func(r rune) {
		n := len(tokens) - 1
		canadd := n >= 0 && tokens[n].elem == nil
		if canadd && tokens[n].delim == 0 {
			tokens[n].text += string(r)
		} else {
			tokens = append(tokens, token{text: string(r)})
		}
	}

	for i, line := range lines {
		escapenext := false
		for _, r := range line {
			if escapenext {
				pushrune(r)
				escapenext = false
				continue
			}
			if r == '\\' {
				escapenext = true
				continue
			}

			if markupDelimiter(r) {
				pushdelim(r)
			} else {
				pushrune(r)
			}
		}

		if i+1 != len(lines) {
			tokens = append(tokens, token{
				elem: SoftBreak{},
			})
		}
	}

	return tokens
}

func linesToParagraph(lines []string) *Paragraph {
	tokens := tokenize(lines)
	tokens = replaceCodeSpans(tokens)

	var inlines []Inline

	text := ""
	for _, t := range tokens {
		if t.elem != nil {
			if text != "" {
				inlines = append(inlines, Text(text))
				text = ""
			}
			inlines = append(inlines, t.elem)
		} else {
			text += t.String()
		}
	}
	if text != "" {
		inlines = append(inlines, Text(text))
	}

	return &Paragraph{inlines}
}

/* utilities for tokens */

func (t *token) isempty() bool {
	return t.text == "" && t.level == 0 && t.elem == nil
}

func (t *token) String() string {
	if t.delim != 0 {
		if t.level == 0 {
			return ""
		}
		return strings.Repeat(string(t.delim), t.level)
	}
	if t.elem != nil {
		if _, ok := t.elem.(SoftBreak); ok {
			return "\n"
		}
		if _, ok := t.elem.(HardBreak); ok {
			return "\n"
		}
		panic("invalid token to String conversion")
	}
	return t.text
}

func (t *token) Inline() Inline {
	if t.elem != nil {
		return t.elem
	}
	return Text(t.String())
}
