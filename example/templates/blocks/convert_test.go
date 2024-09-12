package blocks

import (
	"fmt"
	"testing"
	"time"

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

func TestUpdatedOf(t *testing.T) {
	assert.Equal(t, "Oct 9, 2021", updatedOf("1633728000000000000"))
	assert.Equal(t, "just now", updatedOf(fmt.Sprintf("%d", time.Now().UnixNano())))
}
