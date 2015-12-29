package test

import (
	"testing"
	"time"

	tk "github.com/quintans/toolkit"
	"github.com/quintans/toolkit/collection"
)

const (
	fifoDir = "fifo_test"
)

var (
	messages = [...]string{"zero", "one", "two", "three", "four"}
)

func TestFilePushPop1(t *testing.T) {
	mega := make([]byte, 1024*1024)
	size := len(mega)
	zero := byte(50)
	for i := 0; i < size; i++ {
		mega[i] = zero
	}
	fifo, err := collections.NewFileFifo(fifoDir, 1)
	if err == nil {
		for i := 0; i < 3; i++ {
			err = fifo.Push(mega)
			if err != nil {
				t.Fatal(err)
			}
		}

		for i := 0; i < 3; i++ {
			data, err := fifo.Pop()
			if err != nil {
				t.Fatal(err)
			} else {
				if len(mega) != len(data) {
					t.Errorf("Retrived data size does not match! i: %v, data: %v", i, len(data))
				} else {
					for i, v := range mega {
						if v != data[i] {
							t.Errorf("Retrived data does not match!")
							return
						}
					}
				}
			}

		}
	} else {
		t.Fatal(err)
	}
}

func TestFilePushPop2(t *testing.T) {
	fifo, err := collections.NewFileFifo(fifoDir, 1)
	if err == nil {
		for i := 0; i < 100000; i++ {
			for _, m := range messages {
				err = fifo.Push([]byte(m))
				if err != nil {
					t.Fatal(err)
				}
			}

			for _, m := range messages {
				data, err := fifo.Pop()
				if err != nil {
					t.Fatal(err)
				} else {
					if m != string(data) {
						t.Errorf("Retrived data does not match! Failed comparing %s", m)
						return
					}
				}
			}
		}
	} else {
		t.Fatal(err)
	}
}

func TestFilePushPop3(t *testing.T) {
	fifo, err := collections.NewFileFifo(fifoDir, 1)
	if err == nil {
		for _, m := range messages {
			err = fifo.Push([]byte(m))
			if err != nil {
				t.Fatal(err)
			}
		}
		if fifo.Size() != int64(len(messages)) {
			t.Fatalf("Wrong fifo size. Expected %v got %v.\n", fifo.Size(), len(messages))
		}

		data, err := fifo.Peek()
		if messages[0] != string(data) {
			t.Fatalf("Peeked data does not match! Failed comparing %s\n", messages[0])
		}
		if fifo.Size() != int64(len(messages)) {
			t.Fatalf("Wrong fifo size after peek. Expected %v got %v.\n", fifo.Size(), len(messages))
		}

		for _, m := range messages {
			data, err = fifo.Pop()
			if err != nil {
				t.Fatal(err)
			} else {
				if m != string(data) {
					t.Errorf("Pop data does not match! Failed comparing %s\n", m)
					return
				}
			}
		}
		if fifo.Size() != int64(0) {
			t.Fatalf("Wrong fifo size after full pop. Expected 0 got %v.\n", fifo.Size())
		}

		data, err = fifo.Pop()
		if err != nil {
			t.Fatal(err)
		} else if data != nil {
			t.Errorf("Wrong Pop data. Expected nil got %s\n", string(data))
		}
		if fifo.Size() != int64(0) {
			t.Fatalf("Wrong fifo size after full pop + 1. Expected 0 got %v.\n", fifo.Size())
		}

		err = fifo.Clear()
		if err != nil {
			t.Fatal(err)
		}

	} else {
		t.Fatal(err)
	}
}

// TestBigPushPop1 will make some data be stored in memory and other be stored in file
func TestBigPushPop1(t *testing.T) {
	fifo, err := collections.NewBigFifo(3, fifoDir, 1, tk.GobCodec{}, (*string)(nil))
	if err == nil {
		for _, m := range messages {
			err = fifo.Push(m)
			if err != nil {
				t.Fatal(err)
			}
		}
		if fifo.Size() != int64(len(messages)) {
			t.Fatalf("Wrong fifo size. Expected %v got %v.\n", fifo.Size(), len(messages))
		}

		data := fifo.Peek()
		if messages[0] != data.(string) {
			t.Fatalf("Peeked data does not match! Failed comparing %s\n", messages[0])
		}
		if fifo.Size() != int64(len(messages)) {
			t.Fatalf("Wrong fifo size after peek. Expected %v got %v.\n", fifo.Size(), len(messages))
		}

		for _, m := range messages {
			data, err = fifo.Pop()
			if err != nil {
				t.Fatal(err)
			} else {
				if m != data.(string) {
					t.Errorf("Pop data does not match! Failed comparing %s\n", m)
					return
				}
			}
		}
		if fifo.Size() != int64(0) {
			t.Fatalf("Wrong fifo size after full pop. Expected 0 got %v.\n", fifo.Size())
		}
		err = fifo.Clear()
		if err != nil {
			t.Fatal(err)
		}

	} else {
		t.Fatal(err)
	}
}

func TestBigPushPop2(t *testing.T) {
	fifo, err := collections.NewBigFifo(3, fifoDir, 1, tk.GobCodec{}, (*string)(nil))
	if err == nil {
		msg := "hello"
		go func() {
			time.Sleep(time.Second)
			err := fifo.Push(msg)
			if err != nil {
				t.Error(err)
			}
		}()

		data, err := fifo.PopOrWait()
		if err != nil {
			t.Error(err)
		}
		if msg != data.(string) {
			t.Fatalf("Wrong data. Expected %s got %s.\n", msg, data.(string))
		}

	} else {
		t.Fatal(err)
	}
}
