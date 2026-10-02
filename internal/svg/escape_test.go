package svg

import (
	"strings"
	"testing"
)

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"a & b", "a &amp; b"},
		{"<script>", "&lt;script&gt;"},
		{`"quoted"`, "&quot;quoted&quot;"},
		{"it's", "it&#39;s"},
	}
	for _, tt := range tests {
		got := EscapeXML(tt.input)
		if got != tt.want {
			t.Errorf("EscapeXML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEscapeXMLPreventsInjection(t *testing.T) {
	malicious := `<script>alert("xss")</script><rect width="99999"/>`
	escaped := EscapeXML(malicious)
	if strings.Contains(escaped, "<script>") || strings.Contains(escaped, "<rect") {
		t.Errorf("escape failed to neutralize markup: %s", escaped)
	}
}

func TestEscapeXMLEliminatesControlChars(t *testing.T) {
	input := "hello\x00\x01\x02world"
	got := EscapeXML(input)
	if strings.ContainsAny(got, "\x00\x01\x02") {
		t.Errorf("control characters survived: %q", got)
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("hello", 10); got != "hello" {
		t.Errorf("Truncate(hello,10) = %q", got)
	}
	if got := Truncate("hello world", 5); got != "hell…" {
		t.Errorf("Truncate(hello world,5) = %q, want 'hell…'", got)
	}
	if got := Truncate("abc", 0); got != "" {
		t.Errorf("Truncate(abc,0) = %q, want ''", got)
	}
}

func TestGetLanguageColor(t *testing.T) {
	if c := GetLanguageColor("Go"); c != "#00ADD8" {
		t.Errorf("Go color = %s, want #00ADD8", c)
	}
	if c := GetLanguageColor("Python"); c != "#3572A5" {
		t.Errorf("Python color = %s, want #3572A5", c)
	}
	if c := GetLanguageColor("Nonexistent Language"); c != FallbackColor {
		t.Errorf("unknown color = %s, want %s", c, FallbackColor)
	}
}

func TestGetTheme(t *testing.T) {
	_, ok := GetTheme("dark")
	if !ok {
		t.Error("dark theme not found")
	}
	_, ok = GetTheme("nonexistent")
	if ok {
		t.Error("nonexistent theme should not be found")
	}
}
