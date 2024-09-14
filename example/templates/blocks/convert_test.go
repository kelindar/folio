package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorOf(t *testing.T) {
	tests := []struct {
		name     string
		v        string
		expected string
	}{
		{
			name:     "Active",
			v:        "active",
			expected: "emerald",
		},
		{
			name:     "Inactive",
			v:        "inactive",
			expected: "red",
		},
		{
			name:     "Warning",
			v:        "warning",
			expected: "amber",
		},
		{
			name:     "Random",
			v:        "random",
			expected: "zinc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorOf(tt.v)
			assert.Equal(t, tt.expected, got)
		})
	}
}
