package textformats

import (
	"net/url"
	"strings"

	"github.com/gissleh/sarfya"
)

func DiscordFormatter() Formatter {
	return &discordFormatter{}
}

type discordFormatter struct{}

func (d *discordFormatter) StartQuote(w *strings.Builder, example sarfya.FilterMatch) {}

func (d *discordFormatter) EndQuote(w *strings.Builder, example sarfya.FilterMatch) {
	w.WriteString("> -# — [")
	w.WriteString(example.Source.Title)
	w.WriteString("](")
	w.WriteString(example.Source.URL)
	w.WriteString(") (")
	if u, err := url.Parse(example.Source.URL); err == nil {
		w.WriteString(u.Hostname())
		w.WriteString(", ")
	}
	w.WriteString(example.Source.Date)
	w.WriteString(")")
}

func (d *discordFormatter) SelectionTags() (string, string) {
	return "**", "**"
}

func (d *discordFormatter) StressTags() (string, string) {
	return "__", "__"
}

func (d *discordFormatter) AdjacentTags() (string, string) {
	return "**", "**"
}

func (d *discordFormatter) TextTags(_ bool) (string, string) {
	return "", ""
}

func (d *discordFormatter) LineTags(navi bool, _ bool) (string, string) {
	if navi {
		return "> ", "\n"
	}

	return "> -# ", "\n"
}
