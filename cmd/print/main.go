package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/ebnf"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s EBNF_FILE START_PRODUCTION", os.Args[0])
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	g, err := ebnf.Parse(os.Args[1], f)
	if err != nil {
		log.Fatal(err)
	}
	if err := ebnf.Verify(g, os.Args[2]); err != nil {
		log.Fatal(err)
	}

	for _, v := range g {
		printProduction(v, "")
	}
}

func printProduction(e ebnf.Expression, indent string) {
	fmt.Print(indent)
	switch e := e.(type) {
	case *ebnf.Production:
		fmt.Println(e.Name)
		printProduction(e.Expr, indent+strings.Repeat(" ", 2))
	case ebnf.Alternative:
		fmt.Println("alternative")
		for _, alt := range e {
			printProduction(alt, indent+strings.Repeat(" ", 2))
		}
	case ebnf.Sequence:
		fmt.Println("sequence")
		for _, seq := range e {
			printProduction(seq, indent+strings.Repeat(" ", 2))
		}
	case *ebnf.Name:
		fmt.Printf("%q\n", e.String)
	case *ebnf.Token:
		fmt.Printf("%q\n", e.String)
	case *ebnf.Range:
		fmt.Println("range")
		printProduction(e.Begin, indent+strings.Repeat(" ", 2))
		printProduction(e.End, indent+strings.Repeat(" ", 2))
	case *ebnf.Group:
		fmt.Println("group")
		printProduction(e.Body, indent+strings.Repeat(" ", 2))
	case *ebnf.Option:
		fmt.Println("option")
		printProduction(e.Body, indent+strings.Repeat(" ", 2))
	case *ebnf.Repetition:
		fmt.Println("repetition")
		printProduction(e.Body, indent+strings.Repeat(" ", 2))
	case *ebnf.Bad:
		fmt.Println(e.Error)
	}
}
