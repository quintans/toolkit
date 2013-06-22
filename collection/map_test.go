package collections

import (
	"fmt"
	"testing"
)

const value1 = "World"

func TestPutAndGet(t *testing.T) {
	hmap := NewHashMap()
	k := String("Hello")
	hmap.Put(k, value1)
	v, ok := hmap.Get(k)
	if !ok || value1 != v {
		t.Error("Expected "+value1+", got ", v)
	}
	v, ok = hmap.Get(String("lixo"))
	if ok {
		t.Error("Expected to find nothing, got ", v)
	}
}

func TestResize(t *testing.T) {
	hmap := NewHashMap()
	loop := 20
	// Insert
	for i := 0; i < loop; i++ {
		hmap.Put(String("Hello"+string(i)), i*10)
	}
	if hmap.Size() != loop {
		t.Error("Expected "+string(loop)+", got ", hmap.Size())
	}
	// Check
	for i := 0; i < loop; i++ {
		v, _ := hmap.Get(String("Hello" + string(i)))
		k := i * 10
		if k != v {
			t.Error("Expected "+string(k)+", got ", v)
		}
	}
	// Delete
	for i := 0; i < loop; i++ {
		v := hmap.Delete(String("Hello" + string(i)))
		k := i * 10
		if k != v {
			t.Error("Expected deletion of "+string(k)+", got ", v)
		}
	}
	if hmap.Size() != 0 {
		t.Error("Expected 0, got ", hmap.Size())
	}
}

var dics []KeyValue = []KeyValue{
	KeyValue{String("Martim"), 9},
	KeyValue{String("Paulo"), 41},
	KeyValue{String("Monica"), 33},
	KeyValue{String("Francisca"), 15},
}

func TestHashMapIterator(t *testing.T) {
	dic := NewHashMap()
	dic.Put(dics[0].Key, dics[0].Value)
	dic.Put(dics[1].Key, dics[1].Value)
	dic.Put(dics[2].Key, dics[2].Value)
	dic.Put(dics[3].Key, dics[3].Value)

	fmt.Println("========> Iterating throw ", dics)
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

	fmt.Println("========> Iterating throw ", dics)
	for it := dic.Iterator(); it.HasNext(); {
		fmt.Println(it.Next())
	}
	dic.Delete(dics[1].Key)

	fmt.Println("========> Deleting ", dics[1])
	for it := dic.Iterator(); it.HasNext(); {
		fmt.Println(it.Next())
	}
}
