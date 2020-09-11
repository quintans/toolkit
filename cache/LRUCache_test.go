package cache

import (
	"testing"
)

func TestCapacity(t *testing.T) {
	lru := NewLRUCache(3)
	lru.Put("one", "um")
	lru.Put("two", "dois")
	lru.Put("three", "tres")
	lru.Put("ofourne", "quatro")
}
