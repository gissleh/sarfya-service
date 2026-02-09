package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
)

var flagSourceFile = flag.String("source-file", "./data-compiled.json", "File containing data.")

func main() {
	flag.Parse()

	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	data, err := loadSarfyaData(*flagSourceFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Examples loaded:", len(data.Examples))

	matchesStr := strings.Builder{}

	ids := make([]string, 0, 16)
	for _, example := range data.Examples {
		for idx, word := range example.Text.WordMap() {
			res, err := dict.Lookup(context.Background(), word, slices.Contains(example.Flags, sarfya.EFReefDialect))
			if err != nil {
				fmt.Println(word, "-- failed:", err)
				fmt.Println("   example:", maxStr(example.Text.RawText(), 64))
				fmt.Println("   url:", example.Source.URL)
				continue
			}

			ids = ids[:0]
			matchesStr.Reset()
			for i, res := range res {
				ids = append(ids, res.ID)

				if i > 0 {
					matchesStr.WriteString(", ")
				}
				matchesStr.WriteString(res.ID)
				matchesStr.WriteString(" (")
				matchesStr.WriteString(res.Word)
				matchesStr.WriteByte(')')
			}

			for _, w := range example.Words[idx] {
				idPref, _ := utf8.DecodeRuneInString(w.ID)
				if !unicode.IsNumber(idPref) {
					continue
				}

				if !slices.Contains(ids, w.ID) {
					fmt.Println(word, "-- does not match", w.ID, "("+w.Word+")")
					fmt.Println("   example:", maxStr(example.Text.RawText(), 64))
					fmt.Println("   matches:", matchesStr.String())
					fmt.Println("   url:", example.Source.URL)
				}
			}
		}
	}
}

func maxStr(str string, length int) string {
	if len(str) > length {
		return str[:length-3] + "..."
	}

	newLine := strings.Index(str, "\n")
	if newLine != -1 {
		return str[:newLine] + "..."
	}

	return str
}

func loadSarfyaData(path string) (*sarfyaJson, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := &sarfyaJson{}
	err = json.NewDecoder(file).Decode(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type sarfyaJson struct {
	Examples map[string]sarfya.Example `json:"examples"`
}
