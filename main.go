package main

import (
	"fmt"
	"os"
	"strings"

	"TypescriptToGolang/pkg/csv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . \"a,b,c\\n1,2,3\"")
		os.Exit(1)
	}

	text := strings.NewReplacer(
		"\\n", "\n",
		"\\r", "\r",
		"\\t", "\t",
	).Replace(strings.Join(os.Args[1:], " "))

	parser := csv.NewParser(nil)
	rows, err := parser.Parse(text)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, row := range rows {
		fmt.Printf("%q\n", row)
	}
}
