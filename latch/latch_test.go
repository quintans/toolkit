package latch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLatch(t *testing.T) {
	latch := NewLatch()
	latch.Lock()
	var ok bool
	go func() {
		time.Sleep(time.Second)
		latch.Unlock()
		latch.Unlock() // has no effect
		ok = true
	}()
	<-latch.Wait()
	require.True(t, ok)
}
