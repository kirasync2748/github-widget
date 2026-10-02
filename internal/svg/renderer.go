package svg

import (
	"fmt"
	"strings"
	"time"

	"github.com/kirasync2748/github-widget/internal/github"
)

// RenderOptions controls the SVG output dimensions and theme.
type RenderOptions struct {
	Width  int
	Height int
	Theme  Theme
}

// DefaultWidth and DefaultHeight are the default card dimensions.
const (
	DefaultWidth  = 540
	DefaultHeight = 200
	MinWidth      = 400
	MaxWidth      = 1200
	MinHeight     = 160
	MaxHeight     = 600
)

// RenderRepositoryCard generates a self-contained SVG card from repository data.
// The output is valid, safe XML with no scripts, external resources, or foreignObject.
func RenderRepositoryCard(data github.Repository, langs []github.ComputedLanguage, opts RenderOptions) []byte {
	if opts.Width < MinWidth {
		opts.Width = DefaultWidth
	}
	if opts.Width > MaxWidth {
		opts.Width = MaxWidth
	}
	if opts.Height < MinHeight {
		opts.Height = DefaultHeight
	}
	if opts.Height > MaxHeight {
		opts.Height = MaxHeight
	}

	t := opts.Theme
	w := opts.Width
	h := opts.Height

	// --- Layout constants ---
	padX := 28
	titleY := 36
	subTitleY := 64
	descY := 92
	statsY := 130
	langLabelY := 160
	langBarY := 172
	langBarH := 8
	langTextY := 192
	langBarW := w - padX*2

	// --- Language legend: always a single row (see langlabels.go) ---
	var legend langLegend
	if len(langs) > 0 {
		legend = layoutLangLegend(langs, padX, langBarW)

		// Grow card height to fit the legend row.
		if needed := langTextY + 12; needed > h {
			h = min(needed, MaxHeight)
		}
	}

	var b strings.Builder
	b.Grow(8192)

	// SVG root
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, h, w, h)
	b.WriteByte('\n')

	// Defs: rounded clip for the card
	fmt.Fprintf(&b, `<defs><clipPath id="card"><rect width="%d" height="%d" rx="12" ry="12"/></clipPath></defs>`, w, h)
	b.WriteByte('\n')

	// Background
	fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="12" ry="12" fill="%s"/>`, w, h, t.Background)
	b.WriteByte('\n')
	fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="12" ry="12" fill="none" stroke="%s" stroke-width="1"/>`, w, h, t.Border)
	b.WriteByte('\n')

	// --- Owner avatar (profile image or fallback circle) ---
	avatarR := 10
	avatarCX := padX + avatarR
	avatarCY := titleY - 3
	avatarSize := avatarR * 2
	if data.OwnerAvatar != "" {
		fmt.Fprintf(&b, `<clipPath id="avatar-clip"><circle cx="%d" cy="%d" r="%d"/></clipPath>`, avatarCX, avatarCY, avatarR)
		fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%d" fill="%s"/>`, avatarCX, avatarCY, avatarR, t.Surface)
		fmt.Fprintf(&b, `<image href="%s" x="%d" y="%d" width="%d" height="%d" clip-path="url(#avatar-clip)"/>`,
			EscapeXML(data.OwnerAvatar), avatarCX-avatarR, avatarCY-avatarR, avatarSize, avatarSize)
	} else {
		fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%d" fill="%s"/>`, avatarCX, avatarCY, avatarR, t.Surface)
	}
	fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="%d" fill="none" stroke="%s" stroke-width="1"/>`, avatarCX, avatarCY, avatarR, t.Border)
	b.WriteByte('\n')

	// --- Title: owner / repo name ---
	titleText := Truncate(EscapeXML(data.Owner)+" / "+EscapeXML(data.Name), 50)
	fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="16" font-weight="600" fill="%s">`+titleText+`</text>`, padX+28, titleY, t.PrimaryText)
	b.WriteByte('\n')

	// --- Subtitle: primary language + license ---
	subParts := []string{}
	if data.PrimaryLang != "" {
		subParts = append(subParts, EscapeXML(data.PrimaryLang))
	}
	if data.License != "" {
		subParts = append(subParts, EscapeXML(data.License))
	}
	if len(subParts) > 0 {
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="12" fill="%s">`+strings.Join(subParts, " • ")+`</text>`, padX, subTitleY, t.SecondaryText)
		b.WriteByte('\n')
	}

	// --- Description ---
	if data.Description != "" {
		desc := Truncate(EscapeXML(data.Description), 72)
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="13" fill="%s">`+desc+`</text>`, padX, descY, t.SecondaryText)
		b.WriteByte('\n')
	}

	// --- Stats row: stars, forks, issues ---
	stats := buildStats(data, t)
	statsX := padX
	iconSize := 12
	for _, s := range stats {
		drawIcon(&b, s.icon, statsX, statsY-6, iconSize, s.color)
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="12" fill="%s">%s</text>`,
			statsX+iconSize+4, statsY+4, t.SecondaryText, s.text)
		statsX += s.width
	}
	b.WriteByte('\n')

	// --- Last updated ---
	if !data.UpdatedAt.IsZero() {
		updated := relativeTime(data.UpdatedAt)
		updatedX := w - padX
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="11" fill="%s" text-anchor="end">Updated %s</text>`, updatedX, statsY+4, t.MutedText, EscapeXML(updated))
		b.WriteByte('\n')
	}

	// --- Languages section ---
	if len(langs) > 0 {
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="11" font-weight="600" fill="%s">Languages</text>`, padX, langLabelY, t.SecondaryText)
		b.WriteByte('\n')

		// Progress bar background
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" rx="4" ry="4" fill="%s"/>`, padX, langBarY, langBarW, langBarH, t.ProgressBg)
		b.WriteByte('\n')

		// Language segments
		segX := padX
		for _, lang := range langs {
			segW := int(lang.Percentage / 100 * float64(langBarW))
			if segW < 1 {
				segW = 1
			}
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`, segX, langBarY, segW, langBarH, lang.Color)
			segX += segW
		}

		// Language legend (one row, smaller text)
		r := legend.dotRadius
		for _, lp := range legend.labels {
			fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s"/>`, lp.x+r, float64(langTextY)-r, r, lp.color)
			fmt.Fprintf(&b, `<text x="%.1f" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="%g" fill="%s">%s</text>`,
				lp.x+legend.textOffset, langTextY, legend.fontSize, t.SecondaryText, EscapeXML(lp.text))
		}
		b.WriteByte('\n')
	} else {
		// No language data available
		fmt.Fprintf(&b, `<text x="%d" y="%d" font-family="Segoe UI,Arial,Helvetica,sans-serif" font-size="11" fill="%s">No language data</text>`, padX, langLabelY, t.MutedText)
		b.WriteByte('\n')
	}

	b.WriteString("</svg>")
	return []byte(b.String())
}

// statItem is a single stat badge.
type statItem struct {
	icon  string
	text  string
	color string
	width int
}

// Stat icon colors — distinct per stat, consistent across all themes.
const (
	starColor  = "#E3B341" // gold
	forkColor  = "#58A6FF" // blue
	issueColor = "#3FB950" // green
)

func buildStats(data github.Repository, _ Theme) []statItem {
	return []statItem{
		{icon: "star", text: formatNumber(data.Stars), color: starColor, width: 55},
		{icon: "fork", text: fmt.Sprintf("%s Forks", formatNumber(data.Forks)), color: forkColor, width: 95},
		{icon: "issue", text: fmt.Sprintf("%s Issues", formatNumber(data.OpenIssues)), color: issueColor, width: 105},
	}
}

// formatNumber converts an int to a human-readable string (1.2k, 3.4M).
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
}

// relativeTime returns a human-readable relative time string.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return "just now"
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	default:
		return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
	}
}
