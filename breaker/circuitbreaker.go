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

type Metrics interface {
	IncSuccess()
	Sucess() uint64
	IncFailure()
	Data() Stats
	Clear() Stats
}

type Stats struct {
	Successes uint64
	Fails     uint64
}

type Config struct {
	Timeout      time.Duration
	Maxfailures  int //consecutive failures
	ResetTimeout time.Duration
}

type CircuitBreaker struct {
	Config
	sync.RWMutex

	OnChange func(EState)

	failures  int
	state     EState
	openUntil time.Time
	metrics   Metrics
}

func New(cfg Config) *CircuitBreaker {
	var cb = &CircuitBreaker{
		state: CLOSE,
	}

	var defaultConfig = Config{
		Maxfailures:  5,
		ResetTimeout: time.Second * 10,
	}

	mergo.Merge(&cfg, defaultConfig)
	cb.Config = cfg
	return cb
}

func (cb *CircuitBreaker) Try(fn func() error, fallback func(err error) error) <-chan error {
	cb.Lock()
	defer cb.Unlock()

	var cherr = make(chan error, 1)
	go func(run bool) {
		if run {
			var err = cb.call(fn)
			if err != nil {
				cb.fail()
				if fallback != nil {
					err = fallback(err)
				}
			} else {
				cb.reset()
			}
			cherr <- err
		} else if fallback != nil {
			cherr <- fallback(nil)
		} else {
			cherr <- nil
		}

	}(cb.state == CLOSE || time.Now().After(cb.openUntil))
	return cherr
}

func (cb *CircuitBreaker) call(fn func() error) error {
	var ch = make(chan error, 1)
	go func() {
		ch <- fn()
	}()
	if cb.Timeout != time.Duration(0) {
		var err error
		select {
		case <-time.After(cb.Timeout):
			err = TimeoutError
		case err = <-ch:
		}
		return err
	} else {
		return <-ch
	}
}

func (cb *CircuitBreaker) fail() {
	cb.Lock()

	var changed = false

	if cb.metrics != nil {
		cb.metrics.IncFailure()
	}
	if cb.state == CLOSE {
		cb.failures++
		if cb.failures >= cb.Maxfailures {
			cb.state = OPEN
			cb.openUntil = time.Now().Add(cb.ResetTimeout)
			changed = true
		}
	} else {
		cb.openUntil = time.Now().Add(cb.ResetTimeout)
	}

	cb.Unlock()

	if changed && cb.OnChange != nil {
		go cb.OnChange(OPEN)
	}

}

func (cb *CircuitBreaker) reset() {
	cb.Lock()

	var changed = cb.state != CLOSE
	cb.state = CLOSE
	cb.failures = 0
	if cb.metrics != nil {
		cb.metrics.IncSuccess()
	}
	cb.Unlock()

	if cb.OnChange != nil && changed {
		go cb.OnChange(CLOSE)
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
	if cb.metrics != nil {
		return cb.metrics.Data()
	}
	return Stats{}
}

func (cb *CircuitBreaker) ClearStats() Stats {
	cb.RLock()
	defer cb.RUnlock()
	if cb.metrics != nil {
		return cb.metrics.Clear()
	}
	return Stats{}
}
