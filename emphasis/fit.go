package emphasis

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/gissleh/litxap"
	"github.com/gissleh/sarfya"
)

var skipMissing = []string{
	"X", "Y", "A", "B", "C", "PAWL",
}

var skipWords = []string{
	"ìlä", "tsatseng",
}

var alwaysChoose = map[string]string{
	"frapo":    "frapo",
	"frapor":   "frapo",
	"frapol":   "frapo",
	"frapot":   "frapo",
	"frapoti":  "frapo",
	"fraporu":  "frapo",
	"frapori":  "frapo",
	"frapone":  "frapo",
	"frakrr":   "frakrr",
	"fratseng": "fratseng",
}

type FitResult struct {
	Underlinings map[int][][2]int       `json:"underlinings" yaml:"underlinings"`
	Ambiguities  []FitResultAmbiguity   `json:"ambiguities,omitempty" yaml:"ambiguities"`
	MissingParts []FitResultMissingPart `json:"missingParts,omitempty" yaml:"missing_parts"`
}

func (result FitResult) IsSafe() bool {
	return len(result.Ambiguities) == 0 && len(result.MissingParts) == 0
}

type FitResultAmbiguity struct {
	LineIndex     int                    `json:"li" yaml:"li"`
	LinePartIndex int                    `json:"lpi" yaml:"lpi"`
	Raw           string                 `json:"raw" yaml:"raw"`
	Matches       []litxap.LinePartMatch `json:"matches" yaml:"matches"`
}

type FitResultMissingPart struct {
	LineIndex     int    `json:"li" yaml:"li"`
	LinePartIndex int    `json:"lpi" yaml:"lpi"`
	Raw           string `json:"raw" yaml:"raw"`
}

func Fit(sentence sarfya.Sentence, lines []litxap.Line, countRunes bool, selections map[string]string) FitResult {
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
		for j, part := range line {
			if len(part.Matches) == 0 {
				if part.IsWord {
					if !slices.Contains(skipMissing, part.Raw) {
						res.MissingParts = append(res.MissingParts, FitResultMissingPart{
							LineIndex:     i,
							LinePartIndex: j,
							Raw:           part.Raw,
						})
					}
				}

				raw = strings.TrimPrefix(raw, part.Raw)
			} else {
				if slices.Contains(skipWords, strings.ToLower(part.Raw)) {
					raw = strings.TrimPrefix(raw, part.Raw)
					continue
				}

				selectionKey := fmt.Sprintf("%s[%d,%d]", part.Raw, i, j)
				selection, hasSelection := selections[selectionKey]
				if !hasSelection {
					selection, hasSelection = selections[part.Raw]
				}

				if !hasSelection && len(part.Matches) != 1 {
					if alwaysChoose, ok := alwaysChoose[strings.ToLower(part.Raw)]; ok {
						for k, match := range part.Matches {
							if match.Entry.Word == alwaysChoose {
								selection = fmt.Sprintf("*[%d]", k)
								hasSelection = true
								break
							}
						}
					}

					if !hasSelection {
						firstStress := part.Matches[0].Stress
						for _, other := range part.Matches {
							if other.Stress != firstStress {
								res.Ambiguities = append(res.Ambiguities, FitResultAmbiguity{
									LineIndex:     i,
									LinePartIndex: j,
									Raw:           part.Raw,
									Matches:       slices.Clone(part.Matches),
								})

								break
							}
						}
					}
				}

				selectionIndex := 0
				if hasSelection {
					selectionIndex = -1
					for k, match := range part.Matches {
						if selection == fmt.Sprintf("*[%d]", k) ||
							selection == fmt.Sprintf("%s[%d]", match.Entry.Word, k) ||
							selection == fmt.Sprintf("%s", match.Entry.Word) {
							selectionIndex = k
							break
						}
					}

					if selectionIndex == -1 {
						res.Ambiguities = append(res.Ambiguities, FitResultAmbiguity{
							LineIndex:     i,
							LinePartIndex: j,
							Raw:           part.Raw,
							Matches:       slices.Clone(part.Matches),
						})
						selectionIndex = 0
					}
				}

				match := part.Matches[selectionIndex]
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
					}

					raw = strings.TrimPrefix(raw, syllable)
				}
			}
		}
	}

	return res
}
