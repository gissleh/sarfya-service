package textformats

import (
	"strings"

	"github.com/gissleh/sarfya"
)

/*
[quote author=https://naviteri.org/2018/04/100a-liu-amip-64-new-words-part-2]
[u]Ra[/u]luri fwa [b]tì[u]fme[/u]tokit[/b] ke emzo[u]la[/u]'u lä[u]ngu[/u] fwìng a[u]txan[/u].
[size=1]Ralu's not passing [b]the test[/b] was a great humiliation to him.[/size]
[/quote]
*/

func BBCodeFormatter() Formatter {
	return &bbcodeFormatter{}
}

type bbcodeFormatter struct{}

func (f *bbcodeFormatter) StartQuote(w *strings.Builder, example sarfya.FilterMatch) {
	w.WriteString("[quote=")
	w.WriteString(example.Source.URL)
	w.WriteString("]\n")
}

func (f *bbcodeFormatter) EndQuote(w *strings.Builder, example sarfya.FilterMatch) {
	w.WriteString("[/quote]")
}

func (f *bbcodeFormatter) LineTags(navi bool, last bool) (string, string) {
	if navi {
		if last {
			return "[b]", "[/b]"
		}

		return "[b]", "[/b]\n"
	}

	if last {
		return "", ""
	}

	return "", "\n"
}

func (f *bbcodeFormatter) TextTags(navi bool) (string, string) {
	if navi {
		return "", "\n"
	}

	return "", "\n"
}

func (f *bbcodeFormatter) SelectionTags() (string, string) {
	return "[color=orange]", "[/color]"
}

func (f *bbcodeFormatter) AdjacentTags() (string, string) {
	return "[color=orange]", "[/color]"
}

func (f *bbcodeFormatter) StressTags() (string, string) {
	return "[u]", "[/u]"
}
