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
