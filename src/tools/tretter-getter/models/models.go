// Package models defines data structures for TVP API responses and internal types.
package models

import "time"

// Programme represents a TV programme from the TVP API.
type Programme struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Lead        string    `json:"lead"`
	Since       time.Time `json:"since"`
	Till        time.Time `json:"till"`
	Live        LiveInfo  `json:"live"`
	WebURL      string    `json:"webUrl"`
	ImageURL    string    `json:"imageUrl,omitempty"` // Extracted from images array
	Year        int       `json:"year"`
}

// LiveInfo contains information about the live stream/station.
type LiveInfo struct {
	Type string `json:"type_"`
	ID   int    `json:"id"`
}

// RecordingMetadata represents the metadata saved alongside each recording.
type RecordingMetadata struct {
	EpisodeNumber    int       `json:"episodeNumber"`
	ProgrammeID      int       `json:"programmeId"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Description      string    `json:"description"`
	WebURL           string    `json:"webUrl"`
	Since            time.Time `json:"since"`
	Till             time.Time `json:"till"`
	RecordingStarted time.Time `json:"recordingStarted"`
	DurationSeconds  int       `json:"durationSeconds"`
	Year             int       `json:"year"`
}

// ScheduledRecording represents a programme scheduled for recording.
type ScheduledRecording struct {
	Programme     Programme
	EpisodeNumber int
	RecordStart   time.Time // since - buffer
	RecordEnd     time.Time // till + buffer
	ContainerName string    // tretter-getter-download-{EpisodeNumber} (4 digit with leading zeros)
}
