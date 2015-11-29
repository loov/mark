package mark

import "strings"

func trimsuffix(s string, b byte) string {
	i := len(s)
	for i > 0 && s[i-1] == b {
		i--
	}
	return s[:i]
}

type partial struct {
	tokens []token
}

type token struct {
	delim rune
	count int
	text  string
	elem  Inline
}

func (t *token) isempty() bool {
	return t.text == "" && t.count == 0 && t.elem == nil
}

func (t *token) Inline() Inline {
	if t.elem != nil {
		return t.elem
	}
	if t.delim != 0 {
		return Text(strings.Repeat(string(t.delim), t.count))
	}
	return Text(t.text)
}

func (p *partial) pushdelim(r rune) {
	n := len(p.tokens) - 1
	if n >= 0 && p.tokens[n].delim == r {
		p.tokens[n].count++
	} else {
		p.tokens = append(p.tokens, token{delim: r, count: 1})
	}
}

func (p *partial) pushrune(r rune) {
	n := len(p.tokens) - 1
	if n >= 0 && p.tokens[n].delim == 0 {
		p.tokens[n].text += string(r)
	} else {
		p.tokens = append(p.tokens, token{text: string(r)})
	}
}

func delimtype(t token, elems []Inline) Inline {
	if t.count == 1 {
		return Emphasis(elems)
	} else if t.count == 2 {
		return Bold(elems)
	}
	panic("unhandled")
}

func findpair(tokens []token, delim rune, mincount int) (s, e int) {
	s, e = -1, -1
	for i, t := range tokens {
		if t.delim == delim && t.count >= mincount {
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

	s, e := findpair(tokens, '*', count)
	if s < 0 || e < 0 {
		s, e = findpair(tokens, '_', count)
	}
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

func isdelim(r rune) bool {
	return r == '*' || r == '_'
}

func tokenizeParagraph(lines []string) *Paragraph {
	p := partial{}
	for _, line := range lines {
		escapenext := false
		for _, r := range line {
			if escapenext {
				p.pushrune(r)
				escapenext = false
				continue
			}
			if r == '\\' {
				escapenext = true
				continue
			}

			if isdelim(r) {
				p.pushdelim(r)
			} else {
				p.pushrune(r)
			}
		}
	}

	c := 0
	for _, t := range p.tokens {
		if c < t.count {
			c = t.count
		}
	}

	return &Paragraph{merge(p.tokens, c)}
}
