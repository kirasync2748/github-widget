package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kirasync2748/github-widget/internal/cache"
	githubclient "github.com/kirasync2748/github-widget/internal/github"
	"github.com/kirasync2748/github-widget/internal/service"
)

func newTestHandler(t *testing.T) (*WidgetHandler, *httptest.Server) {
	t.Helper()
	mux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/languages") {
			_ = json.NewEncoder(w).Encode(map[string]int64{"Go": 1000, "Shell": 200})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name": "hello-world", "description": "Test repo",
				"stargazers_count": 100, "forks_count": 10, "open_issues_count": 5,
				"license":    map[string]string{"spdx_id": "MIT"},
				"updated_at": "2024-01-15T00:00:00Z", "language": "Go",
				"owner": map[string]string{"login": "octocat"},
			})
		}
	})
	server := httptest.NewServer(mux)

	client := githubclient.NewClient(githubclient.WithBaseURL(server.URL))
	c := cache.New(time.Minute, 100)
	svc := service.NewRepositoryService(client, c)
	h := NewWidgetHandler(svc, slog.Default())
	return h, server
}

func TestWidgetHandlerSuccess(t *testing.T) {
	h, server := newTestHandler(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/api/widget/octocat/hello-world", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "image/svg+xml") {
		t.Errorf("Content-Type = %s, want image/svg+xml", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "octocat") {
		t.Error("SVG missing owner name")
	}
	if !strings.Contains(body, "hello-world") {
		t.Error("SVG missing repo name")
	}
	if !strings.Contains(body, "<svg") {
		t.Error("response is not SVG")
	}
}

func TestWidgetHandlerInvalidOwner(t *testing.T) {
	h, server := newTestHandler(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/api/widget/../etc/passwd", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("status = %d, want 400 for path traversal", rec.Code)
	}
}

func TestWidgetHandlerUnknownTheme(t *testing.T) {
	h, server := newTestHandler(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/api/widget/octocat/hello-world?theme=invalid", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("status = %d, want 400 for unknown theme", rec.Code)
	}
}

func TestWidgetHandlerDarkTheme(t *testing.T) {
	h, server := newTestHandler(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/api/widget/octocat/hello-world?theme=dark", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestWidgetHandlerNeonTheme(t *testing.T) {
	h, server := newTestHandler(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/api/widget/octocat/hello-world?theme=neon", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "#0d0221") {
		t.Error("SVG missing neon background color")
	}
}

func TestWidgetHandlerNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
	}))
	defer server.Close()

	client := githubclient.NewClient(githubclient.WithBaseURL(server.URL))
	c := cache.New(time.Minute, 100)
	svc := service.NewRepositoryService(client, c)
	h := NewWidgetHandler(svc, slog.Default())

	req := httptest.NewRequest("GET", "/api/widget/octocat/nonexistent", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestWidgetHandlerRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(403)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "rate limited"})
	}))
	defer server.Close()

	client := githubclient.NewClient(githubclient.WithBaseURL(server.URL))
	c := cache.New(time.Minute, 100)
	svc := service.NewRepositoryService(client, c)
	h := NewWidgetHandler(svc, slog.Default())

	req := httptest.NewRequest("GET", "/api/widget/octocat/hello-world", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("status = %d, want 429", rec.Code)
	}
}

func TestWidgetHandlerCustomDimensions(t *testing.T) {
	h, server := newTestHandler(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/api/widget/octocat/hello-world?width=800&height=300", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `width="800"`) {
		t.Error("SVG missing custom width")
	}
}

func TestValidOwnerRepo(t *testing.T) {
	tests := []struct {
		owner, repo string
		want        bool
	}{
		{"octocat", "hello-world", true},
		{"user", "repo.name", true},
		{"user_name", "repo_name", true},
		{"", "repo", false},
		{"owner", "", false},
		{"../etc", "passwd", false},
		{"-bad", "repo", false},
		{"user@bad", "repo", false},
	}
	for _, tt := range tests {
		got := validOwnerRepo(tt.owner, tt.repo)
		if got != tt.want {
			t.Errorf("validOwnerRepo(%q,%q) = %v, want %v", tt.owner, tt.repo, got, tt.want)
		}
	}
}
