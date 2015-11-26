package main

import (
	"fmt"
	"io/ioutil"

	"github.com/kr/pretty"

	"github.com/loov/mark"
	"github.com/loov/mark/html"
)

func main() {
	pretty.Printf("Parsing example.md\n\n")
	sequence, errs := mark.ParseFile("example.md")
	for _, err := range errs {
		pretty.Printf("%v\n", err)
	}
	pretty.Printf("\n%# v\n", sequence)

	result := html.Convert(sequence)
	ioutil.WriteFile("~example.html", []byte(`
		<style>
			body { width: 500px; margin: 1em auto; }
		</style>
	`+result), 0755)
	fmt.Println(result)
}
