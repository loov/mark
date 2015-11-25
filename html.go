package mark

import (
	"html"
	"strconv"
)

type HTML interface {
	HTML() string
}

// Inline elements
func (el Text) HTML() string {
	return html.EscapeString(string(el))
}
func (el Emphasis) HTML() string {
	x := html.EscapeString(string(el))
	return "<em>" + x + "</em>"
}
func (el Bold) HTML() string {
	x := html.EscapeString(string(el))
	return "<b>" + x + "</b>"
}
func (el CodeSpan) HTML() string {
	x := html.EscapeString(string(el))
	return "<code>" + x + "</code>"
}
func (el LineBreak) HTML() string {
	return "<br>"
}
func (el Callout) HTML() string {
	return "<!-- TODO Callout -->"
}
func (el Index) HTML() string {
	return "<!-- TODO Index -->"
}
func (el Link) HTML() string {
	return "<!-- TODO Link -->"
}
func (el InlineModifier) HTML() string {
	return "<!-- TODO ." + string(el.Class) + " -->"
}

func (el *Sequence) HTML() (r string) {
	for _, item := range *el {
		if h, ok := item.(HTML); ok {
			r += h.HTML()
		}
	}
	return r
}

func (el *Paragraph) HTML() (r string) {
	for _, item := range el.Items {
		if h, ok := item.(HTML); ok {
			r += h.HTML()
		}
	}
	return "<p>" + r + "</p>"
}

func (el *Section) HTML() (r string) {
	ht := "h" + strconv.Itoa(el.Level)
	return "<section>" +
		"<" + ht + ">" + el.Title.HTML() + "</" + ht + ">" +
		el.Content.HTML() +
		"</section>"
}

func (el *Quote) HTML() string {
	return "<!-- TODO Quote -->"
}

func (el *Modifier) HTML() string {
	return "<!-- TODO Modifier -->"
}

func (el *Code) HTML() string {
	return "<!-- TODO Code -->"
}

func (el *List) HTML() string {
	return "<!-- TODO List -->"
}

func (el *Image) HTML() string {
	return "<!-- TODO Image -->"
}

func (el *Separator) HTML() string {
	if el.Title.IsEmpty() {
		return "<hr>"
	}
	return "<!-- TODO HR -->"
}
