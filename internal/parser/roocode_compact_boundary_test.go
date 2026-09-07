package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRooCodeIsCompactBoundary(t *testing.T) {
	tests := []struct {
		say  string
		want bool
	}{
		{"condense_context", true},
		{"sliding_window_truncation", true},
		{"condense_context_error", false},
		{"text", false},
		{"error", false},
		{"", false},
	}

	for _, tt := range tests {
		got := rooCodeIsCompactBoundary(tt.say)
		assert.Equal(t, tt.want, got,
			"rooCodeIsCompactBoundary(%q) = %v, want %v",
			tt.say, got, tt.want)
	}
}
