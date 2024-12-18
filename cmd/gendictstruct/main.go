package main

import (
	"context"
	"fmt"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"log"
	"os"
	"strings"
)

func main() {
	dict := sarfya.WithDerivedPoS(fwewdictionary.Global())
	args := strings.Join(os.Args[1:], " ")

	short := false
	if os.Args[1] == "-s" {
		args = strings.Join(os.Args[2:], " ")
		short = true
	}

	res, err := dict.Lookup(context.Background(), args)
	if err != nil {
		log.Fatalln(err)
		return
	}

	for _, res := range res {
		if short {
			word := res.Word
			if res.InfixIndexes != nil {
				word = word[:res.InfixIndexes[1]] + "<2>" + word[res.InfixIndexes[1]:]
				word = word[:res.InfixIndexes[0]] + "<0><1>" + word[res.InfixIndexes[0]:]
			}

			fmt.Printf("%s:%s:%s\n", res.ID, word, strings.ReplaceAll(res.PoS, " ", ""))
		} else {
			res.Definitions = map[string]string{"en": res.Definitions["en"]}
			fmt.Printf("Struct: %#+v\n", res)
			fmt.Printf("Filter: %#+v\n", res.ToFilter().String())
		}
	}
}
