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

func BoolPtr(val bool) *bool {
	return &val
}

func StrPtr(val string) *string {
	return &val
}

func Int64Ptr(val int64) *int64 {
	return &val
}

func Float64Ptr(val float64) *float64 {
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
		return BoolPtr(*test)
	}
}

func CloneStr(test *string) *string {
	if test == nil {
		return nil
	} else {
		return StrPtr(*test)
	}
}

func CloneInt64(test *int64) *int64 {
	if test == nil {
		return nil
	} else {
		return Int64Ptr(*test)
	}
}

func CloneFloat64(test *float64) *float64 {
	if test == nil {
		return nil
	} else {
		return Float64Ptr(*test)
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
