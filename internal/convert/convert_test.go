package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTitleCase(t *testing.T) {
	tests := map[string]string{
		"":                 "",
		"lowercase":        "Lowercase",
		"Class":            "Class",
		"MyClass":          "My Class",
		"MyC":              "My C",
		"HTML":             "HTML",
		"PDFLoader":        "PDF Loader",
		"AString":          "A String",
		"SimpleXMLParser":  "Simple XML Parser",
		"vimRPCPlugin":     "Vim RPC Plugin",
		"GL11Version":      "GL 11 Version",
		"99Bottles":        "99 Bottles",
		"May5":             "May 5",
		"BFG9000":          "BFG 9000",
		"BöseÜberraschung": "Böse Überraschung",
		"Job title":        "Job Title",
		"jobTitle":         "Job Title",
		"job_title":        "Job Title",
		"Two  Spaces":      "Two Spaces",
	}

	for input, expected := range tests {
		assert.Equal(t, expected, TitleCase(input), "input: %s", input)
	}
}

func TestSince(t *testing.T) {
	assert.Equal(t, "Oct 9, 2021", Since(time.Unix(0, 1633728000000000000)))
	assert.Equal(t, "just now", Since(time.Now()))
}

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
			got := Color(tt.v)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestParse(t *testing.T) {
	assert.Equal(t, 1, Int("1", 0))
	assert.Equal(t, 0, Int("a", 0))
	assert.Equal(t, 1.0, Float("1", 0))
	assert.Equal(t, 0.0, Float("a", 0))
}
