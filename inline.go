package mark

type Inline interface {
	TagInline()
}

func (Text) TagInline()      {}
func (Emphasis) TagInline()  {}
func (Bold) TagInline()      {}
func (CodeSpan) TagInline()  {}
func (LineBreak) TagInline() {}

// Text is plain-text
type Text string

// Emphasis is text that should appear emphasised `<em>`
type Emphasis string

// Bold is text that should appear bold `<b>`
type Bold string

// CodeSpan is text that should appear monospaced `<code>`
type CodeSpan string

// LineBreak is a hard line break
type LineBreak struct{}

func (Callout) TagInline() {}
func (Index) TagInline()   {}
func (Link) TagInline()    {}

// Callout is an element that indicates relation to someother callout `<span class="callout">`
type Callout string

// Index is a hidden point that word-index can link to `<span class="index">`
type Index struct{ Term string }

// Link refers to another page or a node with an ID `<a>`
type Link struct {
	ID      string
	Abbrev  string
	Href    string
	Caption string
	Title   Paragraph
}

func (InlineModifier) TagInline() {}

// InlineModifier creates a span with the specified class `<span>`
type InlineModifier struct {
	Class  string
	Inline Inline
}
