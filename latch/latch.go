package latch

import "sync"

// Latch reusable lock that uses a channel to wait for the release of the lock
type Latch struct {
	ch     chan struct{}
	closed bool
	mu     sync.RWMutex
}

// NewLatch creates a new LockerChan
func NewLatch() *Latch {
	return &Latch{}
}

// Lock locks the release of wait
func (c *Latch) Lock() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ch = make(chan struct{})
	c.closed = false
}

// Unlock releases the lock
func (c *Latch) Unlock() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		close(c.ch)
		c.closed = true
	}
}

// Wait wait for the lock to be released
func (c *Latch) Wait() <-chan struct{} {
	return c.ch
}
