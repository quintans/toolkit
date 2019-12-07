package collections

import (
	"fmt"

	. "github.com/quintans/toolkit"
)

type KeyValue struct {
	Key   Hasher
	Value interface{}
}

func (this *KeyValue) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *KeyValue:
		return this.Key.Equals(t.Key) && this.Value == t.Value
	}
	return false
}

func (this *KeyValue) String() string {
	return fmt.Sprintf("{%s, %s}", this.Key, this.Value)
}

func (this *KeyValue) HashCode() int {
	result := HashType(HASH_SEED, this)
	result = HashInt(result, this.Key.HashCode())
	result = Hash(result, this.Value)
	return result
}

// == HashMap ==

// check if it implements Map interface
var _ Map = &HashMap{}

// check if it implements Base interface
var _ Base = &HashMap{}

func (this *HashMap) String() string {
	s := new(StrBuffer)
	s.Add("[")
	for it := this.Iterator(); it.HasNext(); {
		s.Add(it.Next())
	}
	s.Add("]")

	return s.String()
}

func (this *HashMap) Clone() interface{} {
	m := NewHashMap()
	for it := this.Iterator(); it.HasNext(); {
		kv := it.Next()
		m.Put(kv.Key, kv.Value)
	}

	return m
}

func (this *HashMap) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *HashMap:
		// check size
		if this.Size() != t.Size() {
			return false
		}

		for it1, it2 := this.Iterator(), t.Iterator(); it1.HasNext() && it2.HasNext(); {
			kv1 := it1.Next()
			kv2 := it2.Next()
			if !kv1.Equals(kv2) {
				return false
			}
		}

		return true
	}
	return false
}

type linkedEntry struct {
	key   Hasher
	value interface{}
	next  *linkedEntry
}

const default_table_size = 16

type HashMap struct {
	maxThreshold float32
	minThreshold float32
	maxSize      int
	minSize      int
	tableSize    int
	size         int
	table        []*linkedEntry
}

func NewHashMap() *HashMap {
	hm := new(HashMap)
	hm.Clear()
	return hm
}

func (this *HashMap) Clear() {
	this.maxThreshold = 0.75
	this.minThreshold = 0.25
	this.tableSize = default_table_size
	this.maxSize = int(default_table_size * this.maxThreshold)
	this.minSize = int(default_table_size * this.minThreshold)
	this.size = 0
	this.table = make([]*linkedEntry, default_table_size)
}

func (this *HashMap) resize(newSize int) {
	oldTableSize := this.tableSize
	this.tableSize = newSize
	this.maxSize = int(float32(this.tableSize) * this.maxThreshold)
	this.minSize = int(float32(this.tableSize) * this.minThreshold)
	oldTable := this.table
	this.table = make([]*linkedEntry, this.tableSize)
	this.size = 0
	for hash := 0; hash < oldTableSize; hash++ {
		if oldTable[hash] != nil {
			entry := oldTable[hash]
			for entry != nil {
				this.Put(entry.key, entry.value)
				entry = entry.next
			}
			// discard
			oldTable[hash] = nil
		}
	}
}

func (this *HashMap) index(key Hasher) int {
	return ((key.HashCode() & 0x7FFFFFFF) % this.tableSize)
}

func (this *HashMap) Get(key Hasher) (interface{}, bool) {
	hash := this.index(key)
	if this.table[hash] != nil {
		entry := this.table[hash]
		for entry != nil && !entry.key.Equals(key) {
			entry = entry.next
		}
		if entry != nil {
			return entry.value, true
		}
	}
	return nil, false
}

func (this *HashMap) Put(key Hasher, value interface{}) interface{} {
	var old interface{} = nil
	hash := this.index(key)
	if this.table[hash] == nil {
		this.table[hash] = &linkedEntry{key, value, nil}
		this.size++
	} else {
		entry := this.table[hash]
		var prevEntry *linkedEntry = nil
		for entry != nil && !entry.key.Equals(key) {
			prevEntry = entry
			entry = entry.next
		}
		if entry == nil {
			prevEntry.next = &linkedEntry{key, value, nil}
			this.size++
		} else {
			old = entry.value
			entry.value = value
		}
	}
	if this.size >= this.maxSize {
		this.resize(this.tableSize * 2)
	}

	return old
}

