package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kirasync2748/github-widget/internal/github"
	"github.com/kirasync2748/github-widget/internal/service"
	"github.com/kirasync2748/github-widget/internal/svg"
)

// WidgetHandler serves the SVG widget endpoint.
type WidgetHandler struct {
	svc    *service.RepositoryService
	logger *slog.Logger
}

// NewWidgetHandler creates a new WidgetHandler.
func NewWidgetHandler(svc *service.RepositoryService, logger *slog.Logger) *WidgetHandler {
	return &WidgetHandler{svc: svc, logger: logger}
}

// validOwnerRepo validates that owner and repo are well-formed GitHub identifiers.
func validOwnerRepo(owner, repo string) bool {
	if owner == "" || repo == "" {
		return false
	}
	if len(owner) > 39 || len(repo) > 100 {
		return false
	}
	// GitHub usernames: alphanumeric, hyphens, and underscores; no leading/trailing hyphen.
	for _, r := range owner {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	// Repo names: alphanumeric, hyphens, underscores, dots.
	for _, r := range repo {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.') {
			return false
		}
	}
	if strings.HasPrefix(owner, "-") || strings.HasSuffix(owner, "-") {
		return false
	}
	return true
}

// ServeHTTP handles GET /api/widget/{owner}/{repo}.
func (h *WidgetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// Parse path: /api/widget/{owner}/{repo}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/widget/"), "/")
	if len(parts) != 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected /api/widget/{owner}/{repo}"})
		return
	}
	owner, repo := parts[0], parts[1]

	if !validOwnerRepo(owner, repo) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid owner or repository name"})
		return
	}

	// Parse query params.
	query := r.URL.Query()
	themeName := query.Get("theme")
	if themeName == "" {
		themeName = svg.DefaultTheme
	}
	theme, ok := svg.GetTheme(themeName)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unknown theme: " + themeName})
		return
	}

	opts := svg.RenderOptions{
		Width:  svg.DefaultWidth,
		Height: svg.DefaultHeight,
		Theme:  theme,
	}

	if w := query.Get("width"); w != "" {
		opts.Width = parseIntSafe(w, svg.DefaultWidth)
	}
	if h := query.Get("height"); h != "" {
		opts.Height = parseIntSafe(h, svg.DefaultHeight)
	}

	// Fetch data with timeout.
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	data, err := h.svc.GetWidgetData(ctx, owner, repo)
	if err != nil {
		h.writeError(w, err)
		return
	}

	svgBytes := svg.RenderRepositoryCard(data.Repository, data.Languages, opts)

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(svgBytes)
}

// writeError maps service errors to appropriate HTTP responses.
func (h *WidgetHandler) writeError(w http.ResponseWriter, err error) {
	if ue, ok := err.(*github.UpstreamError); ok {
		switch {
		case ue.IsNotFound():
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "repository not found"})
		case ue.IsRateLimited():
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "GitHub API rate limit exceeded"})
		case ue.Code >= 500:
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "GitHub upstream error"})
		default:
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "GitHub API error"})
		}
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

const requestTimeout = 15 * time.Second

func parseIntSafe(s string, def int) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		return def
	}
	return n
}
