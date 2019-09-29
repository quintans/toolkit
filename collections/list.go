package collections

import (
	"fmt"
	"reflect"
	"sort"

	. "github.com/quintans/toolkit"
)

type ArrayList struct {
	elements []interface{}
}

// check if it implements IList interface
var _ IList = &ArrayList{}

// check if it implements Base interface
var _ Base = &ArrayList{}

func NewArrayList() *ArrayList {
	m := new(ArrayList)
	m.Clear()
	return m
}

func (this *ArrayList) Clear() {
	this.elements = make([]interface{}, 0)
}

type ArrayListEnumerator struct {
	list   *ArrayList
	pos    int
	delPos int // enforce removal only if a Next() was called
}

func (this *ArrayListEnumerator) HasNext() bool {
	return this.pos < len(this.list.elements)
}

func (this *ArrayListEnumerator) Next() interface{} {
	if this.pos < this.list.Size() {
		k := this.list.elements[this.pos]
		this.pos++
		this.delPos = this.pos
		return k
	}

	return nil
}

func (this *ArrayListEnumerator) Peek() interface{} {
	if this.pos < this.list.Size() {
		return this.list.elements[this.pos]
	}

	return nil
}

func (this *ArrayListEnumerator) Remove() {
	if this.delPos > 0 && this.delPos <= this.list.Size() {
		this.list.DeleteAt(this.delPos - 1)
		//since this position was removed steps to the previous position
		this.pos -= 1
		// reset delete position
		this.delPos = 0
	}
}

// returns a function that in every call return the next value
// and a flag to see if a value was retrived, even if it was nil
func (this *ArrayList) Enumerator() Enumerator {
	return &ArrayListEnumerator{list: this}
}

func (this *ArrayList) Elements() []interface{} {
	data := make([]interface{}, len(this.elements), cap(this.elements))
	copy(data, this.elements)
	return data
}

func (this *ArrayList) AsSlice() interface{} {
	if len(this.elements) > 0 {
		typ := this.elements[0]
		t := reflect.TypeOf(typ)
		sliceT := reflect.SliceOf(t)
		sliceV := reflect.MakeSlice(sliceT, this.Size(), this.Size())
		for i, e := range this.elements {
			sliceV.Index(i).Set(reflect.ValueOf(e))
		}
		return sliceV.Interface()
	}
	return nil
}

func (this *ArrayList) Size() int {
	return len(this.elements)
}

func (this *ArrayList) Get(pos int) interface{} {
	return this.elements[pos]
}

func (this *ArrayList) Set(pos int, value interface{}) {
	this.elements[pos] = value
}

func (this *ArrayList) Add(data ...interface{}) bool {
	this.elements = append(this.elements, data...)
	return true
}

/*
func (this *ArrayList) AddAll(values Hasher) bool {
	valuesVal := reflect.ValueOf(values)
	values2 := make([]Hasher, valuesVal.Len())
	for i := range values2 {
		values2[i] = valuesVal.Index(i).Interface()
	}
	this.elements = append(this.elements, values2...)

	return true
}
*/

func (this *ArrayList) Sort(less func(a, b interface{}) bool) []interface{} {
	tmp := this.Elements()
	sort.Slice(tmp, func(x, y int) bool {
		return less(tmp[x], tmp[y])
	})
	return tmp
}

func (this *ArrayList) Find(value interface{}) (int, interface{}) {
	if Eq, isEq := value.(Equaler); isEq {
		for i, v := range this.elements {
			switch t := v.(type) {
			case Equaler:
				if t.Equals(Eq) {
					return i, v
				}
			default:
				if v == value {
					return i, v
				}

			}
		}
	} else {
		for i, v := range this.elements {
			if v == value {
				return i, v
			}
		}
	}

	return -1, nil
}

func (this *ArrayList) Contains(value interface{}) bool {
	k, _ := this.Find(value)
	if k > -1 {
		return true
	}
	return false
}

func (this *ArrayList) Delete(value interface{}) bool {
	k, _ := this.Find(value)
	return this.DeleteAt(k)
}

func (this *ArrayList) DeleteAt(pos int) bool {
	if pos >= 0 && pos < this.Size() {
		// since the slice has a non-primitive, we have to zero it
		copy(this.elements[pos:], this.elements[pos+1:])
		this.elements[len(this.elements)-1] = nil // zero it
		this.elements = this.elements[:len(this.elements)-1]

		return true
	}
	return false
}

func (this *ArrayList) String() string {
	return fmt.Sprint(this.elements)
}

func (this *ArrayList) Clone() interface{} {
	return &ArrayList{this.Elements()}
}

func (this *ArrayList) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *ArrayList:
		// check size
		if this.Size() != t.Size() {
			return false
		}

		// check if it has all entries (keys)
		for k, v := range this.elements {
			if v != t.elements[k] {
				return false
			}
		}

		return true
	}
	return false
}

func (this *ArrayList) HashCode() int {
	panic("ArrayList.HashCode not implemented")
}
