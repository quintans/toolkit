package collections

import (
	"fmt"
	"reflect"
	"sort"

	. "github.com/quintans/toolkit"
)

type ArrayList struct {
	elements []interface{}
	hash     int
}

// check if it implements IList interface
var _ IList = &ArrayList{}

// check if it implements Base interface
var _ Base = &ArrayList{}

func NewArrayList(elems ...interface{}) *ArrayList {
	return &ArrayList{elems, 0}
}

func (a *ArrayList) Clear() {
	a.elements = []interface{}{}
}

type ArrayListEnumerator struct {
	list   *ArrayList
	pos    int
	delPos int // enforce removal only if a Next() was called
}

func (e *ArrayListEnumerator) HasNext() bool {
	return e.pos < e.list.Size()
}

func (e *ArrayListEnumerator) Next() interface{} {
	if e.pos < e.list.Size() {
		k := e.list.Get(e.pos)
		e.pos++
		e.delPos = e.pos
		return k
	}

	return nil
}

func (e *ArrayListEnumerator) Peek() interface{} {
	if e.pos < e.list.Size() {
		return e.list.Get(e.pos)
	}

	return nil
}

func (e *ArrayListEnumerator) Remove() {
	if e.delPos > 0 && e.delPos <= e.list.Size() {
		e.list.DeleteAt(e.delPos - 1)
		//since this position was removed steps to the previous position
		e.pos -= 1
		// reset delete position
		e.delPos = 0
	}
}

// returns a function that in every call return the next value
// and a flag to see if a value was retrived, even if it was nil
func (a *ArrayList) Enumerator() Enumerator {
	return &ArrayListEnumerator{list: a}
}

func (a *ArrayList) Elements() []interface{} {
	return a.elements
}

func (a *ArrayList) AsSlice() interface{} {
	if len(a.elements) > 0 {
		typ := a.elements[0]
		t := reflect.TypeOf(typ)
		sliceT := reflect.SliceOf(t)
		sliceV := reflect.MakeSlice(sliceT, a.Size(), a.Size())
		for i, e := range a.elements {
			sliceV.Index(i).Set(reflect.ValueOf(e))
		}
		return sliceV.Interface()
	}
	return nil
}

func (a *ArrayList) Size() int {
	return len(a.elements)
}

func (a *ArrayList) Empty() bool {
	return a.Size() == 0
}

func (a *ArrayList) Get(pos int) interface{} {
	return a.elements[pos]
}

func (a *ArrayList) Set(pos int, value interface{}) {
	a.elements[pos] = value
}

func (a *ArrayList) Add(data ...interface{}) bool {
	a.elements = append(a.elements, data...)
	return true
}

func (a *ArrayList) Insert(i int, data ...interface{}) {
	if len(a.elements) == 1 {
		// if data has one element it is more efficient this way
		tmp := append(a.elements, nil)
		copy(tmp[i+1:], tmp[i:])
		tmp[i] = data[0]
		a.elements = tmp
	} else {
		arr := a.elements
		tmp := append(data, arr[i:]...)
		a.elements = append(arr[:i], tmp...)
	}
}

func (a *ArrayList) Sort(less func(a, b interface{}) bool) []interface{} {
	tmp := Clone(a.elements)
	sort.Slice(tmp, func(x, y int) bool {
		return less(tmp[x], tmp[y])
	})
	return tmp
}

func (a *ArrayList) First(value interface{}) (int, interface{}) {
	if eq, isEq := value.(Equaler); isEq {
		for i, v := range a.elements {
			switch t := v.(type) {
			case Equaler:
				if t.Equals(eq) {
					return i, v
				}
			default:
				if v == value {
					return i, v
				}

			}
		}
	} else {
		for i, v := range a.elements {
			if v == value {
				return i, v
			}
		}
	}

	return -1, nil
}

func (a *ArrayList) Find(fn func(interface{}) bool) (int, interface{}) {
	for i, v := range a.elements {
		if fn(v) {
			return i, v
		}
	}
	return -1, nil
}

func (a *ArrayList) Contains(value interface{}) bool {
	k, _ := a.First(value)
	if k > -1 {
		return true
	}
	return false
}

func (a *ArrayList) Delete(value interface{}) bool {
	k, _ := a.First(value)
	return a.DeleteAt(k)
}

func (a *ArrayList) DeleteAt(pos int) bool {
	if pos >= 0 && pos < a.Size() {
		// since the slice has a non-primitive, we have to zero it
		arr := a.elements
		copy(arr[pos:], arr[pos+1:])
		arr[len(arr)-1] = nil // zero it
		a.elements = arr[:len(arr)-1]

		return true
	}
	return false
}

func (a *ArrayList) ForEach(fn func(int, interface{})) {
	for k, v := range a.elements {
		fn(k, v)
	}
}

func (a *ArrayList) String() string {
	return fmt.Sprint(a)
}

func (a *ArrayList) Clone() interface{} {
	return Clone(a.elements)
}

func Clone(src []interface{}) []interface{} {
	dest := make([]interface{}, len(src), cap(src))
	copy(dest, src)
	return dest
}

func (a *ArrayList) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case ArrayList:
		// check size
		if a.Size() != t.Size() {
			return false
		}

		// check if it has all entries (keys)
		for k, v := range a.elements {
			if v != t.elements[k] {
				return false
			}
		}

		return true
	}
	return false
}

func (a *ArrayList) HashCode() int {
	if a.hash == 0 {
		result := HashType(HASH_SEED, a)
		result = Hash(result, a.elements...)
		a.hash = result
	}
	return a.hash
}
