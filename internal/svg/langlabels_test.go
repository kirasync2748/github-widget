package svg

import (
	"strings"
	"testing"

	"github.com/kirasync2748/github-widget/internal/github"
)

const legendX = 28

var fiveLangs = []github.ComputedLanguage{
	{Name: "TypeScript", Color: "#3178c6", Percentage: 60.4},
	{Name: "Python", Color: "#3572A5", Percentage: 30.2},
	{Name: "Dockerfile", Color: "#384d54", Percentage: 2.4},
	{Name: "Shell", Color: "#89e051", Percentage: 2.4},
	{Name: "JavaScript", Color: "#f1e05a", Percentage: 1.8},
}

func assertFitsOneRow(t *testing.T, lg langLegend, barW int) {
	t.Helper()
	last := lg.labels[len(lg.labels)-1]
	if end := last.x + itemWidth(last.text, lg.fontSize); end > float64(legendX+barW) {
		t.Errorf("legend overflows: ends at %.0f, bar ends at %d", end, legendX+barW)
	}
}

func TestLayoutLangLegendFullNamesWhenTheyFit(t *testing.T) {
	barW := DefaultWidth - 2*legendX
	lg := layoutLangLegend(fiveLangs, legendX, barW)

	if len(lg.labels) != 5 {
		t.Fatalf("got %d labels, want 5", len(lg.labels))
	}
	for i, l := range lg.labels {
		if !strings.HasPrefix(l.text, fiveLangs[i].Name+" ") {
			t.Errorf("label %d = %q, want full name %q", i, l.text, fiveLangs[i].Name)
		}
	}
	if lg.fontSize != 10 {
		t.Errorf("fontSize = %g, want 10", lg.fontSize)
	}
	assertFitsOneRow(t, lg, barW)
}

func TestLayoutLangLegendShortensLongNamesToSeven(t *testing.T) {
	langs := []github.ComputedLanguage{
		{Name: "Jupyter Notebook", Percentage: 60.4},
		{Name: "Python", Percentage: 30.2},
		{Name: "Dockerfile", Percentage: 2.4},
		{Name: "Shell", Percentage: 2.4},
		{Name: "Objective-C++", Percentage: 1.8},
	}
	barW := DefaultWidth - 2*legendX
	lg := layoutLangLegend(langs, legendX, barW)

	if got := lg.labels[0].text; got != "Jupyter… 60.4%" {
		t.Errorf("first label = %q, want %q", got, "Jupyter… 60.4%")
	}
	if got := lg.labels[1].text; got != "Python 30.2%" {
		t.Errorf("short name changed: %q", got)
	}
	if lg.fontSize != 10 {
		t.Errorf("fontSize = %g, want 10", lg.fontSize)
	}
	assertFitsOneRow(t, lg, barW)
}

func TestLayoutLangLegendNarrowCardStaysOneRow(t *testing.T) {
	barW := MinWidth - 2*legendX
	lg := layoutLangLegend(fiveLangs, legendX, barW)

	if lg.fontSize >= 10 {
		t.Errorf("fontSize = %g, want smaller than 10 on a narrow card", lg.fontSize)
	}
	if !strings.Contains(lg.labels[0].text, "…") {
		t.Errorf("first label = %q, want it shortened", lg.labels[0].text)
	}
	assertFitsOneRow(t, lg, barW)
}

func TestShortenName(t *testing.T) {
	tests := []struct {
		in   string
		n    int
		want string
	}{
		{"TypeScript", 7, "TypeScr…"},
		{"Python", 7, "Python"},
		{"Go", 0, "Go"},
	}
	for _, tt := range tests {
		if got := shortenName(tt.in, tt.n); got != tt.want {
			t.Errorf("shortenName(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
		}
	}
}
