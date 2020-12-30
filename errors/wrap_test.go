package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		format string
	}{
		{
			name: "plain error",
			err:  errors.New("plain"),
		},
		{
			name:   "composit plain error",
			err:    errors.New("something"),
			format: "This has a message: %w",
		},
		{
			name:   "native plain error",
			err:    New("something"),
			format: "This has a message: %w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.format == "" {
				err = Wrap(tt.err)
			} else {
				err = Errorf(tt.format, tt.err)
			}

			err = fmt.Errorf("double wrapping: %w", err)
			assert.Contains(t, err.Error(), tt.err.Error())
			assert.True(t, IsError(err))
			assert.Equal(t, 3, countLines(err.Error()))
		})
	}
}

func countLines(s string) int {
	count := 1
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
