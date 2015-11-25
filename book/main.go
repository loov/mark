package main

import (
	"fmt"
	"io/ioutil"

	"github.com/kr/pretty"
	"github.com/loov/mark"
)

func main() {
	pretty.Printf("Parsing example.md\n\n")
	sequence, errs := mark.ParseFile("example.md")
	for _, err := range errs {
		pretty.Printf("%v\n", err)
	}
	pretty.Printf("\n%# v\n", sequence)
	result := sequence.HTML()
	ioutil.WriteFile("~example.html", []byte(result), 0755)
	fmt.Println(sequence.HTML())
}
