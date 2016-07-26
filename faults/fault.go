package faults

import (
	"bytes"
	"fmt"
	"runtime"
)

var MaxDepth = 50

type Error struct {
	message string
	callers []uintptr
	cause   error
}

func create(a interface{}, args ...interface{}) *Error {
	var msg string
	var cause error

	switch e := a.(type) {
	case *Error:
		cause = e
	case error:
		cause = e
	case string:
		if len(args) > 0 {
			msg = fmt.Sprintf(e, args...)
		} else {
			msg = e
		}
	default:
		msg = fmt.Sprintf("%v", e)
	}

	callers := make([]uintptr, MaxDepth)
	length := runtime.Callers(3, callers)
	return &Error{
		message: msg,
		callers: callers[:length],
		cause:   cause,
	}
}

func New(a interface{}, args ...interface{}) *Error {
	return create(a, args...)
}

func Wrap(err error) error {
	if err != nil {
		switch e := err.(type) {
		case *Error:
			return e
		default:
			return create(err)
		}
	} else {
		return nil
	}
}

func (err *Error) StackTrace() string {
	var buf bytes.Buffer

	err.stackTrace(&buf)

	return string(buf.Bytes())
}

func (err *Error) stackTrace(buf *bytes.Buffer) {

	buf.WriteString(err.message)

	var pc uintptr
	for _, v := range err.callers {
		pc = v - 1
		fun := runtime.FuncForPC(pc)
		if fun == nil {
			buf.WriteString("\n    n/a")
		} else {
			var file, line = fun.FileLine(pc)
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
			buf.WriteString(fmt.Sprintf("\n    %s:%v (%s)", file, line, fun.Name()))
		}
	}

	if err.cause != nil {
		buf.WriteString("\n  caused by: ")
		switch t := err.cause.(type) {
		case *Error:
			t.stackTrace(buf)
		default:
			buf.WriteString(t.Error())
		}
	}
}

func (err *Error) Error() string {
	return err.message
}

func Dump(err error) string {
	switch e := err.(type) {
	case *Error:
		return e.StackTrace()
	default:
		return e.Error()
	}
}

func Has(err error, test error) bool {
	switch e := err.(type) {
	case *Error:
		return Has(e.cause, test)
	default:
		return err == test
	}
}
