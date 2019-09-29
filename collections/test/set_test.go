package test

import (
	"fmt"
	"testing"

	. "github.com/quintans/toolkit/collection"
	"github.com/quintans/toolkit"
	. "github.com/quintans/toolkit/ext"
)

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

func RunSetSame(list Collection, t *testing.T) {
	list.Add(unsortedArray...)
	list.Add(Long(2))
	if list.Size() != 5 {
		t.Error("Expected size of 5, got", list.Size())
	}
	fmt.Println(list.Elements())
}

func RunSetSort(list Collection, t *testing.T) {
	list.Add(unsortedArray...)

	ordered := list.Sort(greater)
	if !compare(ordered, sortedArray) {
		t.Errorf("Expected %s, got %s\n", sortedArray, ordered)
	}
}

func RunSetAddAll(list Collection, t *testing.T) {
	list.Add(unsortedArray...)
	if !compare(list.Elements(), unsortedArray) {
		t.Errorf("Expected %s, got %s\n", unsortedArray, list.Elements())
	}
}

func RunSetContains(list Collection, t *testing.T) {
	list.Add(unsortedArray...)

	if list.Contains(Long(25)) {
		t.Error("Expected NOT to Contain 25")
	}
	if !list.Contains(Long(2)) {
		t.Error("Expected to Contain 2")
	}
}

func RunSetEnumerator(list Collection, t *testing.T) {
	list.Add(unsortedArray...)

	pos := 0
	for e := list.Enumerator(); e.HasNext(); {
		if !toolkit.Match(unsortedArray[pos], e.Next()) {
			t.Error("The enumeration did not return the same elements")
		}
		pos++
	}
}

func RunSetInEnumerator(list Collection, t *testing.T) {
	list.Add(unsortedArray...)

	for _, v := range unsortedArray {
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
