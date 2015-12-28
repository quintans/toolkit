package collections

import (
	"testing"

	tk "github.com/quintans/toolkit"
	"github.com/quintans/toolkit/ext"
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
	fifo, err := NewFileFifo(fifoDir, 1)
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
				for i, v := range mega {
					if v != data[i] {
						t.Errorf("Retrived data does not match!")
						return
					}
				}
			}

		}
	} else {
		t.Fatal(err)
	}
}

func TestFilePushPop2(t *testing.T) {
	fifo, err := NewFileFifo(fifoDir, 1)
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
	fifo, err := NewFileFifo(fifoDir, 1)
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
		err = fifo.Clear()
		if err != nil {
			t.Fatal(err)
		}

	} else {
		t.Fatal(err)
	}
}

func TestBigPushPop3(t *testing.T) {
	fifo, err := NewBigFifo(3, fifoDir, 1, tk.GobCodec{}, func() interface{} { return ext.String("") })
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
		err = fifo.Clear()
		if err != nil {
			t.Fatal(err)
		}

	} else {
		t.Fatal(err)
	}
}

// TODO test concurrent access
