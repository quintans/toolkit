package deadlock

import (
	"fmt"
	"testing"
	"time"
)

func TestDead(t *testing.T) {
	Config.DeadlockTimeout = time.Second

	var mu RWMutex
	test1(mu)
}

func test1(mu RWMutex) {
	mu.Lock()
	test2(mu)
	mu.Unlock()
}

func test2(mu RWMutex) {
	mu.Lock()
	fmt.Println("I am here")
	mu.Unlock()
}
