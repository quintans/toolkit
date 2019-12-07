package collections

import (
	"reflect"
	"sort"

	. "github.com/quintans/toolkit"
)

// == HashSet ==

type HashSet struct {
	entries *HashMap
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

func (this *HashSet) Empty() bool {
	return this.Size() == 0
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

func (this *HashSet) Add(data ...Hasher) bool {
	changed := false
	for _, v := range data {
		if this.add(v) {
			changed = true
		}
	}
	return changed
}

func (this *HashSet) Delete(value interface{}) bool {
	h, ok := value.(Hasher)
	if !ok {
		return false
	}
	old := this.entries.Delete(h)
	return old != nil
}

func (this *HashSet) Sort(less func(a, b interface{}) bool) []interface{} {
	tmp := make([]interface{}, this.entries.Size())
	i := 0
	this.ForEach(func(e interface{}) {
		tmp[i] = e
		i++
	})

	sort.Slice(tmp, func(x, y int) bool {
		return less(tmp[x], tmp[y])
	})

	elems := make([]interface{}, len(tmp))
	for k, v := range tmp {
		elems[k] = v.(interface{})
	}

	return elems
}

func (this *HashSet) Elements() []interface{} {
	data := make([]interface{}, this.entries.Size())
	i := 0
	this.ForEach(func(e interface{}) {
		data[i] = e
		i++
	})
	return data
}

func (this *HashSet) ForEach(fn func(interface{})) {
	for _, entry := range this.entries.table {
		for ; entry != nil; entry = entry.next {
			fn(entry.key)
		}
	}
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
	this.ForEach(func(v interface{}) {
		s.Add(v)
	})
	s.Add("]")

	return s.String()
}

func (this *HashSet) Clone() interface{} {
	x := NewHashSet()
	x.entries = this.entries.Clone().(*HashMap)
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
	lhm *LinkedHashMap
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
	this.lhm = NewLinkedHashMap()
}

func (this *LinkedHashSet) String() string {
	s := new(StrBuffer)
	s.Add("[")
	this.lhm.ForEach(func(kv *KeyValue) {
		s.Add(kv.Key)
	})
	s.Add("]")

	return s.String()
}

func (this *LinkedHashSet) Clone() interface{} {
	x := NewLinkedHashSet()
	x.lhm = this.lhm.Clone().(*LinkedHashMap)
	return x
}

func (this *LinkedHashSet) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *LinkedHashSet:
		return this.lhm.Equals(t.lhm)
	}
	return false
}

func (this *LinkedHashSet) HashCode() int {
	panic("LinkedHashSet.HashCode not implemented")
}

func (this *LinkedHashSet) Size() int {
	return this.lhm.Size()
}

func (this *LinkedHashSet) Empty() bool {
	return this.Size() == 0
}

func (this *LinkedHashSet) Contains(value interface{}) bool {
	hasher := value.(Hasher)
	_, ok := this.lhm.Get(hasher)
	return ok
}

// returns a function that in every call return the next value
// if no value was retrived
func (this *LinkedHashSet) Enumerator() Enumerator {
	return &HashSetEnumerator{this.lhm.Iterator()}
}

// Add if it does not exists
func (this *LinkedHashSet) add(value Hasher) bool {
	_, ok := this.lhm.Get(value)
	if !ok {
		this.lhm.Put(value, struct{}{})
	}

	return !ok
}

func (this *LinkedHashSet) Add(data ...Hasher) bool {
	changed := false
	for _, v := range data {
		if this.add(v) {
			changed = true
		}
	}
	return changed
}

func (this *LinkedHashSet) Delete(value interface{}) bool {
	h, ok := value.(Hasher)
	if !ok {
		return false
	}
	old := this.lhm.Delete(h)
	return old != nil
}

func (this *LinkedHashSet) Sort(less func(a, b interface{}) bool) []interface{} {
	tmp := make([]interface{}, this.lhm.Size())
	i := 0
	this.ForEach(func(e interface{}) {
		tmp[i] = e
		i++
	})

	sort.Slice(tmp, func(x, y int) bool {
		return less(tmp[x], tmp[y])
	})

	elems := make([]interface{}, len(tmp))
	for k, v := range tmp {
		elems[k] = v.(interface{})
	}

	return elems
}

func (this *LinkedHashSet) Elements() []interface{} {
	data := make([]interface{}, this.lhm.Size())
	i := 0
	this.lhm.ForEach(func(kv *KeyValue) {
		data[i] = kv.Key
		i++
	})
	return data
}

func (this *LinkedHashSet) ForEach(fn func(interface{})) {
	this.lhm.ForEach(func(kv *KeyValue) {
		fn(kv.Key)
	})
}

func (this *LinkedHashSet) AsSlice() interface{} {
	it := this.lhm.Iterator()
	if it.HasNext() {
		typ := it.Peek().Key
		t := reflect.TypeOf(typ)
		sliceT := reflect.SliceOf(t)
		sliceV := reflect.MakeSlice(sliceT, this.Size(), this.Size())
		i := 0
		this.lhm.ForEach(func(kv *KeyValue) {
			v := reflect.ValueOf(it.Next().Key)
			sliceV.Index(i).Set(v)
			i++
		})
		return sliceV.Interface()
	}
	return nil
}
