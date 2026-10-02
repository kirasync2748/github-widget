package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/octocat/hello-world" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":              "hello-world",
			"description":       "My first repository",
			"stargazers_count":  100,
			"forks_count":       10,
			"open_issues_count": 5,
			"license":           map[string]string{"spdx_id": "MIT"},
			"updated_at":        "2024-01-15T00:00:00Z",
			"html_url":          "https://github.com/octocat/hello-world",
			"language":          "Go",
			"owner": map[string]string{
				"login":      "octocat",
				"avatar_url": "https://github.com/octocat.png",
			},
		})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	repo, err := client.FetchRepository(context.Background(), "octocat", "hello-world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.Name != "hello-world" {
		t.Errorf("name = %q, want hello-world", repo.Name)
	}
	if repo.Stars != 100 {
		t.Errorf("stars = %d, want 100", repo.Stars)
	}
	if repo.License != "MIT" {
		t.Errorf("license = %q, want MIT", repo.License)
	}
	if repo.PrimaryLang != "Go" {
		t.Errorf("language = %q, want Go", repo.PrimaryLang)
	}
	if repo.Owner != "octocat" {
		t.Errorf("owner = %q, want octocat", repo.Owner)
	}
}

func TestFetchRepositoryNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.FetchRepository(context.Background(), "octocat", "nonexistent")
	if err == nil {
		t.Fatal("expected error for 404")
	}

	ue, ok := err.(*UpstreamError)
	if !ok {
		t.Fatalf("expected *UpstreamError, got %T", err)
	}
	if !ue.IsNotFound() {
		t.Error("expected IsNotFound() to be true")
	}
}

func TestFetchRepositoryRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "API rate limit exceeded"})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.FetchRepository(context.Background(), "octocat", "hello-world")
	if err == nil {
		t.Fatal("expected error for rate limit")
	}

	ue, ok := err.(*UpstreamError)
	if !ok {
		t.Fatalf("expected *UpstreamError, got %T", err)
	}
	if !ue.IsRateLimited() {
		t.Error("expected IsRateLimited() to be true")
	}
}

func TestFetchLanguages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/octocat/hello-world/languages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int64{
			"Go":    1000,
			"Shell": 200,
			"HTML":  100,
		})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	stats, err := client.FetchLanguages(context.Background(), "octocat", "hello-world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stats) != 3 {
		t.Fatalf("expected 3 languages, got %d", len(stats))
	}

	found := false
	for _, s := range stats {
		if s.Name == "Go" && s.Bytes == 1000 {
			found = true
		}
	}
	if !found {
		t.Error("Go with 1000 bytes not found")
	}
}

func TestFetchRepositoryServer5xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.FetchRepository(context.Background(), "octocat", "hello-world")
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestClientTokenHeader(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL), WithToken("test-token-123"))
	_, _ = client.FetchRepository(context.Background(), "octocat", "hello-world")

	if gotAuth != "Bearer test-token-123" {
		t.Errorf("Authorization = %q, want Bearer test-token-123", gotAuth)
	}
}

func TestFetchRepositoryTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // never respond
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.FetchRepository(context.Background(), "octocat", "hello-world")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
