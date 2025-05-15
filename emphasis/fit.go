package emphasis

import (
	"github.com/gissleh/litxap"
	"github.com/gissleh/sarfya"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

type FitResult struct {
	Underlinings map[int][][2]int       `json:"underlinings" yaml:"underlinings"`
	Ambiguities  []FitResultAmbiguity   `json:"ambiguities,omitempty" yaml:"ambiguities"`
	MissingParts []FitResultMissingPart `json:"missingParts,omitempty" yaml:"missing_parts"`
}

type FitResultAmbiguity struct {
	LineIndex int                    `json:"li" yaml:"li"`
	PartIndex int                    `json:"pi" yaml:"pi"`
	Matches   []litxap.LinePartMatch `json:"matches" yaml:"matches"`
}

type FitResultMissingPart struct {
	LineIndex int `json:"li" yaml:"li"`
	PartIndex int `json:"pi" yaml:"pi"`
}

func Fit(sentence sarfya.Sentence, lines []litxap.Line, countRunes bool, selections []map[int]int) FitResult {
	positions := make([]int, 0, len(sentence))
	rawSB := strings.Builder{}
	rawSB.Grow(len(sentence) * 8)
	for _, part := range sentence {
		part.Newline = false // hax since we're not dealing with newline
		if countRunes {
			positions = append(positions, utf8.RuneCountInString(rawSB.String()))
		} else {
			positions = append(positions, rawSB.Len())
		}
		_ = part.WriteRawTo(&rawSB)
	}

	raw := rawSB.String()
	initialLen := len(raw)
	if countRunes {
		initialLen = utf8.RuneCountInString(raw)
	}

	res := FitResult{
		Underlinings: make(map[int][][2]int, len(sentence)),
		Ambiguities:  make([]FitResultAmbiguity, 0),
		MissingParts: make([]FitResultMissingPart, 0),
	}

	for i, line := range lines {
		var selectionsI map[int]int
		if i < len(selections) {
			selectionsI = selections[i]
		}

		for j, part := range line {
			if len(part.Matches) == 0 {
				if part.IsWord {
					res.MissingParts = append(res.MissingParts, FitResultMissingPart{
						LineIndex: i,
						PartIndex: j,
					})
				}

				//log.Println(part.Raw, "<!<", raw)
				raw = strings.TrimPrefix(raw, part.Raw)
			} else {
				selection, hasSelection := selectionsI[j]
				if !hasSelection && len(part.Matches) != 1 {
					firstStress := part.Matches[0].Stress
					for _, other := range part.Matches {
						if other.Stress != firstStress {
							res.Ambiguities = append(res.Ambiguities, FitResultAmbiguity{
								LineIndex: i,
								PartIndex: j,
								Matches:   slices.Clone(part.Matches),
							})

							break
						}
					}
				}

				match := part.Matches[selection]
				for k, syllable := range match.Syllables {
					if k == match.Stress && len(match.Syllables) > 1 {
						currentPos := initialLen - len(raw)
						if countRunes {
							currentPos = initialLen - utf8.RuneCountInString(raw)
						}

						partIndex, found := sort.Find(len(positions), func(i int) int {
							return currentPos - positions[i]
						})
						if !found {
							partIndex -= 1
						}

						syllableLength := len(syllable)
						if countRunes {
							syllableLength = utf8.RuneCountInString(syllable)
						}

						res.Underlinings[partIndex] = append(res.Underlinings[partIndex], [2]int{currentPos - positions[partIndex], syllableLength})
						//log.Println(match.Entry.Word, sentence[partIndex].RawText(), res.Underlinings[partIndex])
					}

					//log.Println(syllable, "<+<", raw)
					raw = strings.TrimPrefix(raw, syllable)
				}
			}
		}
	}

	return res
}
