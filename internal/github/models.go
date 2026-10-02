package github

import "time"

// Repository represents the domain-level repository data needed to render a widget.
// It is deliberately decoupled from the GitHub API JSON shape.
type Repository struct {
	Owner       string
	OwnerAvatar string
	Name        string
	Description string
	Stars       int
	Forks       int
	OpenIssues  int
	PrimaryLang string
	License     string
	UpdatedAt   time.Time
	HTMLURL     string
}

// LanguageStat represents a single language and its byte count.
type LanguageStat struct {
	Name  string
	Bytes int64
}

// LanguageBreakdown holds a computed language distribution with percentages.
type LanguageBreakdown struct {
	Languages []ComputedLanguage
}

// ComputedLanguage is a language with a percentage of total bytes.
type ComputedLanguage struct {
	Name       string
	Color      string
	Percentage float64
}
