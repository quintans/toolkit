// DRAFT ONLY. THIS IS STILL IN DEVELOPEMENT. NO TESTS WHERE MADE

// the Idea is to have a FIFO with a windowing (circular) feature.
// If the max size is reached, the element oldest element will be removed.

package collections

type Element struct {
	next     *Element
	previous *Element
	Value    interface{}
}

func (this *Element) Next() *Element {
	return this.next
}

func (this *Element) Previous() *Element {
	return this.previous
}

type Fifo struct {
	head     *Element
	tail     *Element
	size     int
	capacity int
}

func NewFifo(capacity int) *Fifo {
	this := new(Fifo)
	this.capacity = capacity
	return this
}

func (this *Fifo) Capacity(capacity int) {
	if capacity > 0 {
		for capacity < this.size {
			this.Pop()
		}
	}
	if capacity < 0 {
		capacity = 0
	} else {
		this.capacity = capacity
	}
}

func (this *Fifo) Size() int {
	return this.size
}

func (this *Fifo) Clear() {
	this.head = nil
	this.tail = nil
	this.size = 0
}

func (this *Fifo) Push(value interface{}) *Element {
	e := new(Element)
	old := this.head

	if old == nil {
		this.head = e
		this.tail = e
	} else {
		this.head = e
		old.previous = e
		e.next = old
	}
	// check capacity limit
	if this.capacity == 0 || this.size < this.capacity {
		this.size++
	} else {
		this.Pop()
	}
	return e
}

func (this *Fifo) Pop() interface{} {
	old := this.tail
	if old != nil {
		if old.previous != nil {
			old.previous.next = nil
		}
		old.next = nil
		old.previous = nil
		this.size--
		return old.Value
	}
	return nil
}

func (this *Fifo) Peek() *Element {
	return this.tail
}

// searches from the last position
func (this *Fifo) PeekAt(pos int) *Element {
	if pos < 0 || pos >= this.size {
		return nil
	}

	half := this.size / 2
	if pos < half {
		var i = 0
		// search from TAIL
		for e := this.tail; e != nil; e = e.previous {
			if i == pos {
				return e
			}
			i++
		}
	} else {
		// search from HEAD
		var i = this.size - 1
		for e := this.head; e != nil; e = e.next {
			if i == pos {
				return e
			}
			i--
		}
	}
	return nil
}

func (this *Fifo) First() *Element {
	return this.head
}
