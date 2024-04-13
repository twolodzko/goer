package reader

import (
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"1.", []string{"1."}},
		{" 2 + 2 / 4 .", []string{"2 + 2 / 4 ."}},
		{"1. 2+2. 3+3+3.", []string{"1.", "2+2.", "3+3+3."}},
		{"fun foo(X) -> X end. foo(X).", []string{"fun foo(X) -> X end.", "foo(X)."}},
	}

	for _, tt := range testCases {
		r := NewReader(strings.NewReader(tt.input))
		for _, expected := range tt.expected {
			result, err := r.Next()
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			} else if result != expected {
				t.Errorf("expected '%s', got '%s'", expected, result)
			}
		}
	}
}
