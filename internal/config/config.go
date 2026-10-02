// Package config provides typed configuration for the github-widget server,
// loaded from environment variables with sensible defaults.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Port            string
	GitHubToken     string
	GitHubAPIURL    string
	CacheTTL        time.Duration
	CacheMaxEntries int
	LogLevel        string
}

// Load reads configuration from environment variables.
// It fails fast on genuinely invalid values.
func Load() (Config, error) {
	cfg := Config{
		Port:            envStr("PORT", "3000"),
		GitHubToken:     envStr("GITHUB_TOKEN", ""),
		GitHubAPIURL:    strings.TrimRight(envStr("GITHUB_API_URL", "https://api.github.com"), "/"),
		CacheTTL:        envDuration("CACHE_TTL", 10*time.Minute),
		CacheMaxEntries: envInt("CACHE_MAX_ENTRIES", 1000),
		LogLevel:        strings.ToLower(envStr("LOG_LEVEL", "info")),
	}

	if cfg.CacheMaxEntries < 1 {
		return Config{}, fmt.Errorf("CACHE_MAX_ENTRIES must be >= 1, got %d", cfg.CacheMaxEntries)
	}
	if cfg.CacheTTL < 1*time.Second {
		return Config{}, fmt.Errorf("CACHE_TTL must be >= 1s, got %s", cfg.CacheTTL)
	}
	if cfg.Port == "" {
		return Config{}, fmt.Errorf("PORT must not be empty")
	}

	return cfg, nil
}

// Addr returns the listen address derived from Port.
func (c Config) Addr() string {
	return ":" + strings.TrimPrefix(c.Port, ":")
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return def
		}
		return n
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return def
		}
		return d
	}
	return def
}
