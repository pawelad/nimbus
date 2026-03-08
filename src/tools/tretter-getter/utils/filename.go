// Package utils provides common utility functions used across tretter-getter.
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var episodeNumberRegex = regexp.MustCompile(`(?i)odc\.?\s*(\d+)`)

// ExtractEpisodeNumber extracts the episode number from a programme title.
// Returns the episode number or an error if not found.
func ExtractEpisodeNumber(title string) (int, error) {
	matches := episodeNumberRegex.FindStringSubmatch(title)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no episode number found in title: %s", title)
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("parsing episode number: %w", err)
	}

	return num, nil
}

// FormatEpisodeDir returns the directory name for an episode (e.g., "E.0034").
func FormatEpisodeDir(episodeNumber int) string {
	return fmt.Sprintf("E.%04d", episodeNumber)
}

// SanitizeFilename removes or replaces characters that are problematic in filenames.
func SanitizeFilename(name string) string {
	// Remove all characters except letters, numbers, spaces, and dashes (hyphens, en-dashes, etc.).
	// This removes other punctuation (., !, ?) and symbols while keeping the title readable.
	re := regexp.MustCompile(`[^\p{L}\p{N}\s\p{Pd}]`)
	clean := re.ReplaceAllString(name, " ")

	// Normalize spaces (trim and collapse multiple spaces)
	return strings.Join(strings.Fields(clean), " ")
}
