package timers

import (
	"fmt"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	cnt := 0
	tick := NewTicker(time.Second, func(t time.Time) {
		cnt++
		fmt.Println(t, "count=", cnt)
	})

	<-time.After(time.Millisecond * 3500)
	tick.Stop()
	if cnt != 3 {
		t.Fatal("Expected 3 counts, got", cnt)
	}
}
