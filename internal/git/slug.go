package git

import (
	"regexp"
	"strings"
)

// Slugify converts a string to a URL-safe slug
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace any non-alphanumeric character with a dash
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")

	// Remove leading/trailing dashes
	s = strings.Trim(s, "-")

	// Truncate to 50 characters
	if len(s) > 50 {
		s = s[:50]
		// Don't end with a dash
		s = strings.TrimRight(s, "-")
	}

	return s
}
