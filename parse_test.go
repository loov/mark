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
		Exp: Seq(Para(Text("ABC"))),
	}, { // line-break
		In:  "ABC\nDEF",
		Exp: Seq(Para(Text("ABC"), SB, Text("DEF"))),
	}, { // non-strict 3 spaces in front of line
		In:  "A\n B\n  C\n   D",
		Exp: Seq(Para(Text("A"), SB, Text("B"), SB, Text("C"), SB, Text("D"))),
	}, { // multiple paragraphs
		In:  "A\n\n\n\n\nB",
		Exp: Seq(Para(Text("A")), Para(Text("B"))),
	}}.Run(t)
}

func TestSection(t *testing.T) {
	TestCases{{
		In:  "# Hello\nWorld",
		Exp: Seq(H(1, Para(Text("Hello")), Para(Text("World")))),
	}, { // trim extra space
		In:  "#     Hello    \nWorld",
		Exp: Seq(H(1, Para(Text("Hello")), Para(Text("World")))),
	}, { // trim trailing #
		In:  "#     Hello    #########   \nWorld",
		Exp: Seq(H(1, Para(Text("Hello")), Para(Text("World")))),
	}, { // h3
		In:  "### Hello\nWorld",
		Exp: Seq(H(3, Para(Text("Hello")), Para(Text("World")))),
	}, { // require space
		In:  "###Hello\nWorld",
		Exp: Seq(Para(Text("###Hello"), SB, Text("World"))),
	}, { // too many ###
		In:   "######## Hello",
		Exp:  Seq(Para(Text("######## Hello"))),
		Errs: []string{"main.md:1: Expected heading, but contained too many #"},
	}, { // nested sections
		In: "# A1\n## A2\n#### A4\n ## B2",
		Exp: Seq(
			H(1, Para(Text("A1")),
				H(2, Para(Text("A2")),
					H(4, Para(Text("A4")))),
				H(2, Para(Text("B2"))),
			)),
	}}.Run(t)
}

func TestQuote(t *testing.T) {
	TestCases{{ // basic
		In:  "> A",
		Exp: Seq(Quote(Para(Text("A")))),
	}, { // multiple lines
		In:  "> A\n> B",
		Exp: Seq(Quote(Para(Text("A"), SB, Text("B")))),
	}, { // lazy spacing
		In:  "> A\n >B\n  >    C",
		Exp: Seq(Quote(Para(Text("A"), SB, Text("B"), SB, Text("C")))),
	}, { // two blocks
		In:  "> A\n\n>B",
		Exp: Seq(Quote(Para(Text("A"))), Quote(Para(Text("B")))),
	}, { // H in block
		In:  "> # Hello\n> World",
		Exp: Seq(Quote(H(1, Para(Text("Hello")), Para(Text("World"))))),
	}, { // nested quote
		In:  ">> A\n>  >B",
		Exp: Seq(Quote(Quote(Para(Text("A"), SB, Text("B"))))),
	}}.Run(t)
}

func TestFence(t *testing.T) {
	TestCases{{ // basic
		In:  "```\nCODE\n```",
		Exp: Seq(Code("", "CODE")),
	}, { // language
		In:  "``` md\nCODE\n```",
		Exp: Seq(Code("md", "CODE")),
	}, { // preserve empty lines
		In:  "```md\n\nCO\n\nDE\n\n```",
		Exp: Seq(Code("md", "", "CO", "", "DE", "")),
	}, { // different symbols
		In:  "```md\n!@#$%^&*()_+/*-+!@#$%^&*()_+/*-+\n```",
		Exp: Seq(Code("md", "!@#$%^&*()_+/*-+!@#$%^&*()_+/*-+")),
	}, { // preserve tabs/spaces
		In:  "```md\n{\n\tX\n   \n    }    \n```",
		Exp: Seq(Code("md", "{", "\tX", "   ", "    }    ")),
	}}.Run(t)
}

func TestIndentCode(t *testing.T) {
	TestCases{{ // basic
		In:  "    CODE",
		Exp: Seq(Code("", "CODE")),
	}, { // preserve empty lines
		In:  "    \n    CO\n    \n    DE\n    ",
		Exp: Seq(Code("", "", "CO", "", "DE", "")),
	}, { // different symbols
		In:  "    !@#$%^&*()_+/*-+!@#$%^&*()_+/*-+",
		Exp: Seq(Code("", "!@#$%^&*()_+/*-+!@#$%^&*()_+/*-+")),
	}, { // preserve tabs/spaces
		In:  "    \tX  ",
		Exp: Seq(Code("", "\tX  ")),
	}, { // lazy lines
		In:  "    A\n\n\n    B",
		Exp: Seq(Code("", "A", "", "", "B")),
	}, { // paragraph ends
		In:  "    A\nB",
		Exp: Seq(Code("", "A"), Para(Text("B"))),
	}}.Run(t)
}

// Convenience functions
func H(level int, title *mark.Paragraph, content ...mark.Block) *mark.Section {
	return &mark.Section{
		Level:   level,
		Title:   *title,
		Content: mark.Sequence(content),
	}
}
func Seq(blocks ...mark.Block) mark.Sequence    { return mark.Sequence(blocks) }
func Quote(blocks ...mark.Block) *mark.Quote    { return &mark.Quote{Content: blocks} }
func Para(elems ...mark.Inline) *mark.Paragraph { return &mark.Paragraph{Items: elems} }
func Text(s string) mark.Text                   { return mark.Text(s) }

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
