package svg

import (
	"strings"
)

// EscapeXML escapes a string for safe insertion into XML/SVG text content
// and attribute values. It prevents SVG/XML/HTML injection.
func EscapeXML(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			if r < 0x20 && r != '\n' && r != '\t' && r != '\r' {
				// Skip other control characters that are invalid in XML.
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Truncate shortens s to at most maxLen runes, appending an ellipsis if truncated.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-1]) + "…"
}
