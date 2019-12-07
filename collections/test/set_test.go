package test

import (
	"fmt"
	"testing"

	"github.com/quintans/toolkit"
	. "github.com/quintans/toolkit/collections"
	. "github.com/quintans/toolkit/ext"
)

var unsortedHashedArray = []toolkit.Hasher{
	Long(10),
	Long(2),
	Long(6),
	Long(71),
	Long(3),
}
var sortedHashedArray = []toolkit.Hasher{
	Long(2),
	Long(3),
	Long(6),
	Long(10),
	Long(71),
}

func TestSetHashSet(t *testing.T) {
	RunSetSame(NewHashSet(), t)
	RunSetContains(NewHashSet(), t)
	RunSetInEnumerator(NewHashSet(), t)
	RunSetSort(NewHashSet(), t)
}

func TestSetLinkedHashSet(t *testing.T) {
	RunSetSame(NewLinkedHashSet(), t)
	RunSetAddAll(NewLinkedHashSet(), t)
	RunSetContains(NewLinkedHashSet(), t)
	RunSetEnumerator(NewLinkedHashSet(), t)
	RunSetSort(NewLinkedHashSet(), t)
}

func RunSetSame(list ISet, t *testing.T) {
	list.Add(unsortedHashedArray...)
	list.Add(Long(2))
	if list.Size() != 5 {
		t.Error("Expected size of 5, got", list.Size())
	}
	fmt.Println(list.Elements())
}

func RunSetSort(list ISet, t *testing.T) {
	list.Add(unsortedHashedArray...)

	ordered := list.Sort(greater)
	if !compare(ordered, sortedArray) {
		t.Errorf("Expected %s, got %s\n", sortedArray, ordered)
	}
}

func RunSetAddAll(list ISet, t *testing.T) {
	list.Add(unsortedHashedArray...)
	if !compare(list.Elements(), unsortedArray) {
		t.Errorf("Expected %s, got %s\n", unsortedHashedArray, list.Elements())
	}
}

func RunSetContains(list ISet, t *testing.T) {
	list.Add(unsortedHashedArray...)

	if list.Contains(Long(25)) {
		t.Error("Expected NOT to Contain 25")
	}
	if !list.Contains(Long(2)) {
		t.Error("Expected to Contain 2")
	}
}

func RunSetEnumerator(list ISet, t *testing.T) {
	list.Add(unsortedHashedArray...)

	pos := 0
	for e := list.Enumerator(); e.HasNext(); {
		if !toolkit.Match(unsortedHashedArray[pos], e.Next()) {
			t.Error("The enumeration did not return the same elements")
		}
		pos++
	}
}

func RunSetInEnumerator(list ISet, t *testing.T) {
	list.Add(unsortedHashedArray...)

	for _, v := range unsortedHashedArray {
		var found = false
		for e := list.Enumerator(); e.HasNext(); {
			if toolkit.Match(v, e.Next()) {
				found = true
			}
		}
		if !found {
			t.Errorf("The enumeration did not found the element %s in %s\n", v, list.Elements())
			return
		}
	}

}
