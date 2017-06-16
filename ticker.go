package toolkit

import (
	"sync"
	"time"
)

type Ticker struct {
	once   sync.Once
	stop   chan struct{}
	ticker *time.Ticker
}

func NewTicker(duration time.Duration, hnd func(time.Time)) Ticker {
	var tck = Ticker{
		stop:   make(chan struct{}, 1),
		ticker: time.NewTicker(duration),
	}

	go func() {
		for {
			select {
			case t := <-tck.ticker.C:
				hnd(t)
			case <-tck.stop:
				return
			}
		}
	}()

	return tck
}

func (tck Ticker) Stop() {
	tck.once.Do(func() {
		close(tck.stop)
		tck.ticker.Stop()
	})
}
