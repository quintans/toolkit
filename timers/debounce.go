package timers

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
	d := &Debouncer{}
	d.input = make(chan interface{}, 10)

	go func() {
		// clean up
		defer func() {
			if d.OnExit != nil {
				d.OnExit()
			}
		}()

		var item interface{}
		var ok bool
		t := time.NewTimer(interval)
		for {
			select {
			case item, ok = <-d.input:
				if !ok {
					// was closed
					return
				}
				t.Reset(interval)
			case <-t.C:
				action(item)
				return
			}
		}
	}()

	return d
}

// Delay delays the execution of action declared when we created the debouncer
func (d *Debouncer) Delay(item interface{}) {
	d.input <- item
}

// Kill terminates the debouncer
func (d *Debouncer) Kill() {
	d.once.Do(func() {
		close(d.input)
	})
}
