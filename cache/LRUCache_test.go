package cache

import (
	"io"
	"os"
	"github.com/quintans/toolkit/log"
	"testing"
)

func TestCapacity(t *testing.T) {
	log.Init(io.Writer(os.Stdout))
	log.SetLevel("pqp", log.DEBUG)

	lru := NewLRUCache(3)
	lru.Put("one", "um")
	lru.Put("two", "dois")
	lru.Put("three", "tres")
	lru.Put("ofourne", "quatro")
}
