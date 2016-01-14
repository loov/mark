package mark

import "strings"

func findnext(delim rune, level int, tokens []token, start int) int {
	for i, t := range tokens[start:] {
		if t.isempty() {
			continue
		}
		if t.delim == delim && level == t.level {
			return start + i
		}
	}
	return -1
}

func findnextdelim(delim rune, tokens []token, start int) int {
	for i, t := range tokens[start:] {
		if t.isempty() {
			continue
		}
		if t.delim == delim && t.level > 0 {
			return start + i
		}
	}
	return -1
}

func cloneTokens(tokens []token) []token {
	r := make([]token, len(tokens))
	copy(r, tokens)
	return r
}

func resolveSimple(tokens []token, links bool) (resolved []token) {
	for s := 0; s < len(tokens); s++ {
		t := tokens[s]
		if t.isempty() {
			continue
		}

		if links {
			switch t.delim {
			case '`':
				e := findnext('`', t.level, tokens, s+1)
				if e < 0 {
					resolved = append(resolved, t)
					continue
				}
				resolved = append(resolved, token{elem: CodeSpan(rawtext(tokens[s+1 : e]))})
				s = e
			case '[':
				e := findnextdelim(']', tokens, s+1)
				if e < 0 || e+1 >= len(tokens) || tokens[e+1].delim != '(' {
					resolved = append(resolved, t)
					continue
				}
				linkend := findnextdelim(')', tokens, e+1)
				if linkend < 0 {
					resolved = append(resolved, t)
					continue
				}

				caption := cloneTokens(tokens[s : e+1])
				link := cloneTokens(tokens[e+1 : linkend+1])

				// remove wrapping []
				caption[0].level--
				caption[len(caption)-1].level--
				// remove wrapping ()
				link[0].level--
				link[len(link)-1].level--

				resolved = append(resolved, token{
					elem: Link{
						Title: Paragraph{resolve(caption)},
						Href:  rawtext(link),
					},
				})
				s = linkend
			default:
				resolved = append(resolved, t)
			}
		} else {
			switch t.delim {
			case '*', '_':
				e := findnextdelim(t.delim, tokens, s+1)
				if e < 0 {
					resolved = append(resolved, t)
					continue
				}

				minlevel := t.level
				if minlevel > tokens[e].level {
					minlevel = tokens[e].level
				}

				content := cloneTokens(tokens[s:e])
				content[0].level -= minlevel
				tokens[e].level -= minlevel

				if minlevel >= 2 {
					resolved = append(resolved, token{elem: Bold(resolve(content))})
				} else {
					resolved = append(resolved, token{elem: Emphasis(resolve(content))})
				}

				s = e - 1
			default:
				resolved = append(resolved, t)
			}
		}
	}
	return
}

func resolveText(tokens []token) (resolved []token) {
	text := ""
	for _, t := range tokens {
		if t.isempty() {
			continue
		}
		if t.elem != nil {
			if text != "" {
				resolved = append(resolved, token{elem: Text(text)})
				text = ""
			}
			resolved = append(resolved, token{elem: t.elem})
		} else {
			text += t.String()
		}
	}
	if text != "" {
		resolved = append(resolved, token{elem: Text(text)})
	}
	return
}

func rawtext(tokens []token) (text string) {
	for _, t := range tokens {
		text += t.String()
	}
	return
}

func resolve(tokens []token) []Inline {
	tokens = resolveSimple(tokens, true)
	tokens = resolveSimple(tokens, false)
	tokens = resolveText(tokens)
	var inlines []Inline
	for _, t := range tokens {
		if t.elem == nil {
			panic("unhandled delimiter or text")
		}
		inlines = append(inlines, t.elem)
	}
	return inlines
}

func linesToParagraph(lines []string) *Paragraph {
	return &Paragraph{resolve(tokenizeLines(lines))}
}

/* tokenization */
func markupDelimiter(r rune) bool {
	switch r {
	case '`', '[', ']', '(', ')', '*', '_':
		return true
	}
	return false
}

type token struct {
	delim rune
	level int
	text  string
	elem  Inline
}

func tokenizeLines(lines []string) (tokens []token) {
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
