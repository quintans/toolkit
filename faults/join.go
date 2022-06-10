package faults

import (
	"errors"
	"fmt"
)

func Join(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	err := errs[0]
	for i := 1; i < len(errs); i++ {
		err = &CompositeError{err, errs[i]}
	}
	return err
}

type CompositeError struct {
	Err1 error
	Err2 error
}

func (e *CompositeError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err1, e.Err2)
}

func (e *CompositeError) Is(target error) bool {
	return errors.Is(e.Err1, target)
}

func (e *CompositeError) As(target interface{}) bool {
	return errors.As(e.Err1, target)
}

func (e *CompositeError) Unwrap() error {
	return e.Err2
}
