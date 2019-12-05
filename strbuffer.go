package toolkit

import "bytes"

import "fmt"

type StrBuffer struct {
	buffer bytes.Buffer
	hash   int
}

// check if it implements Base interface
var _ Base = &StrBuffer{}

func NewStrBuffer(str ...interface{}) *StrBuffer {
	s := new(StrBuffer)
	s.Add(str...)
	return s
}

func (this *StrBuffer) Addf(template string, a ...interface{}) *StrBuffer {
	this.Add(fmt.Sprintf(template, a...))
	return this
}

func (this *StrBuffer) Add(a ...interface{}) *StrBuffer {
	for _, v := range a {
		this.buffer.WriteString(fmt.Sprintf("%v", v))
	}
	return this
}

func (this *StrBuffer) Size() int {
	return this.buffer.Len()
}

func (this *StrBuffer) IsEmpty() bool {
	return this.buffer.Len() == 0
}

func (this *StrBuffer) Clear() {
	this.buffer.Reset()
}

func (this *StrBuffer) String() string {
	return this.buffer.String()
}

func (this *StrBuffer) Equals(e interface{}) bool {
	switch t := e.(type) { //type switch
	case *StrBuffer:
		if this.Size() != t.Size() {
			return false
		}
		b1 := this.buffer.Bytes()
		b2 := t.buffer.Bytes()
		max := len(b1)
		for i := 0; i < max; i++ {
			if b1[i] != b2[i] {
				return false
			}
		}

		return true
		//return this.String() == t.String()
	}
	return false
}

func (this *StrBuffer) Clone() interface{} {
	x := new(StrBuffer)
	x.Add(x.String())
	return x
}

func (this *StrBuffer) HashCode() int {
	if this.hash == 0 {
		this.hash = HashString(HASH_SEED, this.String())
	}
	return this.hash
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

func (this *Joiner) Append(a ...interface{}) *Joiner {
	this.StrBuffer.Add(a...)
	return this
}

func (this *Joiner) AddAsOne(a ...interface{}) *Joiner {
	if this.StrBuffer.Size() > 0 {
		this.StrBuffer.Add(this.separator)
	}
	this.StrBuffer.Add(a...)
	return this
}

func (this *Joiner) Add(a ...interface{}) *Joiner {
	for _, v := range a {
		if this.StrBuffer.Size() > 0 {
			this.StrBuffer.Add(this.separator)
		}
		this.StrBuffer.Add(v)
	}
	return this
}
