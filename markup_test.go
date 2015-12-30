package mark_test

import "testing"

func TestBoldEmphasis(t *testing.T) {
	TestCases{{ // emphasis and bold
		In: "Paragraph *x* **x** ***x*** ****x**** *****x*****.",
		Exp: Seq(Para(
			Text("Paragraph "),
			Em(Text("x")), Text(" "),
			Bold(Text("x")), Text(" "),
			Bold(Text("x")), Text(" "),
			Bold(Text("x")), Text(" "),
			Bold(Text("x")), Text("."),
		)),
	}, { // emphasis and bold, no leading text
		In: "*x* **x** ***x*** ****x**** *****x*****.",
		Exp: Seq(Para(
			Em(Text("x")), Text(" "),
			Bold(Text("x")), Text(" "),
			Bold(Text("x")), Text(" "),
			Bold(Text("x")), Text(" "),
			Bold(Text("x")), Text("."),
		)),
	}, { // emphasis and bold side-by-side
		In: "*x***x**",
		Exp: Seq(Para(
			Em(Text("x")),
			Bold(Text("x")),
		)),
	}, { // bold and emphasis side-by-side
		In: "**x***x*",
		Exp: Seq(Para(
			Bold(Text("x")),
			Em(Text("x")),
		)),
	}, { // asterix inside list
		In:  "* x *",
		Exp: Seq(Ul(Seq(Para(Text("x *"))))),
	}, { // bold nested in em
		In:  "****x*** testing*",
		Exp: Seq(Para(Em(Bold(Text("x")), Text(" testing")))),
	}, { // bold nested in em
		In:  "*testing ***x****",
		Exp: Seq(Para(Em(Text("testing "), Bold(Text("x"))))),
	}, { // em nested in bold
		In:  "****x* testing***",
		Exp: Seq(Para(Bold(Em(Text("x")), Text(" testing")))),
	}, { // em nested in bold
		In:  "***testing *x****",
		Exp: Seq(Para(Bold(Text("testing "), Em(Text("x"))))),
	}}.Run(t)
}

func TestLinks(t *testing.T) {
	TestCases{{ // simple link
		In:  "[title](http://example.com)",
		Exp: Seq(Para(Link("http://example.com", Text("title")))),
	}, { // link with injection
		In:  "[title](\"http://example.com)",
		Exp: Seq(Para(Link("#ZgotmplZ", Text("title")))),
	}, { // link with injection 2
		In: "[title](javascript:console.log('hello'))",
		Exp: Seq(Para(Link("#ZgotmplZ",
			Text("title")))),
	}, { // emph and bold in text
		In: "[*x* **x**](http://example.com)",
		Exp: Seq(Para(Link("http://example.com",
			Em(Text("x")), Text(" "), Bold(Text("x"))))),
	}, { // link takes precedence
		In:  "*[x*](http://example.com)",
		Exp: Seq(Para(Text("*"), Link("http://example.com", Text("x*")))),
	}}.Run(t)
}
