// Package github provides a client for the GitHub REST API with clean domain models.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client is the GitHub API client. It is safe for concurrent use.
type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
	userAgent  string
}

// Option configures a Client.
type Option func(*Client)

// WithToken sets the GitHub authentication token.
func WithToken(token string) Option {
	return func(c *Client) { c.token = token }
}

// WithBaseURL overrides the default API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(url, "/") }
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// NewClient creates a configured GitHub API client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.github.com",
		userAgent:  "github-widget/1.0",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// apiError represents a GitHub API error response.
type apiError struct {
	Message string `json:"message"`
}

// repoResponse is the raw GitHub /repos/{owner}/{repo} response.
type repoResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	OpenIssues  int    `json:"open_issues_count"`
	License     *struct {
		SpdxID string `json:"spdx_id"`
	} `json:"license"`
	UpdatedAt string `json:"updated_at"`
	HTMLURL   string `json:"html_url"`
	Owner     struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
	Language string `json:"language"`
}

// FetchRepository retrieves repository metadata.
func (c *Client) FetchRepository(ctx context.Context, owner, repo string) (Repository, error) {
	owner = strings.ToLower(strings.TrimSpace(owner))
	repo = strings.TrimSpace(repo)

	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	var raw repoResponse
	if err := c.doJSON(ctx, http.MethodGet, path, &raw); err != nil {
		return Repository{}, err
	}

	r := Repository{
		Owner:       raw.Owner.Login,
		OwnerAvatar: raw.Owner.AvatarURL,
		Name:        raw.Name,
		Description: raw.Description,
		Stars:       raw.Stars,
		Forks:       raw.Forks,
		OpenIssues:  raw.OpenIssues,
		PrimaryLang: raw.Language,
		HTMLURL:     raw.HTMLURL,
	}
	if raw.License != nil && raw.License.SpdxID != "" && raw.License.SpdxID != "NOASSERTION" {
		r.License = raw.License.SpdxID
	}
	if raw.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, raw.UpdatedAt); err == nil {
			r.UpdatedAt = t
		}
	}
	return r, nil
}

// FetchLanguages retrieves language byte counts for a repository.
func (c *Client) FetchLanguages(ctx context.Context, owner, repo string) ([]LanguageStat, error) {
	owner = strings.ToLower(strings.TrimSpace(owner))
	repo = strings.TrimSpace(repo)

	path := fmt.Sprintf("/repos/%s/%s/languages", owner, repo)
	var raw map[string]int64
	if err := c.doJSON(ctx, http.MethodGet, path, &raw); err != nil {
		return nil, err
	}

	stats := make([]LanguageStat, 0, len(raw))
	for name, bytes := range raw {
		stats = append(stats, LanguageStat{Name: name, Bytes: bytes})
	}
	return stats, nil
}

// doJSON performs a GitHub API GET request and decodes JSON into v.
func (c *Client) doJSON(ctx context.Context, method, path string, v any) error {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &UpstreamError{Code: http.StatusBadGateway, Message: "failed to reach GitHub API"}
	}
	defer resp.Body.Close()

	// Limit response body to 5 MiB to prevent excessive memory use.
	body := io.LimitReader(resp.Body, 5<<20)

	if resp.StatusCode >= 400 {
		var ae apiError
		_ = json.NewDecoder(body).Decode(&ae)
		msg := ae.Message
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return &UpstreamError{Code: resp.StatusCode, Message: msg, RateLimit: resp.StatusCode == http.StatusForbidden}
	}

	if err := json.NewDecoder(body).Decode(v); err != nil {
		return &UpstreamError{Code: http.StatusBadGateway, Message: "malformed response from GitHub"}
	}
	return nil
}

// RateLimitRemaining returns the remaining rate limit from a response if available.
func (c *Client) RateLimitRemaining(resp *http.Response) int {
	if resp == nil {
		return -1
	}
	n, err := strconv.Atoi(resp.Header.Get("X-RateLimit-Remaining"))
	if err != nil {
		return -1
	}
	return n
}

// UpstreamError represents an error from the GitHub API.
type UpstreamError struct {
	Code      int
	Message   string
	RateLimit bool
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("github api error %d: %s", e.Code, e.Message)
}

// IsNotFound returns true if the upstream returned 404.
func (e *UpstreamError) IsNotFound() bool { return e.Code == http.StatusNotFound }

// IsRateLimited returns true if the upstream returned a rate-limit error.
func (e *UpstreamError) IsRateLimited() bool { return e.RateLimit }
