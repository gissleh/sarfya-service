package emphasis

import (
	"context"
	"github.com/gissleh/sarfya"
)

func Run(ctx context.Context, client *LitxapClient, example *sarfya.Example, input Input) (*FitResult, error) {
	litxapRes, err := client.Call(ctx, example.Text.RawText(), input.CustomWords)
	if err != nil {
		return nil, err
	}

	fitRes := Fit(example.Text, litxapRes, true, input.Selections)

	return &fitRes, nil
}
