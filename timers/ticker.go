package timers

import (
	"sync"
	"time"
)

type Ticker struct {
	once   sync.Once
	stop   chan struct{}
	ticker *time.Ticker
}

func NewTicker(duration time.Duration, hnd func(time.Time)) *Ticker {
	return NewDelayedTicker(duration, duration, hnd)
}

func NewDelayedTicker(delay time.Duration, duration time.Duration, hnd func(time.Time)) *Ticker {
	tck := &Ticker{
		stop: make(chan struct{}, 1),
	}

	go func() {
		if delay > 0 {
			time.Sleep(delay)
		}
		hnd(time.Now())

		tck.ticker = time.NewTicker(duration)
		for {
			select {
			case <-tck.stop:
				return
			case t := <-tck.ticker.C:
				hnd(t)
			}
		}
	}()

	return tck
}

func (tck *Ticker) Stop() {
	tck.once.Do(func() {
		close(tck.stop)
	})
}
