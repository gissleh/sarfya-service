package main

import (
	"context"
	"log"
	"os"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
)

func main() {
	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "sar +> fko")
	}

	input := sarfya.Input{
		ID:   "0",
		Text: "1Fìfya 2sar 3fkol 4fìvefyat!",
		LookupFilter: map[int]string{
			2: "vtr.",
			4: "9316",
		},
		Translations: map[string]string{
			"en": "4(This system) 2+3(is used) 1(like this)!",
			"no": "1Slik 2bruker 3man 4(dette systemet)!",
		},
		Source: sarfya.Source{
			ID:     "1234",
			Date:   "2024-07-07",
			URL:    "github.com/gissleh/sarfya",
			Title:  "Code Example",
			Author: "Meypll",
		},
		Flags: []sarfya.ExampleFlag{
			sarfya.EFNonCanon,
		},
	}

	example, err := sarfya.NewExample(context.Background(), input, dict)
	if err != nil {
		log.Fatalln("Failed to create example from input:", err)
	}

	for i, part := range example.Text {
		log.Printf("example.Text[%d] = { Text: %#+v, IDs: %v }\n", i, part.Text, part.IDs)
	}

	log.Println("Searching:", os.Args[1])
	filter, resolvedSets, err := sarfya.ParseFilter(context.Background(), os.Args[1], dict)
	if err != nil {
		log.Fatalln("Failed to create filter for searching:", err)
	}

	for i, resolvedSet := range resolvedSets {
		match := filter.CheckExample(*example, resolvedSet)
		if match != nil {
			log.Printf("match[%d].Spans = %v", i, match.Spans)
			for lang, spans := range match.TranslationSpans {
				log.Printf("match[%d].TranslationSpans[%#+v] = %v", i, lang, spans)
			}
		}
	}
}
