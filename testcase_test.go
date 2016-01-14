package mark_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/loov/mark"
	"github.com/loov/mark/html"
)

// Convenience functions
func H(level int, title *mark.Paragraph, content ...mark.Block) *mark.Section {
	return &mark.Section{
		Level:   level,
		Title:   *title,
		Content: mark.Sequence(content),
	}
}
func Seq(blocks ...mark.Block) mark.Sequence { return mark.Sequence(blocks) }
func Ul(seqs ...mark.Sequence) *mark.List    { return &mark.List{Ordered: false, Content: seqs} }
func Ol(seqs ...mark.Sequence) *mark.List    { return &mark.List{Ordered: true, Content: seqs} }
func Quote(blocks ...mark.Block) *mark.Quote { return &mark.Quote{Content: blocks} }
func Text(s string) mark.Text                { return mark.Text(s) }

func Para(elems ...mark.Inline) *mark.Paragraph { return &mark.Paragraph{Items: elems} }
func Em(elems ...mark.Inline) mark.Emphasis     { return mark.Emphasis(elems) }
func Bold(elems ...mark.Inline) mark.Bold       { return mark.Bold(elems) }
func CodeSpan(s string) mark.CodeSpan           { return mark.CodeSpan(s) }

func Link(href string, title ...mark.Inline) mark.Link {
	t := Para(title...)
	return mark.Link{
		Title: *t,
		Href:  href,
	}
}

var SB = mark.SoftBreak{}

func Code(lang string, lines ...string) *mark.Code {
	return &mark.Code{
		Language: lang,
		Lines:    lines,
	}
}

type TestCase struct {
	In   string
	Exp  mark.Sequence
	Skip bool
	FS   mark.FileSystem
	Errs []string
}

type TestCases []TestCase

func (cases TestCases) Run(t *testing.T) {
	for i, tc := range cases {
		if tc.Skip {
			continue
		}

		// unix
		t1 := tc
		t1.In = strings.Replace(tc.In, "\n", "\x0A", -1)
		if !t1.Run("↓ ", i, t) {
			continue
		}

		// old mac
		t2 := tc
		t2.In = strings.Replace(tc.In, "\n", "\x0D", -1)
		if !t2.Run("← ", i, t) {
			continue
		}

		// windows
		t3 := tc
		t3.In = strings.Replace(tc.In, "\n", "\x0D\x0A", -1)
		if !t3.Run("←↓", i, t) {
			continue
		}

		// why would you do this?
		t4 := tc
		t4.In = strings.Replace(tc.In, "\n", "\x0A\x0D", -1)
		if !t4.Run("↓←", i, t) {
			continue
		}
	}
}

func (tc *TestCase) Run(br string, i int, t *testing.T) (ok bool) {
	ok = true
	out, errs := mark.ParseContent(tc.FS, "main.md", []byte(tc.In))

	sameerr := len(errs) == len(tc.Errs)
	if sameerr {
		for i, errtext := range tc.Errs {
			if errtext != errs[i].Error() {
				sameerr = false
				break
			}
		}
	}
	if !sameerr {
		t.Errorf("#%d%s invalid errors: got %s exp %s", i, br, errs, tc.Errs)
		ok = false
	}

	if !reflect.DeepEqual(out, tc.Exp) {
		outs := strconv.Quote(html.Convert(out))
		exps := strconv.Quote(html.Convert(tc.Exp))
		t.Errorf("#%d%s invalid output:\ngot %v\nexp %v", i, br, outs, exps)
		/* verbose
		pretty.Println(out)
		pretty.Println(tc.Exp)
		/**/
		ok = false
	}
	return
}

func ifthen(v bool, a, b mark.Sequence) mark.Sequence {
	if v {
		return a
	}
	return b
}
