package toolkit

import (
	"fmt"
	"testing"
	"time"
)

func TestRate(t *testing.T) {
	var rl = NewRateLimiter(1) // per second
	rl.SetBurst(3)
	for i := 0; i < 6; i++ {
		var d = rl.Take()
		fmt.Println(time.Now(), d)
	}
}
