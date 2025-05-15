package emphasis

import (
	"maps"
	"slices"
)

type Input struct {
	ID          string        `json:"id,omitempty" yaml:"id,omitempty"`
	Selections  []map[int]int `json:"selections,omitempty" yaml:"selections,omitempty"`
	CustomWords []string      `json:"customWords,omitempty" yaml:"custom_words,omitempty"`
}

func (input *Input) Copy() *Input {
	var selections []map[int]int
	for _, selectionMap := range input.Selections {
		selections = append(selections, maps.Clone(selectionMap))
	}

	return &Input{
		ID:          input.ID,
		Selections:  selections,
		CustomWords: slices.Clone(input.CustomWords),
	}
}

type Data struct {
	EmphasizedParts map[int][2]int
}

type SentencePart struct {
	PartIndex      int `json:"pi" yaml:"pi"`
	SyllableStart  int `json:"ss" yaml:"ss"`
	SyllableLength int `json:"sl" yaml:"sl"`
}
