package mark

import "strings"

type reader struct {
	content string
	head    int
}

func (r *reader) HasPrefix(x string) bool {
	return strings.HasPrefix(r.content[r.head:], x)
}
