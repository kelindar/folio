package walk

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameOf(t *testing.T) {
	tests := map[string]struct {
		field    reflect.StructField
		expected string
	}{
		"No JSON tag": {
			field: reflect.StructField{
				Name: "FieldNoTag",
				Tag:  reflect.StructTag(""),
			},
			expected: "FieldNoTag",
		},
		"Empty JSON tag": {
			field: reflect.StructField{
				Name: "FieldEmptyTag",
				Tag:  reflect.StructTag(`json:""`),
			},
			expected: "FieldEmptyTag",
		},
		"JSON tag with name only": {
			field: reflect.StructField{
				Name: "FieldWithName",
				Tag:  reflect.StructTag(`json:"custom_name"`),
			},
			expected: "custom_name",
		},
		"JSON tag with name and single option": {
			field: reflect.StructField{
				Name: "FieldWithOption",
				Tag:  reflect.StructTag(`json:"custom_name,omitempty"`),
			},
			expected: "custom_name",
		},
		"JSON tag with name and multiple options": {
			field: reflect.StructField{
				Name: "FieldWithMultipleOptions",
				Tag:  reflect.StructTag(`json:"custom_name,omitempty,required"`),
			},
			expected: "custom_name",
		},
		"JSON tag with options only (no name)": {
			field: reflect.StructField{
				Name: "FieldOptionsOnly",
				Tag:  reflect.StructTag(`json:",omitempty"`),
			},
			expected: "FieldOptionsOnly",
		},
		"JSON tag starting with comma (no name)": {
			field: reflect.StructField{
				Name: "FieldStartsWithComma",
				Tag:  reflect.StructTag(`json:",omitempty,required"`),
			},
			expected: "FieldStartsWithComma",
		},
		"JSON tag with multiple commas and no name": {
			field: reflect.StructField{
				Name: "FieldMultipleCommasNoName",
				Tag:  reflect.StructTag(`json:",omitempty,required,another_option"`),
			},
			expected: "FieldMultipleCommasNoName",
		},
		"JSON tag with multiple commas and name": {
			field: reflect.StructField{
				Name: "FieldMultipleCommasWithName",
				Tag:  reflect.StructTag(`json:"custom_name,omitempty,required,another_option"`),
			},
			expected: "custom_name",
		},
		"JSON tag with name '-' and options": {
			field: reflect.StructField{
				Name: "FieldNameDash",
				Tag:  reflect.StructTag(`json:"-,"`),
			},
			expected: "FieldNameDash",
		},
		"JSON tag with uppercase letters": {
			field: reflect.StructField{
				Name: "FieldUpperCase",
				Tag:  reflect.StructTag(`json:"CUSTOM_NAME"`),
			},
			expected: "CUSTOM_NAME",
		},
		"JSON tag with spaces around name": {
			field: reflect.StructField{
				Name: "FieldWithSpaces",
				Tag:  reflect.StructTag(`json:" custom_name "`),
			},
			expected: " custom_name ",
		},
		"JSON tag with special characters in name": {
			field: reflect.StructField{
				Name: "FieldSpecialChars",
				Tag:  reflect.StructTag(`json:"custom-name_123"`),
			},
			expected: "custom-name_123",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := nameOf(&tc.field)
			assert.Equal(t, tc.expected, result, "nameOf(%v) should be %v", tc.field, tc.expected)
		})
	}
}
