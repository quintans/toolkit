package toolkit

import (
	"sync"
	"time"
)

type Rate interface {
	TakeN(int64) time.Duration
}

var _ Rate = &RateLimiter{}

// RateLimiter is a simple implementation of the Leaky Bucket algorithm.
//
// Simple use:
// var rl = NewRateLimiter(2) // per second
// for i := 0; i < 10; i++ {
//   var wait = rl.Take()
//   fmt.Println(time.Now(), wait)
// }
type RateLimiter struct {
	sync.Mutex
	nextTake time.Time
	perTake  int64
}

// NewRateLimiter creates an instance of RateLimiter
// rate sets the number of takes that can occur per second
func NewRateLimiter(rate int64) *RateLimiter {
	return &RateLimiter{
		perTake: int64(time.Second) / rate,
	}
}

// TakeN enforces the rate limit.
// amount is the value over which we apply the rate and it returns the time waiting before returning.
// If previous calls to TakeN() had the consequence of breaking the rate limit,
// then the current call will wait until the rate is again below the rate limit.
// If we want to limit a stream of data to 4Kb/s we do the following:
// var rl = NewRateLimiter(4000) // per second
// func submit(data []byte)
//   rl.Take(len(data))
//   service.send(data);
// }
func (rl *RateLimiter) TakeN(amount int64) time.Duration {
	rl.Lock()
	defer rl.Unlock()

	var now = time.Now()
	var t time.Duration
	if now.Before(rl.nextTake) {
		t = rl.nextTake.Sub(now)
		time.Sleep(t)
	}
	rl.nextTake = time.Now().Add(time.Duration(rl.perTake * amount))
	return t
}

// Take is the same as TakeN(1)
func (rl *RateLimiter) Take() time.Duration {
	return rl.TakeN(1)
}
