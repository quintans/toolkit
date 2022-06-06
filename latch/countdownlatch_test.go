package latch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCountdownLatchOK(t *testing.T) {
	latch := NewCountDownLatch()
	latch.Add(2)
	go func() {
		latch.Done()
	}()
	go func() {
		latch.Done()
	}()

	timeout := latch.WaitWithTimeout(time.Second)
	require.False(t, timeout)
}

func TestCountdownLatchClose(t *testing.T) {
	latch := NewCountDownLatch()
	latch.Add(2)
	go func() {
		latch.Close()
	}()
	timeout := latch.WaitWithTimeout(time.Second)
	require.False(t, timeout)
}

func TestCountdownLatchWithTimeout(t *testing.T) {
	latch := NewCountDownLatch()
	latch.Add(2)

	timeout := latch.WaitWithTimeout(time.Second)
	require.True(t, timeout)
}
