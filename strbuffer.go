package toolkit

import (
	"fmt"
	"strings"
)

type StrBuffer struct {
	builder strings.Builder
	hash    int
}

// check if it implements Base interface
var _ Base = (*StrBuffer)(nil)

func NewStrBuffer(str ...interface{}) *StrBuffer {
	s := new(StrBuffer)
	s.Add(str...)
	return s
}

func (s *StrBuffer) Addf(template string, a ...interface{}) *StrBuffer {
	s.Add(fmt.Sprintf(template, a...))
	return s
}

func (s *StrBuffer) Add(a ...interface{}) *StrBuffer {
	for _, v := range a {
		switch t := v.(type) {
		case string:
			s.builder.WriteString(t)
		case fmt.Stringer:
			s.builder.WriteString(t.String())
		default:
			s.builder.WriteString(fmt.Sprintf("%v", v))
		}
	}
	return s
}

func (s *StrBuffer) Size() int {
	return s.builder.Len()
}

func (s *StrBuffer) IsEmpty() bool {
	return s.builder.Len() == 0
}

func (s *StrBuffer) Clear() {
	s.builder.Reset()
}

func (s *StrBuffer) String() string {
	return s.builder.String()
}

func (s *StrBuffer) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *StrBuffer:
		if s.Size() != t.Size() {
			return false
		}
		b1 := s.builder.String()
		b2 := t.builder.String()

		return b1 == b2
	}
	return false
}

func (s *StrBuffer) Clone() interface{} {
	x := &StrBuffer{}
	x.Add(x.String())
	return x
}

func (s *StrBuffer) HashCode() int {
	if s.hash == 0 {
		s.hash = HashString(HASH_SEED, s.String())
	}
	return s.hash
}

type Joiner struct {
	*StrBuffer
	separator string
}

func NewJoiner(separator string) *Joiner {
	this := new(Joiner)
	this.StrBuffer = new(StrBuffer)
	this.separator = separator
	return this
}

func (j *Joiner) Append(a ...interface{}) *Joiner {
	j.StrBuffer.Add(a...)
	return j
}

func (j *Joiner) AddAsOne(a ...interface{}) *Joiner {
	if j.StrBuffer.Size() > 0 {
		j.StrBuffer.Add(j.separator)
	}
	j.StrBuffer.Add(a...)
	return j
}

func (j *Joiner) Add(a ...interface{}) *Joiner {
	for _, v := range a {
		if j.StrBuffer.Size() > 0 {
			j.StrBuffer.Add(j.separator)
		}
		j.StrBuffer.Add(v)
	}
	return j
}
