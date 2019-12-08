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
	Clear()
	Contains(value interface{}) bool
	Delete(key interface{}) bool

	Enumerator() Enumerator
	Elements() []interface{}
	AsSlice() interface{} // returns elements in an array. ex: []int
	Sort(greater func(a, b interface{}) bool) []interface{}
}

type IList interface {
	Collection

	Add(data ...interface{}) bool
	Get(pos int) interface{}
	Set(pos int, value interface{})
	First(value interface{}) (int, interface{})
	Find(func(interface{}) bool) (int, interface{})
	DeleteAt(pos int) bool
	Insert(pos int, data ...interface{})
	ForEach(func(int, interface{}))
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
