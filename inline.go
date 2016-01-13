package mark

type Inline interface {
	TagInline()
}

func (Text) TagInline()      {}
func (Emphasis) TagInline()  {}
func (Bold) TagInline()      {}
func (CodeSpan) TagInline()  {}
func (SoftBreak) TagInline() {}
func (HardBreak) TagInline() {}

// Text is plain-text
type Text string

// Emphasis is text that should appear emphasised `<em>`
type Emphasis []Inline

// Bold is text that should appear bold `<b>`
type Bold []Inline

// CodeSpan is text that should appear monospaced `<code>`
type CodeSpan string

// SoftBreak is a soft line break
type SoftBreak struct{}

// HardBreak is a hard line break
type HardBreak struct{}

func (Callout) TagInline() {}
func (Index) TagInline()   {}
func (Link) TagInline()    {}

// Callout is an element that indicates relation to some other callout `<span class="callout">`
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
