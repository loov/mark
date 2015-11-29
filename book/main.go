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
	fmt.Println(strings.Repeat("\n", 32))
	pretty.Printf("Parsing example.md\n\n")
	sequence, errs := mark.ParseFile("example.md")
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
		</style>
	`+result), 0755)

	fmt.Println(result)
}
