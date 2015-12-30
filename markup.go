package mark

import "strings"

// Algorithm for `markup`
//
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
		if t.count == 0 {
			return ""
		}
		return strings.Repeat(string(t.delim), t.count)
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

func findpairmin(tokens []token, mincount int) (s, e int) {
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

func rawtext(tokens []token) (text string) {
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

func isdelim(r rune) bool {
	return r == '*' || r == '_' || r == '`'
}

func replaceCodeSpans(tokens []token) []token {
	s, e := finddelimpair(tokens, '`')
	if s < 0 || e < 0 {
		return tokens
	}

	tokens[s].count -= 1
	if tokens[s].count == 0 {
		tokens[s].delim = 0
	}
	tokens[e].count -= 1
	if tokens[e].count == 0 {
		tokens[e].delim = 0
	}

	text := rawtext(tokens[s:e])
	tail := replaceCodeSpans(tokens[e:])

	result := tokens[:s]
	result = append(result, token{elem: CodeSpan(text)})
	result = append(result, tail...)
	return result
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

	p.tokens = replaceCodeSpans(p.tokens)

	c := 0
	for _, t := range p.tokens {
		if c < t.count {
			c = t.count
		}
	}
	inline := merge(p.tokens, c)

	return &Paragraph{inline}
}
