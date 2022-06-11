package faults

import (
	"errors"
	"fmt"
)

func Join(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	first := &LinkedError{Err: errs[0]}
	last := first
	for i := 1; i < len(errs); i++ {
		last.Next = &LinkedError{Err: errs[i]}
		last = last.Next
	}
	return first
}

type LinkedError struct {
	Err  error
	Next *LinkedError
}

func (e *LinkedError) Error() string {
	if e.Next == nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %s", e.Err, e.Next)
}

func (e *LinkedError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

func (e *LinkedError) As(target interface{}) bool {
	return errors.As(e.Err, target)
}

func (e *LinkedError) Unwrap() error {
	if e.Next == nil {
		return nil
	}
	return e.Next
}
