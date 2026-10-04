package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kirasync2748/github-widget/internal/cache"
	githubclient "github.com/kirasync2748/github-widget/internal/github"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/repos/octocat/hello-world":
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
		case r.URL.Path == "/repos/octocat/hello-world/commits":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"commit": map[string]any{
					"committer": map[string]string{"date": "2024-06-01T12:00:00Z"},
				}},
			})
		case r.URL.Path == "/repos/octocat/hello-world/languages":
			_ = json.NewEncoder(w).Encode(map[string]int64{
				"Go":    7200,
				"HTML":  1800,
				"Shell": 1000,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
		}
	}))
}

func TestGetWidgetData(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := githubclient.NewClient(githubclient.WithBaseURL(server.URL))
	c := cache.New(time.Minute, 100)
	svc := NewRepositoryService(client, c)

	data, err := svc.GetWidgetData(context.Background(), "octocat", "hello-world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Repository.Name != "hello-world" {
		t.Errorf("name = %q", data.Repository.Name)
	}
	if len(data.Languages) != 3 {
		t.Fatalf("expected 3 languages, got %d", len(data.Languages))
	}

	// Languages should be sorted by percentage descending.
	if data.Languages[0].Name != "Go" {
		t.Errorf("expected Go first, got %s", data.Languages[0].Name)
	}
	if data.Languages[0].Percentage < 70 || data.Languages[0].Percentage > 73 {
		t.Errorf("Go percentage = %.1f, expected ~72", data.Languages[0].Percentage)
	}

	wantCommit := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if !data.Repository.LastCommitAt.Equal(wantCommit) {
		t.Errorf("LastCommitAt = %v, want %v", data.Repository.LastCommitAt, wantCommit)
	}
}

func TestCacheHitAvoidsNetworkCall(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	callCount := 0
	wrappedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/repos/octocat/hello-world":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name": "hello-world", "stargazers_count": 100, "forks_count": 10,
				"open_issues_count": 5, "updated_at": "2024-01-15T00:00:00Z",
				"language": "Go", "owner": map[string]string{"login": "octocat"},
			})
		case r.URL.Path == "/repos/octocat/hello-world/languages":
			_ = json.NewEncoder(w).Encode(map[string]int64{"Go": 100})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer wrappedServer.Close()

	client := githubclient.NewClient(githubclient.WithBaseURL(wrappedServer.URL))
	c := cache.New(time.Minute, 100)
	svc := NewRepositoryService(client, c)

	// First call hits network.
	_, _ = svc.GetWidgetData(context.Background(), "octocat", "hello-world")
	firstCount := callCount

	// Second call should be cached.
	_, _ = svc.GetWidgetData(context.Background(), "octocat", "hello-world")
	if callCount != firstCount {
		t.Errorf("second call should be cached; callCount went from %d to %d", firstCount, callCount)
	}
}

func TestConcurrentGetWidgetData(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := githubclient.NewClient(githubclient.WithBaseURL(server.URL))
	c := cache.New(time.Minute, 100)
	svc := NewRepositoryService(client, c)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.GetWidgetData(context.Background(), "octocat", "hello-world")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()
}
