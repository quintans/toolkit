package toolkit

import "fmt"

type Equaler interface {
	Equals(other interface{}) bool
}

type Clonable interface {
	Clone() interface{}
}

type Hasher interface {
	Equaler
	HashCode() int
}

type Base interface {
	Hasher
	Clonable
	fmt.Stringer
}
