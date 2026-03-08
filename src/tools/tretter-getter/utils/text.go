package utils

import (
	"regexp"
	"strings"
)

// CleanText normalizes whitespace in text.
// This handles tab characters, newlines, and other problematic whitespace.
func CleanText(text string) string {
	// Replace tab characters and other problematic whitespace with spaces
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")

	// Normalize multiple consecutive spaces to single space
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// Normalize multiple consecutive dashes to single dash (for better slugs)
	dashRegex := regexp.MustCompile(`-+`)
	text = dashRegex.ReplaceAllString(text, "-")

	// Trim leading and trailing whitespace
	return strings.TrimSpace(text)
}
