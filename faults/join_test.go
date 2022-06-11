package faults_test

import (
	"errors"
	"testing"

	"github.com/quintans/toolkit/faults"
	"github.com/stretchr/testify/require"
)

type CustomError struct {
	msg string
}

func (e CustomError) Error() string {
	return e.msg
}

var (
	Err1 = errors.New("E1")
	Err2 = &CustomError{"CUSTOM"}
	Err3 = errors.New("E3")
)

func TestIs(t *testing.T) {
	err := faults.Join(Err1, Err2)
	require.True(t, errors.Is(err, Err1))
	require.True(t, errors.Is(err, Err2))
	require.False(t, errors.Is(err, Err3))
	require.Equal(t, "E1: CUSTOM", err.Error())

	err = faults.Join(Err1, Err2, Err3)
	require.True(t, errors.Is(err, Err1))
	require.True(t, errors.Is(err, Err2))
	require.True(t, errors.Is(err, Err3))
	require.Equal(t, "E1: CUSTOM: E3", err.Error())
}

func TestAs(t *testing.T) {
	err := faults.Join(Err1, Err2)
	var e *faults.LinkedError
	require.True(t, errors.As(err, &e))
	require.Equal(t, "E1: CUSTOM", e.Error())
}

func TestAsCustom(t *testing.T) {
	err := faults.Join(Err1, Err2, Err3)
	var c *CustomError
	require.True(t, errors.As(err, &c))
	require.Equal(t, "CUSTOM", c.Error())
}
