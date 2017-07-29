package collections

import (
	. "github.com/quintans/toolkit"
	"github.com/quintans/toolkit/log"
)

var logger = log.LoggerFor("github.com/quintans/toolkit/collection")

type Collection interface {
	Base

	Size() int
	Clear()
	Contains(value interface{}) bool
	Add(data ...interface{}) bool
	Delete(key interface{}) bool

	Enumerator() Enumerator
	Elements() []interface{}
	AsSlice() interface{} // returns elements in an array. ex: []int
	Sort(greater func(a, b int) bool) []interface{}
}

type IList interface {
	Collection

	Get(pos int) interface{}
	Set(pos int, value interface{})
	Find(value interface{}) (int, interface{})
	DeleteAt(pos int) bool
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
}

type Iterator interface {
	HasNext() bool
	Next() *KeyValue
	Peek() *KeyValue
	Remove()
}
