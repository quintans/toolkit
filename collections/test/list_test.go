package test

import (
	"testing"

	. "github.com/quintans/toolkit"
	"github.com/quintans/toolkit/collections"
	. "github.com/quintans/toolkit/ext"
)

func compare(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if !Match(v, b[k]) {
			return false
		}
	}

	return true
}

func greater(a, b interface{}) bool {
	return a.(Long) < b.(Long)
}

var unsortedArray = []interface{}{
	Long(10),
	Long(2),
	Long(6),
	Long(71),
	Long(3),
}
var sortedArray = []interface{}{
	Long(2),
	Long(3),
	Long(6),
	Long(10),
	Long(71),
}

func TestSort(t *testing.T) {
	list := collections.NewArrayList()
	list.Add(unsortedArray...)

	ordered := list.Sort(greater)
	if !compare(ordered, sortedArray) {
		t.Error("Expected [2, 3, 6, 10, 71], got ", ordered)
	}
	if !compare(list.Elements(), unsortedArray) {
		t.Error("Expected [10, 2, 3, 6, 71], got ", ordered)
	}
}

func TestAddAll(t *testing.T) {
	list := collections.NewArrayList()
	list.Add(unsortedArray...)
	if !compare(list.Elements(), unsortedArray) {
		t.Error("Expected [12, 2, 30], got ", list.Elements())
	}
}

func TestFind(t *testing.T) {
	list := collections.NewArrayList()
	list.Add(unsortedArray...)

	i, _ := list.First(Long(2))
	if i != 1 {
		t.Error("Expected 1, got ", i)
	}
}

func TestContains(t *testing.T) {
	list := collections.NewArrayList()
	list.Add(unsortedArray...)

	if list.Contains(Long(25)) {
		t.Error("Expected NOT to Contain 25")
	}
	if !list.Contains(Long(2)) {
		t.Error("Expected to Contain 2")
	}
}

func TestEnumerator(t *testing.T) {
	list := collections.NewArrayList()
	list.Add(unsortedArray...)

	pos := 0
	for e := list.Enumerator(); e.HasNext(); {
		if !Match(unsortedArray[pos], e.Next()) {
			t.Error("The enumeration did not return the same elements")
		}
		pos++
	}
}
