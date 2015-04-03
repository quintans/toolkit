package ext

import (
	"reflect"
	"time"
)

func IsNil(value interface{}) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

func IsEmpty(test *string) bool {
	if test == nil {
		return true
	}

	return *test == ""
}

// Pointers

func Bool(val bool) *bool {
	return &val
}

func String(val string) *string {
	return &val
}

func Byte(val byte) *byte {
	return &val
}

func Int(val int) *int {
	return &val
}

func Int8(val int8) *int8 {
	return &val
}

func Int16(val int16) *int16 {
	return &val
}

func Int32(val int32) *int32 {
	return &val
}

func Int64(val int64) *int64 {
	return &val
}

func Float32(val float32) *float32 {
	return &val
}

func Float64(val float64) *float64 {
	return &val
}

// Defaults

func DefBool(test *bool, value bool) bool {
	if test == nil {
		return value
	} else {
		return *test
	}
}

func DefStr(test *string, value string) string {
	if test == nil {
		return value
	} else {
		return *test
	}
}

func DefInt64(test *int64, value int64) int64 {
	if test == nil {
		return value
	} else {
		return *test
	}
}

func DefFloat64(test *float64, value float64) float64 {
	if test == nil {
		return value
	} else {
		return *test
	}
}

func DefTime(test *time.Time, value time.Time) time.Time {
	if test == nil {
		return value
	} else {
		return *test
	}
}

func DefDate(test *Date, value Date) Date {
	if test == nil {
		return value
	} else {
		return *test
	}
}

// Clone

func CloneBool(test *bool) *bool {
	if test == nil {
		return nil
	} else {
		return Bool(*test)
	}
}

func CloneStr(test *string) *string {
	if test == nil {
		return nil
	} else {
		return String(*test)
	}
}

func CloneInt64(test *int64) *int64 {
	if test == nil {
		return nil
	} else {
		return Int64(*test)
	}
}

func CloneFloat64(test *float64) *float64 {
	if test == nil {
		return nil
	} else {
		return Float64(*test)
	}
}

func CloneTime(test *time.Time) *time.Time {
	if test == nil {
		return nil
	} else {
		v := *test
		return &v
	}
}

func CloneDate(test *Date) *Date {
	if test == nil {
		return nil
	} else {
		v := *test
		return &v
	}
}
