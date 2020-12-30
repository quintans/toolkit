package faults

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
)

const maxStackLength = 50

// Error is the type that implements the error interface.
// It contains the underlying err and its stacktrace.
type Error struct {
	Err        error
	stack      []uintptr
	stackTrace string
}

// Unwrap unpacks wrapped errors
func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	if e.stackTrace == "" {
		e.stackTrace = e.getStackTrace(e.Err)
	}
	return e.stackTrace
}

func (e Error) getStackTrace(err error) string {
	var trace bytes.Buffer
	trace.WriteString(err.Error())
	frames := runtime.CallersFrames(e.stack)
	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			trace.WriteString(fmt.Sprintf("\n\t%s:%d", frame.File, frame.Line))
		}
		if !more {
			break
		}
	}
	return trace.String()
}

// New returns a new error creates a new
func New(text string) error {
	return wrap(errors.New(text))
}

// Errorf creates a new error based on format and wraps it in a stack trace.
// The format string can include the %w verb.
func Errorf(format string, args ...interface{}) error {
	return wrap(fmt.Errorf(format, args...))
}

// Wrap annotates the given error with a stack trace
func Wrap(err error) error {
	return wrap(err)
}

func wrap(err error) error {
	var e *Error
	if err == nil || errors.As(err, &e) {
		return err
	}

	return &Error{Err: err, stack: getStack()}
}

func getStack() []uintptr {
	stackBuf := make([]uintptr, maxStackLength)
	length := runtime.Callers(4, stackBuf[:])
	return stackBuf[:length]
}

// IsError checks if the error is of type Error
func IsError(err error) bool {
	if err == nil {
		return false
	}

	var e *Error
	return errors.As(err, &e)
}
