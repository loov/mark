package main

import (
	"github.com/kr/pretty"
	"github.com/loov/mark"
)

func main() {
	pretty.Printf("Parsing example.md\n")
	sequence, errs := mark.ParseFile("example.md")
	for _, err := range errs {
		pretty.Printf("%v\n", err)
	}
	pretty.Printf("%# v\n", sequence)
}
