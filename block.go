package mark

// Main organizational Blocbs
type Block interface {
	TagBlock()
}

func (Sequence) TagBlock()  {}
func (Paragraph) TagBlock() {}
func (Section) TagBlock()   {}
func (Quote) TagBlock()     {}
func (Modifier) TagBlock()  {}

// Sequence is a container of blocks
type Sequence []Block

func (s *Sequence) Append(block Block) {
	*s = append(*s, block)
}

// Paragraph represents a `<p>`
type Paragraph struct {
	Items []Inline
}

func (p *Paragraph) IsEmpty() bool { return len(p.Items) == 0 }

// Section contains information about a titled Sequence `<section>`
type Section struct {
	Level   int
	Title   Paragraph
	Content Sequence
	Notes   []Note
}

// Quote represents a nested block, such as quotes or figures `<blockquote>`
type Quote struct {
	Category string
	Title    Paragraph
	Content  Sequence
}

// Modifier represents a container with a specific classname `<div>`
type Modifier struct {
	Class   string
	Content Sequence
}

func (Code) TagBlock()      {}
func (List) TagBlock()      {}
func (Image) TagBlock()     {}
func (Separator) TagBlock() {}

// Code is a block of code `<pre>`
type Code struct {
	Language string
	Content  string
}

// List is a list of different Sequence Blocks `<ul>`, `<ol>`
type List struct {
	Numbered bool
	Content  []Sequence
}

// Image refers to an image `<img>`
type Image struct {
	Path string
	Alt  Paragraph
}

// Separator is a horizontal-rule with an optional title `<hr>`
type Separator struct {
	Title Paragraph
}
