package breaker

import (
	"errors"
	"testing"
	"time"
)

func TestOpenSimple(t *testing.T) {
	var cb = New(Config{
		Maxfailures:  2,
		ResetTimeout: time.Second,
	})
	var calls = 0
	var fails = 0
	var success = func() error {
		calls++
		return nil
	}
	var failure = func() error {
		calls++
		return errors.New("Test")
	}
	var fallback = func(err error) error {
		fails++
		return err
	}

	// calling success and fallback
	<-cb.Try(failure, fallback)
	if cb.State() != CLOSE {
		t.Fatal("Expected CLOSED, got", cb.State())
	}
	if calls != 1 {
		t.Fatal("Expected calls=1, got", calls)
	}
	if fails != 1 {
		t.Fatal("Expected fails=1, got", fails)
	}

	// calling success and fallback and opening the circuit
	<-cb.Try(failure, fallback)
	if cb.State() != OPEN {
		t.Fatal("Expected OPEN, got", cb.State())
	}
	if calls != 2 {
		t.Fatal("Expected calls=2, got", calls)
	}
	if fails != 2 {
		t.Fatal("Expected fails=2, got", fails)
	}

	// calling only failure
	<-cb.Try(failure, fallback)
	if cb.State() != OPEN {
		t.Fatal("Expected OPEN, got", cb.State())
	}
	if calls != 2 {
		t.Fatal("Expected calls=2, got", calls)
	}
	if fails != 3 {
		t.Fatal("Expected fails=3, got", fails)
	}

	// reset timeout
	time.Sleep(time.Second * 2)
	// calling failure and fallback
	<-cb.Try(failure, fallback)
	if cb.State() != OPEN {
		t.Fatal("Expected OPEN, got", cb.State())
	}
	if calls != 3 {
		t.Fatal("Expected calls=3, got", calls)
	}
	if fails != 4 {
		t.Fatal("Expected fails=4, got", fails)
	}

	// calling directly the fallback
	<-cb.Try(failure, fallback)
	if cb.State() != OPEN {
		t.Fatal("Expected OPEN, got", cb.State())
	}
	if calls != 3 {
		t.Fatal("Expected calls=3, got", calls)
	}
	if fails != 5 {
		t.Fatal("Expected fails=5, got", fails)
	}

	// reset timeout
	time.Sleep(time.Second * 2)
	// calling only success
	<-cb.Try(success, fallback)
	if cb.State() != CLOSE {
		t.Fatal("Expected CLOSE, got", cb.State())
	}
	if calls != 4 {
		t.Fatal("Expected calls=3, got", calls)
	}
	if fails != 5 {
		t.Fatal("Expected fails=4, got", fails)
	}
}
