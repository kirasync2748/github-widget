package cache

import (
	"sync"
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	c := New(time.Minute, 10)
	c.Set("key", "value")

	v, ok := c.Get("key")
	if !ok {
		t.Fatal("key not found")
	}
	if v != "value" {
		t.Errorf("value = %v, want value", v)
	}
}

func TestCacheMiss(t *testing.T) {
	c := New(time.Minute, 10)
	if _, ok := c.Get("missing"); ok {
		t.Fatal("expected miss for missing key")
	}
}

func TestCacheExpiration(t *testing.T) {
	c := New(50*time.Millisecond, 10)
	c.Set("key", "value")

	if _, ok := c.Get("key"); !ok {
		t.Fatal("key should exist before expiry")
	}

	time.Sleep(60 * time.Millisecond)
	if _, ok := c.Get("key"); ok {
		t.Fatal("key should be expired")
	}
}

func TestCacheEviction(t *testing.T) {
	c := New(time.Minute, 3)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Set("d", 4) // should evict "a"

	if _, ok := c.Get("a"); ok {
		t.Error("a should have been evicted")
	}
	if _, ok := c.Get("d"); !ok {
		t.Error("d should exist")
	}
}

func TestCacheConcurrent(t *testing.T) {
	c := New(time.Minute, 100)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "key"
			c.Set(key, n)
			c.Get(key)
		}(i)
	}
	wg.Wait()
}

func TestCacheDelete(t *testing.T) {
	c := New(time.Minute, 10)
	c.Set("key", "value")
	c.Delete("key")

	if _, ok := c.Get("key"); ok {
		t.Fatal("key should be deleted")
	}
}
