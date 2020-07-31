package locks

import (
	"sync"
	"time"
)

// NewCountDownLatch creates a new CountDownLatch
func NewCountDownLatch() *CountDownLatch {
	return &CountDownLatch{
		ch: make(chan struct{}),
	}
}

// CountDownLatch is the same as sync.WaitGroup but with the ability to wait with timeout
type CountDownLatch struct {
	mu      sync.RWMutex
	counter int
	ch      chan struct{}
	closed  bool
}

// Add increases/decreases the countdown
func (l *CountDownLatch) Add(delta int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.counter += delta
	if l.counter <= 0 && !l.closed {
		l.closed = true
		close(l.ch)
	}
}

// Done decreases the countdown by one
func (l *CountDownLatch) Done() {
	l.Add(-1)
}

// Wait to be unblocked
func (l *CountDownLatch) Wait() <-chan struct{} {
	return l.ch
}

// WaitWithTimeout waits until the timeout runs out or until the countdown is zero
func (l *CountDownLatch) WaitWithTimeout(timeout time.Duration) bool {
	select {
	case <-l.ch:
		return false // completed normally
	case <-time.After(timeout):
		l.Close()
		return true // timed out
	}
}

// Close closes the latch unblocking wait
func (l *CountDownLatch) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.closed {
		l.closed = true
		close(l.ch)
	}
}

// Counter returns the current count down number
func (l *CountDownLatch) Counter() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.counter
}
