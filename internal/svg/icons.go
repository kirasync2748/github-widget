package svg

import (
	"fmt"
	"strings"
)

// Icon path data for stat icons. All paths use a 16×16 coordinate system,
// matching the geometry of GitHub's Octicons.
const (
	starPath = "M8 .25a.75.75 0 01.673.418l1.882 3.815 4.21.612a.75.75 0 01" +
		".416 1.279l-3.046 2.97.719 4.192a.75.75 0 01-1.088.791L8 12.347l-3." +
		"766 1.98a.75.75 0 01-1.088-.79l.72-4.194L.818 6.374a.75.75 0 01.41" +
		"6-1.28l4.21-.611L7.327.668A.75.75 0 018 .25z"

	forkPath = "M5 5.372v.878c0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75v-.878" +
		"a2.25 2.25 0 111.5 0v.878a2.25 2.25 0 01-2.25 2.25h-1.5v2.128a2.251 2" +
		".251 0 11-1.5 0V8.5h-1.5A2.25 2.25 0 013.5 6.25v-.878a2.25 2.25 0 111" +
		".5 0zM5 3.25a.75.75 0 10-1.5 0 .75.75 0 001.5 0zm6.75.75a.75.75 0 100" +
		"-1.5.75.75 0 000 1.5zm-3 8.75a.75.75 0 10-1.5 0 .75.75 0 001.5 0z"
)

// drawIcon renders an SVG icon at (x, y) with the given pixel size and color.
// Icons are defined in a 16×16 coordinate system and scaled to the requested size.
func drawIcon(b *strings.Builder, name string, x, y, size int, color string) {
	scale := float64(size) / 16.0
	switch name {
	case "star":
		fmt.Fprintf(b,
			`<g transform="translate(%d,%d) scale(%.4f)" fill="%s"><path d="%s"/></g>`,
			x, y, scale, color, starPath)
	case "fork":
		fmt.Fprintf(b,
			`<g transform="translate(%d,%d) scale(%.4f)" fill="%s"><path d="%s"/></g>`,
			x, y, scale, color, forkPath)
	case "issue":
		fmt.Fprintf(b, `<g transform="translate(%d,%d) scale(%.4f)">`, x, y, scale)
		fmt.Fprintf(b, `<circle cx="8" cy="8" r="6.25" fill="none" stroke="%s" stroke-width="1.5"/>`, color)
		fmt.Fprintf(b, `<circle cx="8" cy="8" r="2" fill="%s"/>`, color)
		b.WriteString(`</g>`)
	}
}
