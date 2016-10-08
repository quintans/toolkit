package test

import (
	"testing"

	"github.com/quintans/toolkit/picodi"
)

type Namer interface {
	Name() string
}

type Foo struct {
	name string
}

func (foo Foo) Name() string {
	return foo.name
}

type Bar struct {
	Foo   Foo   `wire:"foo"`
	Foo2   Foo   `wire:""`
	Other Namer `wire:"foo"`
	inner Foo   `wire:"foo"`
	Fun   Foo   `wire:"foofn"`
}

func (b *Bar) SetInner(v Foo) {
	b.inner = v
}

func TestStructWire(t *testing.T) {
	var pico = picodi.New()
	pico.Set("foo", Foo{"Foo"})
	pico.Set("Foo", Foo{"Foo"}) // unnamed wire
	pico.Set("foofn", func() Foo { return Foo{"FooFn"} })
	var bar = Bar{}
	if err := pico.Wire(&bar); err != nil {
		t.Fatal("Unexpected error when wiring bar: ", err)
	}

	if bar.Foo.Name() != "Foo" {
		t.Fatal("Expected \"Foo\" for Foo, got", bar.Foo.Name())
	}

	if bar.Other.Name() != "Foo" {
		t.Fatal("Expected \"Foo\" for Other, got", bar.Other.Name())
	}

	if bar.Foo2.Name() != "Foo" {
		t.Fatal("Expected \"Foo\" for Foo2, got", bar.Foo2.Name())
	}

	if bar.inner.Name() != "Foo" {
		t.Fatal("Expected \"Foo\" for inner, got", bar.inner.Name())
	}

	if bar.Fun.Name() != "FooFn" {
		t.Fatal("Expected \"FooFn\" for Fun, got", bar.Fun.Name())
	}
}

type Faulty struct {
	bar Bar `wire:"missing"`
}

func TestErrorWire(t *testing.T) {
	var pico = picodi.New()
	if err := pico.Wire(&Bar{}); err == nil {
		t.Fatal("Expected error for missing provider, nothing")
	}

	pico.Set("foo", Foo{"Foo"})
	if err := pico.Set("foo", Foo{"Foo"}); err == nil {
		t.Fatal("Expected error for double setting, got nothing")
	}

}
