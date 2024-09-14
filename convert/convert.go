package convert

import (
	"fmt"
	"hash/crc32"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// ---------------------------------- Title Case ----------------------------------

func TitleCase(input string) string {
	words := splitCase(input)
	smallwords := " a an on the to "
	for index, word := range words {
		if strings.Contains(smallwords, " "+word+" ") {
			words[index] = word
		} else {
			words[index] = strings.Title(word)
		}
	}
	return strings.Join(words, " ")
}

// splitCase is a modified version https://github.com/fatih/camelcase
// original Copyright (c) 2015 Fatih Arslan
func splitCase(src string) (entries []string) {
	if !utf8.ValidString(src) {
		return []string{src}
	}

	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0

	// split into fields based on class of unicode character
	for _, r := range src {
		switch true {
		case r == '_':
			class = 0
			runes = append(runes, []rune{})
			continue
		case unicode.IsSpace(r):
			class = 0
			runes = append(runes, []rune{r})
			continue
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}

		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}

	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}

	// construct []string from results
	for _, s := range runes {
		if v := strings.Trim(string(s), " "); len(v) > 0 {
			entries = append(entries, v)
		}
	}
	return
}

// ---------------------------------- Date/Time ----------------------------------

// Since returns a human readable time format
func Since(t time.Time) string {
	d := time.Now().Sub(t)
	switch {
	case d.Minutes() < 1:
		return "just now"
	case d.Minutes() < 60:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d.Hours() < 2:
		return "an hour ago"
	case d.Hours() < 24:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	default:
		return t.Format("Jan 2, 2006")
	}
}

// ---------------------------------- Color ----------------------------------

var palette = []string{
	"slate", "gray", "zinc", "neutral", "stone", "orange",
	"yellow", "lime", "green", "teal", "cyan", "sky", "blue",
	"indigo", "violet", "purple", "fuchsia", "pink", "rose",
}

// Color returns a color for a hashed string (only tailwind colors)
func Color(v string) string {
	switch strings.ToLower(v) {
	case "active", "enabled", "healthy", "success", "up", "completed":
		return "emerald"
	case "inactive", "disabled", "unhealthy", "failure", "down", "error":
		return "red"
	case "warning", "warn", "pending":
		return "amber"
	}

	return palette[crc32.ChecksumIEEE([]byte(v))%uint32(len(palette))]
}
