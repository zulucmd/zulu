package testutil_test

import (
	"testing"

	"github.com/zulucmd/zulu/v2/internal/testutil"
)

func TestIsNil(t *testing.T) {
	t.Parallel()

	var (
		nilPointer *int
		nilMap     map[string]int
		nilChannel chan int
		nilSlice   []int
	)

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{
			name:     "nil interface",
			value:    nil,
			expected: true,
		},
		{
			name:     "nil pointer",
			value:    nilPointer,
			expected: true,
		},
		{
			name:     "nil map",
			value:    nilMap,
			expected: true,
		},
		{
			name:     "nil channel",
			value:    nilChannel,
			expected: true,
		},
		{
			name:     "nil slice",
			value:    nilSlice,
			expected: true,
		},
		{
			name:     "non-nil pointer",
			value:    new(int),
			expected: false,
		},
		{
			name:     "array",
			value:    [3]int{},
			expected: false,
		},
		{
			name:     "struct",
			value:    struct{}{},
			expected: false,
		},
		{
			name:     "int",
			value:    0,
			expected: false,
		},
		{
			name:     "string",
			value:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := testutil.IsNil(tt.value); got != tt.expected {
				t.Errorf("IsNil(%v) = %t, want %t", tt.value, got, tt.expected)
			}
		})
	}
}
