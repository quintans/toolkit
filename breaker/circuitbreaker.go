package breaker

import (
	"errors"
	"sync"
	"time"

	"github.com/imdario/mergo"
)

var TimeoutError = errors.New("Circuit Breaker Timeout")

type EState int

var states = [...]string{"CLOSED", "OPEN", "HALFOPEN"}

func (e EState) String() string {
	return states[e]
}

const (
	CLOSE EState = iota
	OPEN
	HALFOPEN
)

type Config struct {
	Timeout      time.Duration
	Maxfailures  int
	ResetTimeout time.Duration
}

var defaultConfig = Config{
	Maxfailures:  2,
	ResetTimeout: time.Second * 3,
}

type Stats struct {
	Calls int64
	Fails int64
}

type CircuitBreaker struct {
	Config
	sync.RWMutex

	OnChange func(EState)

	failures  int
	state     EState
	openUntil time.Time
	stats     Stats
}

func New(cfg Config) *CircuitBreaker {
	var cb = &CircuitBreaker{}
	mergo.Merge(&cfg, defaultConfig)
	cb.Config = cfg
	return cb
}

func (cb *CircuitBreaker) Try(fn func() error, fallback func(err error) error) <-chan error {
	cb.Lock()
	defer cb.Unlock()

	cb.stats.Calls++
	var cherr = make(chan error, 1)
	if cb.state == CLOSE {
		var ch = cb.call(fn)
		go func() {
			var err = <-ch
			if err != nil {
				cb.fail()
				if fallback != nil {
					err = fallback(err)
				}
			} else {
				cb.reset()
			}
			cherr <- err
		}()
	} else if cb.state == OPEN {
		var now = time.Now()
		if now.After(cb.openUntil) {
			var ch = cb.call(fn)
			go func() {
				var err = <-ch
				if err != nil {
					cb.fail()
					if fallback != nil {
						err = fallback(err)
					}
				} else {
					cb.ok()
				}
				cherr <- err
			}()
		} else {
			if fallback != nil {
				cherr <- fallback(nil)
			} else {
				cherr <- nil
			}
		}
	}
	return cherr
}

func (cb *CircuitBreaker) call(fn func() error) <-chan error {
	var ch = make(chan error, 1)
	go func() {
		ch <- fn()
	}()
	if cb.Timeout != time.Duration(0) {
		var cherr = make(chan error, 1)
		go func() {
			select {
			case <-time.After(cb.Timeout):
				cherr <- TimeoutError
			case err := <-ch:
				cherr <- err
			}
		}()
		return cherr
	} else {
		return ch
	}
}

func (cb *CircuitBreaker) fail() {
	cb.Lock()
	defer cb.Unlock()

	cb.stats.Fails++
	if cb.state == CLOSE {
		cb.failures++
		if cb.failures >= cb.Maxfailures {
			cb.state = OPEN
			cb.openUntil = time.Now().Add(cb.ResetTimeout)
			if cb.OnChange != nil {
				go cb.OnChange(cb.state)
			}
		}
	} else {
		cb.openUntil = time.Now().Add(cb.ResetTimeout)
	}
}

func (cb *CircuitBreaker) reset() {
	cb.Lock()
	cb.failures = 0
	cb.Unlock()
}

func (cb *CircuitBreaker) ok() {
	cb.Lock()
	defer cb.Unlock()

	cb.state = CLOSE
	cb.failures = 0
	if cb.OnChange != nil {
		go cb.OnChange(cb.state)
	}
}

func (cb *CircuitBreaker) State() EState {
	cb.RLock()
	defer cb.RUnlock()

	if cb.state == OPEN {
		if time.Now().After(cb.openUntil) {
			return HALFOPEN
		} else {
			return OPEN
		}
	} else {
		return CLOSE
	}

}

func (cb *CircuitBreaker) Stats() Stats {
	cb.RLock()
	defer cb.RUnlock()

	return cb.stats
}
