package emphasis

import (
	"context"
	"strings"
	"sync"

	"github.com/gissleh/litxap"
	litxapfwew "github.com/gissleh/litxap-fwew"
	"github.com/gissleh/sarfya"
)

var dictOnce sync.Once
var dictionary litxap.Dictionary

func Run(ctx context.Context, example *sarfya.Example, input Input) (*FitResult, error) {
	dictOnce.Do(func() {
		dictionary = litxap.MultiDictionary{
			litxapfwew.Global(),
			litxapfwew.MultiWordPartDictionary(),
			&litxap.NumberDictionary{},
		}
	})

	var resLines []litxap.Line
	for _, line := range strings.Split(strings.Trim(example.Text.RawText(), "\n"), "\n") {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		dictionary := dictionary
		if len(input.CustomWords) > 0 {
			dictionary = litxap.MultiDictionary{dictionary, litxap.CustomWords(input.CustomWords, "")}
		}

		resLine, err := litxap.RunLine(line, dictionary)
		if err != nil {
			return nil, err
		}
		resLines = append(resLines, resLine)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	fitRes := Fit(example.Text, resLines, true, input.Selections)
	return &fitRes, nil
}
