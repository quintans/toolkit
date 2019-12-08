package collections

import (
	. "github.com/quintans/toolkit"
	"github.com/quintans/toolkit/log"
)

var logger = log.LoggerFor("github.com/quintans/toolkit/collections")

type Collection interface {
	Base

	Size() int
	Empty() bool
	Contains(value interface{}) bool
	Enumerator() Enumerator
	Elements() []interface{}
	AsSlice() interface{} // returns elements in an array. ex: []int
	Sort(greater func(a, b interface{}) bool) []interface{}

	Clear()
	Delete(key interface{}) bool
}

type IList interface {
	Collection

	Get(pos int) interface{}
	First(value interface{}) (int, interface{})
	Find(func(interface{}) bool) (int, interface{})
	ForEach(func(int, interface{}))

	Add(data ...interface{}) bool
	Set(pos int, value interface{})
	DeleteAt(pos int) bool
	Insert(pos int, data ...interface{})
}

type ISet interface {
	Collection

	Add(data ...Hasher) bool
	ForEach(func(interface{}))
}

type Enumerator interface {
	HasNext() bool
	Next() interface{}
	Peek() interface{}
	Remove()
}

type Map interface {
	Base

	Put(key Hasher, Value interface{}) interface{}
	Get(key Hasher) (interface{}, bool)
	Delete(key Hasher) interface{}

	Size() int
	Clear()

	Iterator() Iterator
	Elements() []*KeyValue
	Values() []interface{}
	ForEach(func(*KeyValue))
}

type Iterator interface {
	HasNext() bool
	Next() *KeyValue
	Peek() *KeyValue
	Remove()
}
