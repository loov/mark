package mark

// Main organizational nodes
type Node interface {
	TagNode()
}

func (Sequence) TagNode()  {}
func (Paragraph) TagNode() {}
func (Section) TagNode()   {}
func (Block) TagNode()     {}
func (Modifier) TagNode()  {}

// Sequence is a container of nodes
type Sequence []Node

// Paragraph represents a `<p>`
type Paragraph []Inline

// Section contains information about a titled Sequence `<section>`
type Section struct {
	Level   int
	Title   string
	Content Sequence
	Notes   []Note
}

// Block represents a nested block, such as quotes or figures `<blockquote>`
type Block struct {
	Category string
	Title    Paragraph
	Content  Sequence
}

// Modifier represents a container with a specific classname `<div>`
type Modifier struct {
	Class   string
	Content Sequence
}

func (CodeBlock) TagNode() {}
func (List) TagNode()      {}
func (Image) TagNode()     {}
func (HR) TagNode()        {}

// CodeBlock is a block of code `<pre>`
type CodeBlock struct {
	Language string
	Content  string
}

// List is a list of different Sequence nodes `<ul>`, `<ol>`
type List struct {
	Numbered bool
	Content  []Sequence
}

// Image refers to an image `<img>`
type Image struct {
	Path string
	Alt  Paragraph
}

// HR is a horizontal-rule with an optional title `<hr>`
type HR struct {
	Title Paragraph
}
