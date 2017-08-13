package faults

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

var MaxDepth = 50

type Causer interface {
	Cause() error
}

type Fault struct {
	message string
	callers []uintptr
	cause   error
}

// Format implements the interface https://golang.org/pkg/fmt/#Formatter
// eg: https://github.com/go-stack/stack/blob/master/stack.go#L85
func (e *Fault) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		switch {
		case s.Flag('+'):
			s.Write([]byte(e.Stack()))
		default:
			s.Write([]byte(Error(e)))
		}
	default:
		s.Write([]byte(e.Error()))
	}
}

func create(msg string, cause error) *Fault {
	callers := make([]uintptr, MaxDepth)
	length := runtime.Callers(3, callers)
	return &Fault{
		message: msg,
		callers: callers[:length],
		cause:   cause,
	}
}

func format(template string, args ...interface{}) string {
	if len(args) > 0 {
		return fmt.Sprintf(template, args...)
	} else {
		return template
	}
}

func New(template string, args ...interface{}) error {
	return create(format(template, args...), nil)
}

func Wrapf(cause error, template string, args ...interface{}) error {
	if cause == nil {
		return nil
	}

	var msg = format(template, args...)
	var _, ok = cause.(Causer)
	if ok {
		return &Fault{
			message: msg,
			cause:   cause,
		}
	}

	return create(msg, cause)
}

func Wrap(cause error) error {
	if cause == nil {
		return nil
	}
	var _, ok = cause.(Causer)
	if ok {
		return cause
	}

	return create("", cause)
}

func WithStackf(cause error, template string, args ...interface{}) error {
	if cause == nil {
		return nil
	}

	var msg = format(template, args...)
	return create(msg, cause)
}

func WithStack(cause error) error {
	return WithStackf(cause, "")
}

func (err *Fault) Stack() string {
	var buf bytes.Buffer

	err.stack(&buf)

	return string(buf.Bytes())
}

func (err *Fault) stack(buf *bytes.Buffer) {

	buf.WriteString(err.Error())

	var pc uintptr
	for _, v := range err.callers {
		pc = v - 1
		fun := runtime.FuncForPC(pc)
		if fun == nil {
			buf.WriteString("\n    n/a")
		} else {
			var file, line = fun.FileLine(pc)

			const sep = "/"
			// gets only the filename.
			// eg: /home/paulo/go/src/package/file.go -> file.go
			var idx = strings.LastIndex(file, sep)
			if idx > 0 {
				file = file[idx+1:]
			}

			// package name.
			// eg: folder/package.funcname -> folder/package
			pck := fun.Name()
			idx = len(pck) - 1
			for i := idx; i > 0; i-- {
				if pck[i] == '.' {
					idx = i
				}
				if pck[i] == '/' {
					break
				}
			}
			pck = pck[:idx]
			file = pck + sep + file
			buf.WriteString(fmt.Sprintf("\n    %s:%v", file, line))
		}
	}

	var cause = err.Cause()
	if cause != nil {
		buf.WriteString("\ncaused by: ")
		switch t := cause.(type) {
		case *Fault:
			t.stack(buf)
		default:
			buf.WriteString(t.Error())
		}
	}
}

func (err *Fault) Error() string {
	return err.message
}

func Error(err error) string {
	var buf bytes.Buffer
	for err != nil {
		if err.Error() != "" {
			if buf.Len() > 0 {
				buf.WriteString(" > ")
			}
			buf.WriteString(err.Error())
		}
		cause, ok := err.(Causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}

	return buf.String()
}

func (err *Fault) Cause() error {
	return err.cause
}

func Cause(err error) error {
	for err != nil {
		cause, ok := err.(Causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

func Has(err error, test error) bool {
	switch e := err.(type) {
	case *Fault:
		return Has(e.cause, test)
	default:
		return err == test
	}
}
