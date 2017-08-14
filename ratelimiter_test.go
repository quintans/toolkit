package toolkit

import (
	"fmt"
	"testing"
	"time"
)

func TestRate(t *testing.T) {
	var rl = NewRateLimiter(1) // per second
	var start = time.Now()
	for i := 0; i < 5; i++ {
		var d = rl.Take()
		fmt.Println(time.Now(), d)
	}
	var delta = time.Now().Sub(start)
	if delta > time.Second*5 {
		t.Fatal("Expected less than 5s, got", delta)
	}
}
