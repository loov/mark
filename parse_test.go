package mark_test

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/loov/mark"
	"github.com/loov/mark/html"
)

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

func TestSection(t *testing.T) {
	TestCases{{
		In:  "# Hello\nWorld",
		Exp: S(Section(1, P(T("Hello")), P(T("World")))),
	}, { // trim extra space
		In:  "#     Hello    \nWorld",
		Exp: S(Section(1, P(T("Hello")), P(T("World")))),
	}, { // trim trailing #
		In:  "#     Hello    #########   \nWorld",
		Exp: S(Section(1, P(T("Hello")), P(T("World")))),
	}, { // h3
		In:  "### Hello\nWorld",
		Exp: S(Section(3, P(T("Hello")), P(T("World")))),
	}, { // require space
		In:  "###Hello\nWorld",
		Exp: S(P(T("###Hello"), SB, T("World"))),
	}, { // too many ###
		In:   "######## Hello",
		Exp:  S(P(T("######## Hello"))),
		Errs: []string{"main.md:1: Expected heading, but contained too many #"},
	}, { // nested sections
		In: "# A1\n## A2\n#### A4\n ## B2",
		Exp: S(
			Section(1, P(T("A1")),
				Section(2, P(T("A2")),
					Section(4, P(T("A4")))),
				Section(2, P(T("B2"))),
			)),
	}}.Run(t)
}

func TestQuote(t *testing.T) {
	TestCases{{ // basic
		In:  "> A",
		Exp: S(Q(P(T("A")))),
	}, { // multiple lines
		In:  "> A\n> B",
		Exp: S(Q(P(T("A"), SB, T("B")))),
	}, { // lazy spacing
		In:  "> A\n >B\n  >    C",
		Exp: S(Q(P(T("A"), SB, T("B"), SB, T("C")))),
	}, { // two blocks
		In:  "> A\n\n>B",
		Exp: S(Q(P(T("A"))), Q(P(T("B")))),
	}, { // section in block
		In:  "> # Hello\n> World",
		Exp: S(Q(Section(1, P(T("Hello")), P(T("World"))))),
	}, { // nested quote
		In:  ">> A\n>  >B",
		Exp: S(Q(Q(P(T("A"), SB, T("B"))))),
	}}.Run(t)
}

func TestFence(t *testing.T) {
	TestCases{{ // basic
		In:  "```\nCODE\n```",
		Exp: S(Code("", "CODE")),
	}, { // language
		In:  "``` md\nCODE\n```",
		Exp: S(Code("md", "CODE")),
	}, { // preserve empty lines
		In:  "```md\n\nCO\n\nDE\n\n```",
		Exp: S(Code("md", "", "CO", "", "DE", "")),
	}, { // different symbols
		In:  "```md\n!@#$%^&*()_+/*-+!@#$%^&*()_+/*-+\n```",
		Exp: S(Code("md", "!@#$%^&*()_+/*-+!@#$%^&*()_+/*-+")),
	}, { // preserve tabs/spaces
		In:  "```md\n{\n\tX\n   \n    }    \n```",
		Exp: S(Code("md", "{", "\tX", "   ", "    }    ")),
	}}.Run(t)
}

// Convenience functions
func Section(level int, title *mark.Paragraph, content ...mark.Block) *mark.Section {
	return &mark.Section{
		Level:   level,
		Title:   *title,
		Content: mark.Sequence(content),
	}
}
func S(blocks ...mark.Block) mark.Sequence   { return mark.Sequence(blocks) }
func Q(blocks ...mark.Block) *mark.Quote     { return &mark.Quote{Content: blocks} }
func P(elems ...mark.Inline) *mark.Paragraph { return &mark.Paragraph{Items: elems} }
func T(s string) mark.Text                   { return mark.Text(s) }

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
	Errs []string
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
	out, errs := mark.ParseContent(nil, "main.md", []byte(tc.In))

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
		ok = false
	}
	return
}
