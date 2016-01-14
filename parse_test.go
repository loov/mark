package mark_test

import (
	"testing"

	"github.com/loov/mark"
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

func TestSetext(t *testing.T) {
	TestCases{{
		In:  "Hello\n===\nWorld",
		Exp: Seq(H(1, Para(Text("Hello")), Para(Text("World")))),
	}, { // trim extra space
		In:  "   Hello    \n   =\nWorld",
		Exp: Seq(H(1, Para(Text("Hello")), Para(Text("World")))),
	}, { // trim trailing space
		In:  "   Hello    \n   =        \nWorld",
		Exp: Seq(H(1, Para(Text("Hello")), Para(Text("World")))),
	}, { // h2
		In:  "Hello\n---\nWorld",
		Exp: Seq(H(2, Para(Text("Hello")), Para(Text("World")))),
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

func TestInclude(t *testing.T) {
	TestCases{{ // simple include
		In: "{{include.md}}",
		FS: mark.VirtualDir{
			"include.md": "Content",
		},
		Exp: Seq(Para(Text("Content"))),
	}, { // nested include
		In: "{{include.md}}",
		FS: mark.VirtualDir{
			"include.md":  "{{include2.md}}",
			"include2.md": "Content",
		},
		Exp: Seq(Para(Text("Content"))),
	}, { // include in quote
		In: "> {{include.md}}",
		FS: mark.VirtualDir{
			"include.md": "Content",
		},
		Exp: Seq(Quote(Para(Text("Content")))),
	}, { // inclusion should follow same block rules
		In: "# First\n{{include.md}}",
		FS: mark.VirtualDir{
			"include.md": "First\n# Second\nSecond",
		},
		Exp: Seq(
			H(1, Para(Text("First")), Para(Text("First"))),
			H(1, Para(Text("Second")), Para(Text("Second"))),
		),
	}, { // proper error with missing file
		In: "{{include.md}}",
		FS: mark.VirtualDir{
			"include.md": "{{include2.md}}\nContent",
		},
		Exp:  Seq(Para(Text("Content"))),
		Errs: []string{"include.md:1: Failed to read file include2.md: file does not exist"},
	}, { // recursion protection
		In: "{{include.md}}",
		FS: mark.VirtualDir{
			"include.md":  "{{include2.md}}",
			"include2.md": "{{include.md}}",
		},
		Errs: []string{"include2.md:1: Cannot recursively include include.md"},
	}}.Run(t)
}
