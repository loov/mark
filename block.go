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
	Items  []Inline
	closed bool
}

func (p *Paragraph) IsEmpty() bool { return len(p.Items) == 0 }

func (p *Paragraph) Append(next Inline) {
	if last := len(p.Items) - 1; last >= 0 {
		prev := p.Items[last]
		if x, ok := Join(prev, next); ok {
			p.Items[last] = x
			return
		}
	}
	p.Items = append(p.Items, next)
}

func (p *Paragraph) Close() { p.closed = true }

func (leading *Paragraph) AppendLine(trailing *Paragraph) {
	if len(trailing.Items) == 0 {
		return
	}

	if len(leading.Items) > 0 {
		leading.Append(Text(" "))
	}

	for _, inline := range trailing.Items {
		leading.Append(inline)
	}
}

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
