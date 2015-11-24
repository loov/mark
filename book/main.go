package main

import (
	"github.com/kr/pretty"
	"github.com/loov/mark"
)

func main() {
	pretty.Printf("Parsing example.md\n")
	sequence, errs := mark.ParseFile("example.md")
	for i, err := range errs {
		pretty.Printf("E%02d: %v\n", i, err)
	}
	pretty.Printf("%# v\n", sequence)
}
