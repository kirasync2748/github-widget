package svg

import (
	"fmt"
	"strings"

	"github.com/kirasync2748/github-widget/internal/github"
)

// legendStep is one attempt at fitting the language legend on a single row:
// a name length limit (0 = full names) and a font size in px.
type legendStep struct {
	limit    int
	fontSize float64
}

// legendSteps are tried in order until the legend fits on one row:
// full names first, then names shortened to 7 characters ("TypeScr…"),
// then a slightly smaller font, then shorter names for very narrow cards.
var legendSteps = []legendStep{
	{0, 10}, {7, 10}, {7, 9}, {7, 8}, {6, 8}, {5, 8}, {4, 8}, {3, 8},
}

// langLegend is the laid-out single-row language legend.
type langLegend struct {
	fontSize   float64
	dotRadius  float64
	textOffset float64 // text starts this far right of the item's x
	labels     []langLabel
}

// langLabel is one positioned legend item: dot + "Name 12.3%".
type langLabel struct {
	x     float64
	text  string // unescaped; escape when writing
	color string
}

// layoutLangLegend places every language on one row within maxW pixels.
func layoutLangLegend(langs []github.ComputedLanguage, startX, maxW int) langLegend {
	var step legendStep
	var texts []string
	for _, step = range legendSteps {
		texts = langTexts(langs, step.limit)
		if rowWidth(texts, step.fontSize) <= float64(maxW) {
			break
		}
	}

	fs := step.fontSize
	legend := langLegend{fontSize: fs, dotRadius: fs * 0.35, textOffset: fs}
	x := float64(startX)
	for i, text := range texts {
		legend.labels = append(legend.labels, langLabel{x: x, text: text, color: langs[i].Color})
		x += itemWidth(text, fs) + itemGap(fs)
	}
	return legend
}

func langTexts(langs []github.ComputedLanguage, limit int) []string {
	texts := make([]string, len(langs))
	for i, l := range langs {
		texts[i] = fmt.Sprintf("%s %.1f%%", shortenName(l.Name, limit), l.Percentage)
	}
	return texts
}

// shortenName keeps the first n runes and appends "…" when name is longer.
func shortenName(name string, n int) string {
	r := []rune(name)
	if n <= 0 || len(r) <= n {
		return name
	}
	return string(r[:n]) + "…"
}

func rowWidth(texts []string, fontSize float64) float64 {
	w := 0.0
	for _, t := range texts {
		w += itemWidth(t, fontSize)
	}
	if len(texts) > 1 {
		w += itemGap(fontSize) * float64(len(texts)-1)
	}
	return w
}

// itemWidth is the width of one legend item (dot + text).
func itemWidth(text string, fontSize float64) float64 {
	return fontSize + textWidth(text, fontSize)
}

// itemGap is the space between two legend items.
func itemGap(fontSize float64) float64 {
	return fontSize * 1.4
}

// textWidth estimates rendered width for common sans-serif fonts
// (Segoe UI / Arial / Roboto). Slightly generous so text never overflows.
func textWidth(s string, fontSize float64) float64 {
	em := 0.0
	for _, r := range s {
		switch {
		case strings.ContainsRune("ijlI.,:;!|'", r):
			em += 0.25
		case strings.ContainsRune("frt -()", r):
			em += 0.32
		case strings.ContainsRune("mwMW%", r):
			em += 0.86
		case r == '…':
			em += 1.0
		case r >= 'A' && r <= 'Z':
			em += 0.68
		default:
			em += 0.56
		}
	}
	return em * fontSize
}
