package toolkit

import (
	"testing"
)

func TestAdd(t *testing.T) {
	sb := new(StrBuffer)
	sb.Add("Hello").Add(" ").Add("World")
	if "Hello World" != sb.String() {
		t.Error("Expected \"Hello World\", got ", sb.String())
	}

	if sb.Size() != 11 {
		t.Error("Expected 11, got ", sb.Size())
	}

	x := new(StrBuffer)
	x.Add(" teste ")
	sb.Add(x).Add(1).Add(true)
	if "Hello World teste 1true" != sb.String() {
		t.Error("\"Hello World teste 1true\", got ", sb.String())
	}
}

func TestEquals(t *testing.T) {
	sb1 := new(StrBuffer)
	sb1.Add("Hello")
	sb2 := new(StrBuffer)
	sb2.Add("Hello")
	if !sb1.Equals(sb2) {
		t.Error("The StrBuffers are not equal")
	}
}

func TestClone(t *testing.T) {
	sb1 := new(StrBuffer)
	sb1.Add("Hello")
	sb2 := sb1.Clone().(*StrBuffer)
	sb2.Add(" World")
	if sb1.Equals(sb2) {
		t.Error("The StrBuffers are equal")
	}
}
