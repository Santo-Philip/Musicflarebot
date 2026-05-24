package cache

import (
	"sync"
	"time"
)

type item[T any] struct {
	value     T
	expiresAt time.Time
}

type Cache[T any] struct {
	mu       sync.RWMutex
	items    map[string]item[T]
	ttl      time.Duration
	stopChan chan struct{}
}

func New[T any](ttl time.Duration) *Cache[T] {
	c := &Cache[T]{
		items:    make(map[string]item[T]),
		ttl:      ttl,
		stopChan: make(chan struct{}),
	}
	go c.evictor()
	return c
}

func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item[T]{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	it, ok := c.items[key]
	if !ok {
		var zero T
		return zero, false
	}
	if time.Now().After(it.expiresAt) {
		var zero T
		return zero, false
	}
	return it.value, true
}

func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *Cache[T]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]item[T])
}

func (c *Cache[T]) evictor() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.items {
				if now.After(v.expiresAt) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		case <-c.stopChan:
			return
		}
	}
}

func (c *Cache[T]) Stop() {
	close(c.stopChan)
}
