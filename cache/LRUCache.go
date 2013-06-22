package cache

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	sync.Mutex

	entries *list.List
	table   map[string]*list.Element

	capacity int
}

type entry struct {
	key   string
	value interface{}
}

func NewLRUCache(capacity int) *LRUCache {
	this := new(LRUCache)
	this.capacity = capacity
	this.entries = list.New()
	this.table = make(map[string]*list.Element)
	return this
}

func (this *LRUCache) GetIfPresent(key string) (interface{}, bool) {
	this.Lock()
	defer this.Unlock()

	return this.get(key)
}

func (this *LRUCache) get(key string) (interface{}, bool) {
	element := this.table[key]
	if element != nil {
		// move to front
		this.entries.MoveToFront(element)
		return element.Value.(*entry).value, true
	}
	return nil, false
}

func (this *LRUCache) Get(key string, callback func() interface{}) (interface{}, bool) {
	this.Lock()
	defer this.Unlock()

	value, ok := this.get(key)
	if ok {
		// returns true indicating that it not found in the cache
		return value, true
	}

	value = callback()
	this.add(key, value)
	// returns false indicating that it was not found in the cache and was created by the callback
	return value, false
}

func (this *LRUCache) Put(key string, value interface{}) {
	this.Lock()
	defer this.Unlock()

	element := this.table[key]
	if element != nil {
		e := element.Value.(*entry)
		e.value = value
		this.entries.MoveToFront(element)
	} else {
		this.add(key, value)
	}
}

func (this *LRUCache) add(key string, value interface{}) {
	if this.entries.Len() == this.capacity {
		// if at full capacity recycle last element
		element := this.entries.Back()
		e := element.Value.(*entry)
		element.Value = &entry{key, value}
		this.entries.MoveToFront(element)

		delete(this.table, e.key)
		this.table[key] = element
	} else {
		this.table[key] = this.entries.PushFront(&entry{key, value})
	}
}

func (this *LRUCache) Delete(key string) {
	this.Lock()
	defer this.Unlock()

	element := this.table[key]
	if element != nil {
		// remove from list
		this.entries.Remove(element)
	}
	delete(this.table, key)
}
