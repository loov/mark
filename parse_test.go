package mark_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/loov/mark"
	"github.com/loov/mark/html"
)

func S(blocks ...mark.Block) mark.Sequence   { return mark.Sequence(blocks) }
func P(elems ...mark.Inline) *mark.Paragraph { return &mark.Paragraph{Items: elems} }
func T(s string) mark.Text                   { return mark.Text(s) }

var SB = mark.SoftBreak{}

func TestParagraph(t *testing.T) {
	TestCases{{
		In:  "ABC",
		Exp: S(P(T("ABC"))),
	}, { // line-break
		In:  "ABC\nDEF",
		Exp: S(P(T("ABC"), SB, T("DEF"))),
	}, { // non-strict 3 spaces in front of line
		In:  "A\n B\n  C\n   D",
		Exp: S(P(T("A"), SB, T("B"), SB, T("C"), SB, T("D"))),
	}, { // multiple paragraphs
		In:  "A\n\nB",
		Exp: S(P(T("A")), P(T("B"))),
	}}.Run(t)
}

type TestCase struct {
	In   string
	Exp  mark.Sequence
	Errs []error
}

type TestCases []TestCase

func (cases TestCases) Run(t *testing.T) {
	for i, tc := range cases {
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
	out, errs := mark.ParseContent(nil, "", []byte(tc.In))
	if !reflect.DeepEqual(errs, tc.Errs) {
		t.Errorf("#%d%s invalid errors: got %s exp %s", i, br, errs, tc.Errs)
		ok = false
	}
	if !reflect.DeepEqual(out, tc.Exp) {
		outs := strconv.Quote(html.Convert(out))
		exps := strconv.Quote(html.Convert(tc.Exp))
		t.Errorf("#%d%s invalid output:\ngot %v\nexp %v", i, br, outs, exps)
		ok = false
	}
}
