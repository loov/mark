package mark

import "strings"

type markup struct {
	*parse
}

func (markup markup) findnext(delim rune, level int, tokens []token, start int) int {
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

func (markup markup) findnextdelim(delim rune, tokens []token, start int) int {
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

func (markup markup) cloneTokens(tokens []token) []token {
	r := make([]token, len(tokens))
	copy(r, tokens)
	return r
}

func (markup markup) simple(tokens []token, links bool) (resolved []token) {
	for s := 0; s < len(tokens); s++ {
		t := tokens[s]
		if t.isempty() {
			continue
		}

		if links {
			switch t.delim {
			case '`':
				e := markup.findnext('`', t.level, tokens, s+1)
				if e < 0 {
					resolved = append(resolved, t)
					continue
				}
				resolved = append(resolved, token{elem: CodeSpan(markup.rawtext(tokens[s+1 : e]))})
				s = e
			case '!':
				// TODO: implement alternate captions `![Alt text](/path/to/img.jpg "Optional title")`
				// TODO: implement reference links `![Alt text][id]`
				//                                 `[id]: url/to/image  "Optional title attribute"`
				if s+1 >= len(tokens) || tokens[s].level != 1 || tokens[s+1].delim != '[' {
					resolved = append(resolved, t)
					continue
				}

				capstart := s + 1
				capeend := markup.findnextdelim(']', tokens, capstart+1)
				if capeend < 0 || capeend+1 >= len(tokens) || tokens[capeend+1].delim != '(' {
					resolved = append(resolved, t)
					continue
				}
				linkend := markup.findnextdelim(')', tokens, capeend+1)
				if linkend < 0 {
					resolved = append(resolved, t)
					continue
				}

				caption := markup.cloneTokens(tokens[capstart : capeend+1])
				link := markup.cloneTokens(tokens[capeend+1 : linkend+1])

				// remove wrapping []
				caption[0].level--
				caption[len(caption)-1].level--
				// remove wrapping ()
				link[0].level--
				link[len(link)-1].level--

				href := markup.reltoabs(markup.rawtext(link))
				markup.checkPathExists(href)

				resolved = append(resolved, token{
					elem: Image{
						Alt:  Paragraph{markup.resolve(caption)},
						Href: href,
					},
				})
				s = linkend
			case '[':
				//TODO: implement title attribute `[an example](http://example.com/ "Title")`
				//TODO: implement reference links `This is [an example][id] reference-style link.`
				//                                `[id]: example.com  "Optional title attribute"`
				capstart := s
				capeend := markup.findnextdelim(']', tokens, capstart+1)
				if capeend < 0 || capeend+1 >= len(tokens) || tokens[capeend+1].delim != '(' {
					resolved = append(resolved, t)
					continue
				}
				linkend := markup.findnextdelim(')', tokens, capeend+1)
				if linkend < 0 {
					resolved = append(resolved, t)
					continue
				}

				caption := markup.cloneTokens(tokens[capstart : capeend+1])
				link := markup.cloneTokens(tokens[capeend+1 : linkend+1])

				// remove wrapping []
				caption[0].level--
				caption[len(caption)-1].level--
				// remove wrapping ()
				link[0].level--
				link[len(link)-1].level--

				href := markup.reltoabs(markup.rawtext(link))
				markup.checkPathExists(href)

				resolved = append(resolved, token{
					elem: Link{
						Title: Paragraph{markup.resolve(caption)},
						Href:  href,
					},
				})
				s = linkend
			default:
				resolved = append(resolved, t)
			}
		} else {
			switch t.delim {
			case '*', '_':
				e := markup.findnextdelim(t.delim, tokens, s+1)
				if e < 0 {
					resolved = append(resolved, t)
					continue
				}

				minlevel := t.level
				if minlevel > tokens[e].level {
					minlevel = tokens[e].level
				}

				content := markup.cloneTokens(tokens[s:e])
				content[0].level -= minlevel
				tokens[e].level -= minlevel

				if minlevel >= 2 {
					resolved = append(resolved, token{elem: Bold(markup.resolve(content))})
				} else {
					resolved = append(resolved, token{elem: Emphasis(markup.resolve(content))})
				}

				s = e - 1
			default:
				resolved = append(resolved, t)
			}
		}
	}
	return
}

func (markup markup) text(tokens []token) (resolved []token) {
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

func (markup markup) rawtext(tokens []token) (text string) {
	for _, t := range tokens {
		text += t.String()
	}
	return
}

func (markup markup) resolve(tokens []token) []Inline {
	tokens = markup.simple(tokens, true)
	tokens = markup.simple(tokens, false)
	tokens = markup.text(tokens)
	var inlines []Inline
	for _, t := range tokens {
		if t.elem == nil {
			panic("unhandled delimiter or text")
		}
		inlines = append(inlines, t.elem)
	}
	return inlines
}

func (parse *parse) linesToParagraph(lines []string) *Paragraph {
	return &Paragraph{markup{parse}.resolve(tokenizeLines(lines))}
}

/* tokenization */
func markupDelimiter(r rune) bool {
	switch r {
	case '`', '[', ']', '(', ')', '*', '_', '!':
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
