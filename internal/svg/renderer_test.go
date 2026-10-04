package svg

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/kirasync2748/github-widget/internal/github"
)

func TestRenderRepositoryCard(t *testing.T) {
	repo := github.Repository{
		Owner:       "octocat",
		Name:        "hello-world",
		Description: "My first repository on GitHub!",
		Stars:       1500,
		Forks:       120,
		OpenIssues:  24,
		PrimaryLang: "Go",
		License:     "MIT",
		UpdatedAt:   time.Now().Add(-48 * time.Hour),
		HTMLURL:     "https://github.com/octocat/hello-world",
	}

	langs := []github.ComputedLanguage{
		{Name: "Go", Color: "#00ADD8", Percentage: 72.0},
		{Name: "HTML", Color: "#e34c26", Percentage: 18.0},
		{Name: "Shell", Color: "#89e051", Percentage: 10.0},
	}

	theme := Themes[DefaultTheme]
	opts := RenderOptions{Width: 540, Height: 200, Theme: theme}

	svgBytes := RenderRepositoryCard(repo, langs, opts)
	svgStr := string(svgBytes)

	// Must be valid XML.
	var v struct{}
	if err := xml.Unmarshal(svgBytes, &v); err != nil {
		t.Fatalf("SVG is not valid XML: %v\nSVG: %s", err, svgStr)
	}

	// Must contain expected content.
	if !strings.Contains(svgStr, "octocat") {
		t.Error("SVG missing owner name")
	}
	if !strings.Contains(svgStr, "hello-world") {
		t.Error("SVG missing repo name")
	}
	if !strings.Contains(svgStr, "Go") {
		t.Error("SVG missing language name")
	}
	if !strings.Contains(svgStr, "#00ADD8") {
		t.Error("SVG missing Go language color")
	}
	if !strings.Contains(svgStr, "1.5k") {
		t.Error("SVG missing formatted star count")
	}

	// Must not contain scripts.
	if strings.Contains(svgStr, "<script") {
		t.Error("SVG contains script tag")
	}
	if strings.Contains(svgStr, "foreignObject") {
		t.Error("SVG contains foreignObject")
	}
}

func TestRenderRepositoryCardEscapesMalicious(t *testing.T) {
	repo := github.Repository{
		Owner:       `<script>x</script>`,
		Name:        `"><rect width="99999"/>`,
		Description: `<img src=x onerror=alert(1)>`,
		Stars:       0,
		Forks:       0,
		OpenIssues:  0,
		PrimaryLang: "",
		UpdatedAt:   time.Time{},
	}

	theme := Themes["dark"]
	svgBytes := RenderRepositoryCard(repo, nil, RenderOptions{Theme: theme})
	svgStr := string(svgBytes)

	// Must be valid XML.
	var v struct{}
	if err := xml.Unmarshal(svgBytes, &v); err != nil {
		t.Fatalf("SVG with malicious input is not valid XML: %v", err)
	}

	// Must not contain raw script tags.
	if strings.Contains(svgStr, "<script") {
		t.Error("SVG contains unescaped script tag")
	}
}

func TestRenderRepositoryCardDimensions(t *testing.T) {
	repo := github.Repository{Owner: "a", Name: "b"}
	theme := Themes["dark"]

	// Too small → clamped.
	svgBytes := RenderRepositoryCard(repo, nil, RenderOptions{Width: 10, Height: 10, Theme: theme})
	if !strings.Contains(string(svgBytes), `width="540"`) {
		t.Error("width not clamped to default")
	}

	// Too large → clamped.
	svgBytes = RenderRepositoryCard(repo, nil, RenderOptions{Width: 99999, Height: 99999, Theme: theme})
	if !strings.Contains(string(svgBytes), `width="1200"`) {
		t.Error("width not clamped to max")
	}
}

func TestRenderRepositoryCardNoLanguages(t *testing.T) {
	repo := github.Repository{Owner: "a", Name: "b"}
	theme := Themes["dark"]
	svgBytes := RenderRepositoryCard(repo, nil, RenderOptions{Theme: theme})

	var v struct{}
	if err := xml.Unmarshal(svgBytes, &v); err != nil {
		t.Fatalf("SVG not valid XML when no languages: %v", err)
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"seconds", now.Add(-30 * time.Second), "just now"},
		{"one minute", now.Add(-90 * time.Second), "1 min ago"},
		{"minutes", now.Add(-5 * time.Minute), "5 mins ago"},
		{"one hour", now.Add(-90 * time.Minute), "1 hour ago"},
		{"hours", now.Add(-5 * time.Hour), "5 hours ago"},
		{"one day", now.Add(-30 * time.Hour), "1 day ago"},
		{"days", now.Add(-12 * 24 * time.Hour), "12 days ago"},
		{"one month", now.Add(-45 * 24 * time.Hour), "1 month ago"},
		{"months", now.Add(-300 * 24 * time.Hour), "10 months ago"},
		{"one year", now.Add(-400 * 24 * time.Hour), "1 year ago"},
		{"years", now.Add(-800 * 24 * time.Hour), "2 years ago"},
	}
	for _, tc := range cases {
		if got := relativeTime(tc.at); got != tc.want {
			t.Errorf("%s: relativeTime = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestRenderRepositoryCardThemes(t *testing.T) {
	repo := github.Repository{Owner: "a", Name: "b", Stars: 1}
	for name, theme := range Themes {
		svgBytes := RenderRepositoryCard(repo, nil, RenderOptions{Theme: theme})
		var v struct{}
		if err := xml.Unmarshal(svgBytes, &v); err != nil {
			t.Errorf("SVG with theme %s is not valid XML: %v", name, err)
		}
	}
}
