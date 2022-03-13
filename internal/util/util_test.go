package util

import (
	"errors"
	"testing"
)

func TestCheckErr(t *testing.T) {
	tests := []struct {
		name  string
		msg   interface{}
		panic bool
	}{
		{
			name:  "no error",
			msg:   nil,
			panic: false,
		},
		{
			name:  "no panic empty string",
			msg:   "",
			panic: false,
		},
		{
			name:  "panic string",
			msg:   "test",
			panic: true,
		},
		{
			name:  "panic error",
			msg:   errors.New("test error"),
			panic: true,
		},
		{
			name:  "panic empty error",
			msg:   errors.New(""),
			panic: true,
		},
	}

	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if (r == nil) == tt.panic {
					t.Errorf("Didn't panic to be %t", tt.panic)
				}
			}()
			CheckErr(tt.msg)
		})
	}
}