func (this *HashMap) Delete(key Hasher) interface{} {
	var old interface{} = nil
	hash := this.index(key)
	if entry := this.table[hash]; entry != nil {
		var prevEntry *linkedEntry = nil
		for entry != nil && !entry.key.Equals(key) {
			prevEntry = entry
			entry = entry.next
		}
		if entry != nil {
			old = entry.value
			if prevEntry == nil {
				this.table[hash] = entry.next
			} else {
				prevEntry.next = entry.next
			}
			this.size--
		}
	}

	if this.size <= this.minSize && this.tableSize > default_table_size {
		this.resize(this.tableSize / 2)
	}

	return old
}

func (this *HashMap) Size() int {
	return this.size
}

func (this *HashMap) Empty() bool {
	return this.Size() == 0
}

func (this *HashMap) Elements() []*KeyValue {
	data := make([]*KeyValue, this.size)
	i := 0
	this.ForEach(func(kv *KeyValue) {
		data[i] = kv
		i++
	})
	return data
}

func (this *HashMap) ForEach(fn func(*KeyValue)) {
	for _, entry := range this.table {
		for ; entry != nil; entry = entry.next {
			fn(&KeyValue{entry.key, entry.value})
		}
	}
}

func (this *HashMap) Values() []interface{} {
	data := make([]interface{}, this.size)
	i := 0
	for it := this.Iterator(); it.HasNext(); {
		data[i] = it.Next().Value
		i++
	}
	return data
}

type HashMapIterator struct {
	hashmap   *HashMap
	hash      int
	prevEntry *linkedEntry
	entry     *linkedEntry
}

func (this *HashMapIterator) HasNext() bool {
	return this.entry != nil
}

func (this *HashMapIterator) Next() *KeyValue {
	if this.entry != nil {
		kv := &KeyValue{this.entry.key, this.entry.value}
		this.next()
		return kv
	}
	return nil
}

func (this *HashMapIterator) Peek() *KeyValue {
	if this.entry != nil {
		return &KeyValue{this.entry.key, this.entry.value}
	}
	return nil
}

func (this *HashMapIterator) next() {
	max := len(this.hashmap.table)
	var aEntry *linkedEntry = nil
	for i := this.hash; aEntry == nil && i < max; i++ {
		if this.entry == nil {
			this.prevEntry = nil
			this.entry = this.hashmap.table[i]
		} else {
			this.prevEntry = this.entry
			this.entry = this.entry.next
		}
		aEntry = this.entry
		if this.entry != nil {
			this.hash = i
		}
	}
}

func (this *HashMapIterator) Remove() {
	if this.entry != nil {
		if this.prevEntry == nil {
			this.hashmap.table[this.hash] = this.entry.next
		} else {
			this.prevEntry.next = this.entry.next
		}
		this.hashmap.size--
	}
}

// returns a function that in every call return the next value
// if key is null, no value was retrieved
func (this *HashMap) Iterator() Iterator {
	it := &HashMapIterator{hashmap: this}
	// initiates
	it.next()
	return it
}

func (this *HashMap) HashCode() int {
	panic("HashMap.HashCode not implemented")
}

//== LinkedHashMap ==

type LinkedHashMap struct {
	keyOrder []*keyEntry
	entries  *HashMap
}

type keyEntry struct {
	key   Hasher
	entry *entry
}

type entry struct {
	index int
	value interface{}
}

// check if it implements Map interface
var _ Map = &LinkedHashMap{}

// check if it implements Base interface
var _ Base = &LinkedHashMap{}

func NewLinkedHashMap() *LinkedHashMap {
	m := new(LinkedHashMap)
	m.Clear()
	return m
}

func (this *LinkedHashMap) Clear() {
	this.keyOrder = []*keyEntry{}
	this.entries = NewHashMap()
}

