package formatutils

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/emphasis"
)

func DiscordQuote(example sarfya.FilterMatch, lang string, stressFit *emphasis.FitResult) string {
	if lang == "" {
		lang = "en" // All hail our silly global hegemon.
	}
	w := &strings.Builder{}
	w.Grow(1024)

	domain := ""
	if u, err := url.Parse(example.Source.URL); err == nil {
		domain = u.Hostname()
	}

	w.WriteString("> ")
	writeDiscordLine(w, example.Text, example.Spans, example.TranslationAdjacent[lang], stressFit)
	if len(example.Translations[lang]) > 0 {
		w.WriteString("\n> -# ")
		writeDiscordLine(w, example.Translations[lang], example.TranslationSpans[lang], nil, nil)
	}
	w.WriteString("\n> -# — [")
	w.WriteString(example.Source.Title)
	w.WriteString("](")
	w.WriteString(example.Source.URL)
	w.WriteString(") (")
	if domain != "" {
		w.WriteString(domain)
		w.WriteString(", ")
	}
	w.WriteString(example.Source.Date)
	w.WriteString(")")

	return w.String()
}

func writeDiscordLine(w *strings.Builder, sentence sarfya.Sentence, spans [][]int, adjacentSpans [][]int, stressFit *emphasis.FitResult) {
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

	for i, part := range sentence {
		if part.Newline {
			_, _ = w.WriteRune('\n')
		}

		if inSpan[i] {
			w.WriteString("**")
		} else if inAdjacent[i] {
			w.WriteString("*")
		}

		writeStressed(w, stressFit, i, part.RawText())

		if inSpan[i] {
			w.WriteString("**")
		} else if inAdjacent[i] {
			w.WriteString("*")
		}
	}
}

func writeStressed(w *strings.Builder, stress *emphasis.FitResult, partIndex int, rawText string) {
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
			w.WriteString("__")
			w.WriteString(rawText[pos1:currBytes])
			w.WriteString("__")
		}

		// Write the remaining unstressed text
		if currBytes != len(rawText) {
			w.WriteString(rawText[currBytes:])
		}
	} else {
		w.WriteString(rawText)
	}
}
