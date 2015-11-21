package mark

// Note represents a sidemark or a footnote that should not appear in the main
// text flow.
type Note struct {
	ID      string
	Content Sequence
}

// Ref references node with a specific ID `<a class="reference" href="...">`
type Ref struct {
	ID     string
	Abbrev string
}

func (Ref) TagInline() {}
