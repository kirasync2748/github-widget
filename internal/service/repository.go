// Package service orchestrates GitHub data fetching and caching
// to produce widget-ready repository data.
package service

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/kirasync2748/github-widget/internal/cache"
	"github.com/kirasync2748/github-widget/internal/github"
	"github.com/kirasync2748/github-widget/internal/svg"
)

// RepositoryService fetches and caches repository widget data.
type RepositoryService struct {
	client *github.Client
	cache  *cache.Cache
}

// NewRepositoryService creates a new service with the given client and cache.
func NewRepositoryService(client *github.Client, c *cache.Cache) *RepositoryService {
	return &RepositoryService{client: client, cache: c}
}

// WidgetData is the complete data needed to render a repository card.
type WidgetData struct {
	Repository github.Repository
	Languages  []github.ComputedLanguage
}

// GetWidgetData fetches repository and language data, using the cache when possible.
// A singleflight-style mutex prevents cache stampede for concurrent identical requests.
func (s *RepositoryService) GetWidgetData(ctx context.Context, owner, repo string) (WidgetData, error) {
	key := fmt.Sprintf("%s/%s", owner, repo)

	if cached, ok := s.cache.Get(key); ok {
		if wd, ok := cached.(WidgetData); ok {
			return wd, nil
		}
	}

	// Prevent stampede: use a per-key mutex.
	mu := s.getLock(key)
	mu.Lock()
	defer mu.Unlock()
	defer s.releaseLock(key)

	// Double-check after acquiring lock.
	if cached, ok := s.cache.Get(key); ok {
		if wd, ok := cached.(WidgetData); ok {
			return wd, nil
		}
	}

	repository, err := s.client.FetchRepository(ctx, owner, repo)
	if err != nil {
		return WidgetData{}, err
	}

	stats, err := s.client.FetchLanguages(ctx, owner, repo)
	if err != nil {
		return WidgetData{}, err
	}

	languages := computeLanguages(stats)

	data := WidgetData{
		Repository: repository,
		Languages:  languages,
	}

	s.cache.Set(key, data)
	return data, nil
}

// computeLanguages converts raw byte counts into sorted, colored percentages.
func computeLanguages(stats []github.LanguageStat) []github.ComputedLanguage {
	var total int64
	for _, s := range stats {
		total += s.Bytes
	}
	if total == 0 {
		return nil
	}

	computed := make([]github.ComputedLanguage, 0, len(stats))
	for _, s := range stats {
		pct := float64(s.Bytes) / float64(total) * 100
		if pct < 0.1 {
			continue
		}
		computed = append(computed, github.ComputedLanguage{
			Name:       s.Name,
			Color:      svg.GetLanguageColor(s.Name),
			Percentage: pct,
		})
	}

	sort.Slice(computed, func(i, j int) bool {
		return computed[i].Percentage > computed[j].Percentage
	})

	// Show only the top 5 languages in the card.
	if len(computed) > 5 {
		computed = computed[:5]
	}

	return computed
}

// --- per-key mutex for stampede prevention ---

var (
	lockMu   sync.Mutex
	keyLocks = make(map[string]*sync.Mutex)
)

func (s *RepositoryService) getLock(key string) *sync.Mutex {
	lockMu.Lock()
	defer lockMu.Unlock()
	mu, ok := keyLocks[key]
	if !ok {
		mu = &sync.Mutex{}
		keyLocks[key] = mu
	}
	return mu
}

func (s *RepositoryService) releaseLock(key string) {
	lockMu.Lock()
	defer lockMu.Unlock()
	delete(keyLocks, key)
}
