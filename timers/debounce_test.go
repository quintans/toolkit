package timers

import (
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	cnt := 0
	ch := make(chan interface{}, 1)
	d := NewDebounce(500*time.Millisecond, func(arg interface{}) {
		cnt++
		ch <- arg
	})
	time.Sleep(400 * time.Millisecond)
	d.Delay("a")
	time.Sleep(400 * time.Millisecond)
	d.Delay("b")
	time.Sleep(400 * time.Millisecond)
	d.Delay("c")
	time.Sleep(400 * time.Millisecond)
	r := <-ch
	s, ok := r.(string)
	if !ok || s != "c" {
		t.Fatal("Expected 'c', got ", s)
	}
}
