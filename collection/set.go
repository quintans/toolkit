package collections

import (
	. "github.com/quintans/toolkit"
	"reflect"
)

// == HashSet ==

type HashSet struct {
	entries Map
}

func NewHashSet() *HashSet {
	s := new(HashSet)
	s.Clear()
	return s
}

// check if it implements Collection interface
var _ Collection = &HashSet{}

// check if it implements Base interface
var _ Base = &HashSet{}

func (this *HashSet) Clear() {
	this.entries = NewHashMap()
}

func (this *HashSet) Size() int {
	return this.entries.Size()
}

func (this *HashSet) Contains(value interface{}) bool {
	hasher := value.(Hasher)
	_, ok := this.entries.Get(hasher)
	return ok
}

type HashSetEnumerator struct {
	iterator Iterator
}

func (this *HashSetEnumerator) HasNext() bool {
	return this.iterator.HasNext()
}

func (this *HashSetEnumerator) Next() interface{} {
	kv := this.iterator.Next()
	if kv != nil {
		return kv.Key
	}
	return nil
}

func (this *HashSetEnumerator) Peek() interface{} {
	kv := this.iterator.Peek()
	if kv != nil {
		return kv.Key
	}
	return nil
}

func (this *HashSetEnumerator) Remove() {
	this.iterator.Remove()
}

// returns a function that in every call return the next value
// if no value was retrived
func (this *HashSet) Enumerator() Enumerator {
	return &HashSetEnumerator{this.entries.Iterator()}
}

// Add if it does not exists
func (this *HashSet) add(value Hasher) bool {
	_, ok := this.entries.Get(value)
	if !ok {
		this.entries.Put(value, struct{}{})
	}

	return !ok
}

func (this *HashSet) Add(data ...interface{}) bool {
	changed := false
	for _, v := range data {
		h := v.(Hasher)
		if this.add(h) {
			changed = true
		}
	}
	return changed
}

func (this *HashSet) Delete(value interface{}) bool {
	h := value.(Hasher)
	_, ok := this.entries.Get(h)
	if ok {
		this.entries.Delete(h)
	}

	return ok
}

func (this *HashSet) Sort(greater func(a, b interface{}) bool) []interface{} {
	tmp := this.Elements()
	Sort(tmp, greater)
	return tmp
}

func (this *HashSet) Elements() []interface{} {
	data := make([]interface{}, this.entries.Size())
	i := 0
	for it := this.entries.Iterator(); it.HasNext(); i++ {
		data[i] = it.Next().Key
	}
	return data
}

func (this *HashSet) AsSlice() interface{} {
	it := this.entries.Iterator()
	if it.HasNext() {
		typ := it.Peek().Key
		t := reflect.TypeOf(typ)
		sliceT := reflect.SliceOf(t)
		sliceV := reflect.MakeSlice(sliceT, this.Size(), this.Size())
		i := 0
		for it := this.entries.Iterator(); it.HasNext(); i++ {
			v := reflect.ValueOf(it.Next().Key)
			sliceV.Index(i).Set(v)
		}
		return sliceV.Interface()
	}
	return nil
}

func (this *HashSet) String() string {
	s := new(StrBuffer)
	s.Add("[")
	for it := this.entries.Iterator(); it.HasNext(); {
		s.Add(it.Next().Key)
	}
	s.Add("]")

	return s.String()
}

func (this *HashSet) Clone() interface{} {
	x := NewHashSet()
	if c, ok := this.entries.(Clonable); ok {
		x.entries = c.Clone().(Map)
	}
	return x
}

func (this *HashSet) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *HashSet:
		// check size
		if this.Size() != t.Size() {
			return false
		}

		for it1, it2 := this.entries.Iterator(), t.entries.Iterator(); it1.HasNext() && it2.HasNext(); {
			kv1 := it1.Next().Key
			kv2 := it2.Next().Key
			if !kv1.Equals(kv2) {
				return false
			}
		}

		return true
	}
	return false
}

func (this *HashSet) HashCode() int {
	panic("HashSet.HashCode not implemented")
}

//== LinkedHashSet ==

type LinkedHashSet struct {
	HashSet
}

// check if it implements Collection interface
var _ Collection = &LinkedHashSet{}

// check if it implements Base interface
var _ Base = &LinkedHashSet{}

func NewLinkedHashSet() *LinkedHashSet {
	s := new(LinkedHashSet)
	s.Clear()
	return s
}

func (this *LinkedHashSet) Clear() {
	this.entries = NewLinkedHashMap()
}

func (this *LinkedHashSet) HashCode() int {
	panic("LinkedHashSet.HashCode not implemented")
}
