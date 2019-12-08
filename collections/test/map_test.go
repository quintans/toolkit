package test

import (
	"fmt"
	"testing"

	. "github.com/quintans/toolkit/collections"
	. "github.com/quintans/toolkit/ext"
)

const value1 = "World"

func TestPutAndGet(t *testing.T) {
	hmap := NewHashMap()
	k := Str("Hello")
	hmap.Put(k, value1)
	v, ok := hmap.Get(k)
	if !ok || value1 != v {
		t.Error("Expected "+value1+", got ", v)
	}
	v, ok = hmap.Get(Str("lixo"))
	if ok {
		t.Error("Expected to find nothing, got ", v)
	}
}

func TestResize(t *testing.T) {
	hmap := NewHashMap()
	loop := 20
	// Insert
	for i := 0; i < loop; i++ {
		hmap.Put(Str("Hello"+string(i)), i*10)
	}
	if hmap.Size() != loop {
		t.Error("Expected "+string(loop)+", got ", hmap.Size())
	}
	// Check
	for i := 0; i < loop; i++ {
		v, _ := hmap.Get(Str("Hello" + string(i)))
		k := i * 10
		if k != v {
			t.Error("Expected "+string(k)+", got ", v)
		}
	}
	// Delete
	for i := 0; i < loop; i++ {
		v := hmap.Delete(Str("Hello" + string(i)))
		k := i * 10
		if k != v {
			t.Error("Expected deletion of "+string(k)+", got ", v)
		}
	}
	if hmap.Size() != 0 {
		t.Error("Expected 0, got ", hmap.Size())
	}
}

var dics = []KeyValue{
	KeyValue{Key: Str("Martim"), Value: 9},
	KeyValue{Key: Str("Paulo"), Value: 41},
	KeyValue{Key: Str("Monica"), Value: 33},
	KeyValue{Key: Str("Francisca"), Value: 15},
}

func TestHashMapIterator(t *testing.T) {
	dic := NewHashMap()
	dic.Put(dics[0].Key, dics[0].Value)
	dic.Put(dics[1].Key, dics[1].Value)
	dic.Put(dics[2].Key, dics[2].Value)
	dic.Put(dics[3].Key, dics[3].Value)

	fmt.Println("========> Iterating through ", dics)
	for it := dic.Iterator(); it.HasNext(); {
		fmt.Println(it.Next())
	}

	dic.Delete(dics[1].Key)

	fmt.Println("========> Deleting ", dics[1])
	for it := dic.Iterator(); it.HasNext(); {
		fmt.Println(it.Next())
	}
}

func TestLinkedHashMapIterator(t *testing.T) {
	dic := NewLinkedHashMap()
	dic.Put(dics[0].Key, dics[0].Value)
	dic.Put(dics[1].Key, dics[1].Value)
	dic.Put(dics[2].Key, dics[2].Value)
	dic.Put(dics[3].Key, dics[3].Value)

	for k, v := range dic.Elements() {
		if !v.Key.Equals(dics[k].Key) {
			t.Errorf("Value %s at position %v does not match with %s\n", v.Key, k, dics[k].Key)
		}
	}
	dic.Delete(dics[1].Key)

	set := []KeyValue{dics[0], dics[2], dics[3]}
	elems := dic.Elements()
	for k, v := range elems {
		if !v.Key.Equals(set[k].Key) {
			t.Errorf("Value %s at position %d does not match with %s after delete\n", v.Key, k, dics[k].Key)
		}
	}

	//	fmt.Println("========> Deleting ", dics[1])
	//	for it := dic.Iterator(); it.HasNext(); {
	//		fmt.Println(it.Next())
	//	}
}
