package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/kr/pretty"

	"github.com/loov/mark"
	"github.com/loov/mark/html"
)

func main() {
	// print a clearing block
	fmt.Println("CLEAR")
	fmt.Println(strings.Repeat("\n", 32))
	fmt.Println(strings.Repeat("-", 32))

	pretty.Printf("Parsing example.md\n\n")
	sequence, errs := mark.ParseFile(mark.Dir("."), "example.md")
	for _, err := range errs {
		pretty.Printf("%v\n", err)
	}
	pretty.Printf("\n%# v\n\n", sequence)

	result := html.Convert(sequence)
	ioutil.WriteFile("~example.html", []byte(`
		<meta http-equiv="refresh" content="1">
		<style>
			body { width: 500px; margin: 1em auto; }
			code { background: #000; color: #fff; }
			pre > code { background: inherit; color: inherit; }
			pre  { background: #000; color: #fff; padding: 0.5em; }
			.separator { text-align: center; background: #eee; }

			.warning { background: #fee; }

			p { margin: 0.5em 0; }
			ul {
				border: 1px solid #ccc;
			}
		</style>
	`+result), 0755)

	fmt.Println(result)
}
