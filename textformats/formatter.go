package textformats

import (
	"strings"
	"unicode/utf8"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/emphasis"
)

type Formatter interface {
	StartQuote(w *strings.Builder, example sarfya.FilterMatch)
	EndQuote(w *strings.Builder, example sarfya.FilterMatch)

	LineTags(navi bool, last bool) (string, string)
	TextTags(navi bool) (string, string)
	SelectionTags() (string, string)
	AdjacentTags() (string, string)
	StressTags() (string, string)
}

func Generate(formatter Formatter, example sarfya.FilterMatch, lang string, stressFit *emphasis.FitResult) string {
	if lang == "" {
		lang = "en" // All hail our silly global hegemon.
	}
	w := &strings.Builder{}
	w.Grow(1024)

	formatter.StartQuote(w, example)
	writeText(formatter, true, w, example.Text, example.Spans, example.TranslationAdjacent[lang], stressFit)
	if len(example.Translations[lang]) > 0 {
		writeText(formatter, false, w, example.Translations[lang], example.TranslationSpans[lang], nil, nil)
	}
	formatter.EndQuote(w, example)

	return w.String()
}

func writeText(formatter Formatter, navi bool, w *strings.Builder, sentence sarfya.Sentence, spans [][]int, adjacentSpans [][]int, stressFit *emphasis.FitResult) {
	inSpan := make(map[int]bool)
	inAdjacent := make(map[int]bool)
	for _, span := range spans {
		for _, index := range span {
			inSpan[index] = true
		}
	}
	for _, span := range adjacentSpans {
		for _, index := range span {
			inAdjacent[index] = true
		}
	}

	textOpen, textClose := formatter.TextTags(navi)
	lineOpen, lineClose := formatter.LineTags(navi, false)
	_, lastLineClose := formatter.LineTags(navi, true)

	selectedOpen, selectedClose := formatter.SelectionTags()
	adjacentOpen, adjacentClose := formatter.AdjacentTags()

	w.WriteString(textOpen)
	w.WriteString(lineOpen)
	for i, part := range sentence {
		if part.Newline {
			w.WriteString(lineClose)
			w.WriteString(lineOpen)
			part.Newline = false
		}

		if inSpan[i] {
			w.WriteString(selectedOpen)
		} else if inAdjacent[i] {
			w.WriteString(adjacentOpen)
		}

		writeStressedPart(formatter, w, stressFit, i, part.RawText())

		if inSpan[i] {
			w.WriteString(selectedClose)
		} else if inAdjacent[i] {
			w.WriteString(adjacentClose)
		}
	}
	w.WriteString(lastLineClose)
	w.WriteString(textClose)
}

func writeStressedPart(formatter Formatter, w *strings.Builder, stress *emphasis.FitResult, partIndex int, rawText string) {
	stressOpen, stressClose := formatter.StressTags()

	if stress != nil && len(stress.Underlinings[partIndex]) > 0 {
		var currBytes, iOffset int
		for _, pos := range stress.Underlinings[partIndex] {
			// Convert rune offsets to byte offsets
			pos0 := currBytes
			var pos1 int
			for i := iOffset; i < pos[0]+pos[1]; i += 1 {
				if i == pos[0] {
					pos1 = currBytes
				}
				_, size := utf8.DecodeRuneInString(rawText[currBytes:])
				currBytes += size
			}

			// Multiple stresses are not stored relative to each other
			iOffset = pos[0] + pos[1]

			// Unressed part, if any
			if pos0 < pos1 {
				w.WriteString(rawText[pos0:pos1])
			}

			// Underline the stressed part
			w.WriteString(stressOpen)
			w.WriteString(rawText[pos1:currBytes])
			w.WriteString(stressClose)
		}

		// Write the remaining unstressed text
		if currBytes != len(rawText) {
			w.WriteString(rawText[currBytes:])
		}
	} else {
		w.WriteString(rawText)
	}
}
