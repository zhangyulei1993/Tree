package contentsafety

import (
	"sync"
	"time"
)

type verdictCache interface {
	Get(key string) (Result, bool)
	Set(key string, result Result)
}

type memoryVerdictCache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]cacheEntry
}

type cacheEntry struct {
	result    Result
	expiresAt time.Time
}

func newMemoryVerdictCache(ttl time.Duration) *memoryVerdictCache {
	return &memoryVerdictCache{
		ttl:   ttl,
		items: make(map[string]cacheEntry),
	}
}

func (c *memoryVerdictCache) Get(key string) (Result, bool) {
	now := time.Now()
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || now.After(entry.expiresAt) {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}
		return Result{}, false
	}
	return entry.result, true
}

func (c *memoryVerdictCache) Set(key string, result Result) {
	c.mu.Lock()
	c.items[key] = cacheEntry{result: result, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
