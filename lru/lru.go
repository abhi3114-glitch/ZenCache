package lru

import (
	"container/list"
	"sync"
)

// Cache is a thread-safe LRU cache.
type Cache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	order    *list.List // Front = most recently used, Back = least recently used
}

type entry struct {
	key   string
	value string
}

// NewCache creates a new LRU cache with the given capacity.
func NewCache(capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Set adds or updates a key-value pair. Returns evicted key if eviction occurred.
func (c *Cache) Set(key, value string) (evictedKey string, evicted bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		// Key exists, update value and move to front
		c.order.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return "", false
	}

	// New key - check if we need to evict
	if c.order.Len() >= c.capacity {
		// Evict LRU (back of list)
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			evictedEntry := oldest.Value.(*entry)
			delete(c.items, evictedEntry.key)
			evictedKey = evictedEntry.key
			evicted = true
		}
	}

	// Add new entry to front
	elem := c.order.PushFront(&entry{key: key, value: value})
	c.items[key] = elem
	return evictedKey, evicted
}

// Get retrieves a value by key and marks it as recently used.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return "", false
}

// Del removes a key from the cache.
func (c *Cache) Del(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.Remove(elem)
		delete(c.items, key)
		return true
	}
	return false
}

// Len returns the current number of items in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// Keys returns all keys in order from most to least recently used.
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, c.order.Len())
	for e := c.order.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry).key)
	}
	return keys
}

// GetAllData returns a copy of all data for persistence.
func (c *Cache) GetAllData() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data := make(map[string]string, len(c.items))
	for k, v := range c.items {
		data[k] = v.Value.(*entry).value
	}
	return data
}

// LoadData bulk loads data into the cache (used for restoring from persistence).
func (c *Cache) LoadData(data map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, value := range data {
		if c.order.Len() >= c.capacity {
			break // Stop loading if at capacity
		}
		elem := c.order.PushBack(&entry{key: key, value: value})
		c.items[key] = elem
	}
}
