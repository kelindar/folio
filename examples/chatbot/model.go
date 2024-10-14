package main

import (
	"github.com/kelindar/folio"
)

// ---------------------------------- Locale ----------------------------------

// LocalizedText represents a text that is localized in multiple languages.
type LocalizedText struct {
	English string `json:"english" form:"rw" validate:"required"`
	French  string `json:"french" form:"rw" validate:"required"`
}

// ---------------------------------- Intent ----------------------------------

// Intent represents an intent (question) that the user has asked.
type Intent struct {
	folio.Meta `kind:"intent" json:",inline"`
	Examples   []LocalizedText `json:"examples" form:"rw" validate:"required,min=1,dive"`
}

func (v *Intent) Title() string {
	for _, example := range v.Examples {
		return example.English
	}
	return ""
}

func (v *Intent) Subtitle() string {
	return ""
}

func NewIntent(text string) *Intent {
	return &Intent{
		Examples: []LocalizedText{
			{English: text, French: text},
		},
	}
}
