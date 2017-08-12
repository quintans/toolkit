package toolkit

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"time"
	"unicode"
)

func Match(o1, o2 interface{}) bool {
	switch t := o1.(type) { //type switch
	case Equaler:
		return t.Equals(o2)
	}

	switch t := o2.(type) { //type switch
	case Equaler:
		return t.Equals(o1)
	}

	return o1 == o2 // even if both are null
}

func SliceContains(list interface{}, elem interface{}) bool {
	v := reflect.ValueOf(list)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			instance := v.Index(i).Interface()
			switch t := instance.(type) {
			case Equaler:
				if t.Equals(elem) {
					return true
				}
			default:
				if instance == elem {
					return true
				}
			}
		}
	}
	return false
}

func Set(instance interface{}, value interface{}) {
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	r := reflect.ValueOf(value)
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}
	v.Set(r)
}

func Milliseconds() int64 {
	return time.Now().UnixNano() / int64(1e6)
}

// UncapFirst returns the input string with the first letter in lower case
func UncapFirst(str string) string {
	var s string
	if len(str) > 0 {
		s = string(unicode.ToLower(rune(str[0])))
	}
	if len(str) > 1 {
		s += str[1:]
	}
	return s
}

// CapFirst returns the input string with the first letter in Upper case
func CapFirst(str string) string {
	var s string
	if len(str) > 0 {
		s = string(unicode.ToUpper(rune(str[0])))
	}
	if len(str) > 1 {
		s += str[1:]
	}
	return s
}

func ToString(v interface{}) string {
	if t, isT := v.(string); isT {
		return t
	} else if t, isT := v.(fmt.Stringer); isT {
		return t.String()
	} else {
		var isNil bool
		var val reflect.Value
		if v == nil {
			isNil = true
		} else {
			val = reflect.ValueOf(v)
			if val.Kind() == reflect.Ptr && val.IsNil() {
				isNil = true
			}
		}

		if isNil {
			return "<nil>"
		} else {
			x := val.Interface()
			if val.Kind() == reflect.Ptr {
				x = val.Elem().Interface()
			}
			return fmt.Sprint(x)
		}
	}
}

func NewUUID() []byte {
	var b = make([]byte, 16)
	rand.Read(b)
	return b
}
