package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderBy(t *testing.T) {
	tests := map[string]string{
		"":           "",
		"created_by": "created_by",
		"updated_by": "updated_by",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"createdBy":  "created_by",
		"updatedBy":  "updated_by",
		"createdAt":  "created_at",
		"updatedAt":  "updated_at",
		"-createdBy": "created_by DESC",
		"hello":      "json_extract(data, '$.hello')",
		"+hello":     "json_extract(data, '$.hello')",
		"-hello":     "json_extract(data, '$.hello') DESC",
	}

	for in, out := range tests {
		assert.Equal(t, out, queryOrder(in))
	}
}

func TestFilterJSON(t *testing.T) {
	tests := map[string]string{
		"":               "",
		"test":           `(json_extract(data, '$.test') IN ('x'))`,
		"something.name": `(json_extract(data, '$.something.name') IN ('x'))`,
	}

	for in, out := range tests {
		assert.Equal(t, out, queryFilterByJSON(in, []string{"x"}))
	}
}