func (this *LinkedHashMap) Size() int {
	return this.entries.Size()
}

func (this *LinkedHashMap) Empty() bool {
	return this.Size() == 0
}

func (this *LinkedHashMap) Get(key Hasher) (interface{}, bool) {
	if e, ok := this.entries.Get(key); ok {
		if T, isT := e.(*entry); isT {
			return T.value, true
		}
	}

	return nil, false
}

type LinkedHashMapIterator struct {
	hashmap *LinkedHashMap
	pos     int
}

func (this *LinkedHashMapIterator) HasNext() bool {
	return this.pos < this.hashmap.Size()
}

func (this *LinkedHashMapIterator) Next() *KeyValue {
	k := this.Peek()
	this.pos++
	return k
}

func (this *LinkedHashMapIterator) Peek() *KeyValue {
	if this.pos < this.hashmap.Size() {
		k := this.hashmap.keyOrder[this.pos]
		return &KeyValue{k.key, k.entry.value}
	}

	return nil
}

func (this *LinkedHashMapIterator) Remove() {
	max := this.hashmap.Size()
	if this.pos > 0 && this.pos < max {
		k := this.hashmap.keyOrder[this.pos-1]
		this.hashmap.Delete(k.key)
	}
}

// returns a function that in every call return the next value
// if key is null, no value was retrieved
func (this *LinkedHashMap) Iterator() Iterator {
	return &LinkedHashMapIterator{this, 0}
}

func (this *LinkedHashMap) Put(key Hasher, value interface{}) interface{} {
	tmp := &entry{0, value}
	old := this.entries.Put(key, tmp)
	if old == nil {
		// was previously empty, so update the index
		tmp.index = len(this.keyOrder)
		this.keyOrder = append(this.keyOrder, &keyEntry{key, tmp})
	}

	return nil
}

func (this *LinkedHashMap) Delete(key Hasher) interface{} {
	e := this.entries.Delete(key)
	if T, isT := e.(*entry); isT {
		// since the slice has a non-primitive, we have to zero it
		copy(this.keyOrder[T.index:], this.keyOrder[T.index+1:])
		this.keyOrder[len(this.keyOrder)-1] = nil // zero it
		this.keyOrder = this.keyOrder[:len(this.keyOrder)-1]

		return T.value
	}

	return nil
}

func (this *LinkedHashMap) Elements() []*KeyValue {
	data := make([]*KeyValue, len(this.keyOrder))
	for i := 0; i < len(data); i++ {
		ko := this.keyOrder[i]
		data[i] = &KeyValue{ko.key, ko.entry.value}
	}
	return data
}

func (this *LinkedHashMap) Values() []interface{} {
	data := make([]interface{}, len(this.keyOrder))
	for i := 0; i < len(data); i++ {
		ko := this.keyOrder[i]
		data[i] = ko.entry.value
	}
	return data
}

func (this *LinkedHashMap) ForEach(fn func(*KeyValue)) {
	for _, ko := range this.keyOrder {
		fn(&KeyValue{ko.key, ko.entry.value})
	}
}

func (this *LinkedHashMap) String() string {
	s := new(StrBuffer)
	s.Add("[")
	for it := this.Iterator(); it.HasNext(); {
		s.Add(it.Next())
	}
	s.Add("]")

	return s.String()
}

func (this *LinkedHashMap) Clone() interface{} {
	m := NewHashMap()
	for it := this.Iterator(); it.HasNext(); {
		kv := it.Next()
		m.Put(kv.Key, kv.Value)
	}

	return m
}

func (this *LinkedHashMap) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *LinkedHashMap:
		// check size
		if this.Size() != t.Size() {
			return false
		}

		for it1, it2 := this.Iterator(), t.Iterator(); it1.HasNext() && it2.HasNext(); {
			kv1 := it1.Next()
			kv2 := it2.Next()
			if !kv1.Equals(kv2) {
				return false
			}
		}

		return true
	}
	return false
}

func (this *LinkedHashMap) HashCode() int {
	panic("LinkedHashMap.HashCode not implemented")
}
