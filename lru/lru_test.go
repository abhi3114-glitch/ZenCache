package lru

import (
	"testing"
)

func TestLRUBasicOperations(t *testing.T) {
	cache := NewCache(3)

	// Test Set and Get
	cache.Set("a", "1")
	cache.Set("b", "2")
	cache.Set("c", "3")

	if v, ok := cache.Get("a"); !ok || v != "1" {
		t.Errorf("Expected '1', got '%s'", v)
	}
	if v, ok := cache.Get("b"); !ok || v != "2" {
		t.Errorf("Expected '2', got '%s'", v)
	}
}

func TestLRUEviction(t *testing.T) {
	cache := NewCache(3)

	cache.Set("a", "1")
	cache.Set("b", "2")
	cache.Set("c", "3")

	// Access 'a' to make it recently used
	cache.Get("a")

	// Add 'd', should evict 'b' (least recently used)
	evictedKey, evicted := cache.Set("d", "4")
	if !evicted {
		t.Error("Expected eviction")
	}
	if evictedKey != "b" {
		t.Errorf("Expected 'b' to be evicted, got '%s'", evictedKey)
	}

	// Verify 'b' is gone
	if _, ok := cache.Get("b"); ok {
		t.Error("Expected 'b' to be evicted")
	}

	// Verify others exist
	if _, ok := cache.Get("a"); !ok {
		t.Error("Expected 'a' to exist")
	}
	if _, ok := cache.Get("c"); !ok {
		t.Error("Expected 'c' to exist")
	}
	if _, ok := cache.Get("d"); !ok {
		t.Error("Expected 'd' to exist")
	}
}

func TestLRUUpdate(t *testing.T) {
	cache := NewCache(3)

	cache.Set("a", "1")
	cache.Set("b", "2")
	cache.Set("c", "3")

	// Update 'a' to new value
	cache.Set("a", "100")

	if v, ok := cache.Get("a"); !ok || v != "100" {
		t.Errorf("Expected '100', got '%s'", v)
	}

	// Add 'd', should evict 'b' since 'a' was just updated (moved to front)
	evictedKey, _ := cache.Set("d", "4")
	if evictedKey != "b" {
		t.Errorf("Expected 'b' to be evicted, got '%s'", evictedKey)
	}
}

func TestLRUDel(t *testing.T) {
	cache := NewCache(3)

	cache.Set("a", "1")
	cache.Set("b", "2")

	if !cache.Del("a") {
		t.Error("Expected Del to return true")
	}

	if _, ok := cache.Get("a"); ok {
		t.Error("Expected 'a' to be deleted")
	}

	if cache.Del("nonexistent") {
		t.Error("Expected Del of nonexistent key to return false")
	}
}
