package blocks

import (
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"time"
)

var palette = []string{
	"slate", "gray", "zinc", "neutral", "stone", "orange",
	"yellow", "lime", "green", "teal", "cyan", "sky", "blue",
	"indigo", "violet", "purple", "fuchsia", "pink", "rose",
}

// colorOf returns a color for a hashed string (only tailwind colors)
func colorOf(v string) string {
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

func updatedOf(updatedAt string) string {
	i, err := strconv.ParseInt(updatedAt, 10, 64)
	if err != nil {
		return ""
	}

	// Calculate the duration
	t := time.Unix(0, i)
	d := time.Now().Sub(t)

	// Return the duration in a human readable format
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
