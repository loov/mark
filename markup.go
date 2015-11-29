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

func (t *token) String() string {
	if t.delim != 0 {
		return strings.Repeat(string(t.delim), t.count)
	}
	return t.text
}

func (t *token) Inline() Inline {
	if t.elem != nil {
		return t.elem
	}
	return Text(t.String())
}

func (p *partial) pushdelim(r rune) {
	n := len(p.tokens) - 1
	canadd := n >= 0 && p.tokens[n].elem == nil
	if canadd && p.tokens[n].delim == r {
		p.tokens[n].count++
	} else {
		p.tokens = append(p.tokens, token{delim: r, count: 1})
	}
}

func (p *partial) pushrune(r rune) {
	n := len(p.tokens) - 1
	canadd := n >= 0 && p.tokens[n].elem == nil
	if canadd && p.tokens[n].delim == 0 {
		p.tokens[n].text += string(r)
	} else {
		p.tokens = append(p.tokens, token{text: string(r)})
	}
}

func findpair(tokens []token, mincount int) (s, e int) {
	var first [96]int
	first['*'] = -1
	first['_'] = -1

	for i, t := range tokens {
		if t.count >= mincount {
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

func asraw(tokens []token) (text string) {
	for _, t := range tokens {
		text += t.String()
	}
	return text
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

	s, e := findpair(tokens, count)
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
	return r == '*' || r == '_' || r == '`'
}

func tokenizeParagraph(lines []string) *Paragraph {
	p := partial{}
	for i, line := range lines {
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

		if i+1 != len(lines) {
			p.tokens = append(p.tokens, token{
				elem: SoftBreak{},
			})
		}
	}

	// p.tokens = replaceCodeSpans(p.tokens)

	c := 0
	for _, t := range p.tokens {
		if c < t.count {
			c = t.count
		}
	}
	inline := merge(p.tokens, c)

	return &Paragraph{inline}
}
