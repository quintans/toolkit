package toolkit

import (
	"sync"
	"time"
)

// Debouncer struct to support debouncing
type Debouncer struct {
	once   sync.Once
	input  chan interface{}
	OnExit func()
}

// NewDebounce creates a new Debouncer
func NewDebounce(interval time.Duration, action func(arg interface{})) *Debouncer {
	var debounce = &Debouncer{}
	debounce.input = make(chan interface{}, 10)

	go func(input chan interface{}) {
		// clean up
		defer func() {
			if debounce.OnExit != nil {
				debounce.OnExit()
			}
		}()

		var item interface{}
		var ok bool
		for {
			select {
			case item, ok = <-input:
				if !ok {
					// was closed
					return
				}
			case <-time.After(interval):
				action(item)
				return
			}
		}
	}(debounce.input)

	return debounce
}

// Delay delays the execution of action declared when we created the debouncer
func (debounce *Debouncer) Delay(item interface{}) {
	debounce.input <- item
}

// Kill terminates the debouncer
func (debounce *Debouncer) Kill() {
	debounce.once.Do(func() {
		close(debounce.input)
	})
}
