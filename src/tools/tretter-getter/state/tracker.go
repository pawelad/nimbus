// Package state provides file-based tracking of recorded episodes.
package state

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"tretter-getter/models"
	"tretter-getter/utils"
)

// Tracker manages file-based state for recorded episodes.
type Tracker struct {
	dataDir string
	puid    int
	pgid    int
	logger  *slog.Logger
}

// Option is a function that configures the Tracker.
type Option func(*Tracker)

// WithLogger sets the logger for the Tracker.
func WithLogger(logger *slog.Logger) Option {
	return func(t *Tracker) {
		t.logger = logger
	}
}

// WithOwnership sets the PUID and PGID for created files and directories.
func WithOwnership(puid, pgid int) Option {
	return func(t *Tracker) {
		t.puid = puid
		t.pgid = pgid
	}
}

// NewTracker creates a new state tracker.
func NewTracker(dataDir string, opts ...Option) *Tracker {
	t := &Tracker{
		dataDir: dataDir,
		logger:  slog.Default(),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// IsRecorded checks if an episode has already been recorded by looking for an exact file match in the episode directory.
// We delegate the actual os.Stat file existence check to GetRecordingPath so that both
// simple "is it done" checks and "get the download link" requests use the exact same truth
// and we never serve a download link to a file that was manually deleted from disk.
func (t *Tracker) IsRecorded(episodeNumber int, expectedTitle string) (bool, error) {
	_, err := t.GetRecordingPath(episodeNumber, expectedTitle)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetRecordingPath returns the relative path (to DataDir) of the recorded file.
// Returns os.ErrNotExist if not found.
func (t *Tracker) GetRecordingPath(episodeNumber int, expectedTitle string) (string, error) {
	episodeDir := filepath.Join(t.dataDir, utils.FormatEpisodeDir(episodeNumber))

	// Check if directory exists
	info, err := os.Stat(episodeDir)
	if os.IsNotExist(err) {
		return "", os.ErrNotExist
	}
	if err != nil {
		return "", fmt.Errorf("checking episode directory: %w", err)
	}
	if !info.IsDir() {
		return "", os.ErrNotExist
	}

	expectedFilename := utils.SanitizeFilename(expectedTitle) + ".mkv"
	expectedPath := filepath.Join(episodeDir, expectedFilename)

	// Check if exact file exists
	_, err = os.Stat(expectedPath)
	if err == nil {
		t.logger.Debug("found recorded episode",
			"episode", episodeNumber,
			"file", expectedFilename,
		)
		relativePath := filepath.Join(utils.FormatEpisodeDir(episodeNumber), expectedFilename)
		return relativePath, nil
	}

	if !os.IsNotExist(err) {
		return "", fmt.Errorf("checking episode file: %w", err)
	}

	return "", os.ErrNotExist
}

// CreateMetadataFile creates the metadata file for a recording.
// This is called before spawning the Docker worker.
func (t *Tracker) CreateMetadataFile(programme models.Programme, episodeNumber int, now time.Time) error {
	episodeDir := filepath.Join(t.dataDir, utils.FormatEpisodeDir(episodeNumber))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(episodeDir, 0755); err != nil {
		return fmt.Errorf("creating episode directory: %w", err)
	}

	// Ensure ownership of directory (if PUID/PGID set)
	if t.puid != 0 && t.pgid != 0 {
		if err := os.Chown(episodeDir, t.puid, t.pgid); err != nil {
			t.logger.Warn("failed to chown episode directory", "path", episodeDir, "error", err)
			// Don't fail hard, proceed
		}
	}

	// Calculate duration
	duration := programme.Till.Sub(programme.Since)

	metadata := models.RecordingMetadata{
		EpisodeNumber:    episodeNumber,
		ProgrammeID:      programme.ID,
		Title:            programme.Title,
		Slug:             programme.Slug,
		Description:      programme.Description,
		WebURL:           programme.WebURL,
		Since:            programme.Since,
		Till:             programme.Till,
		RecordingStarted: now,
		DurationSeconds:  int(duration.Seconds()),
		Year:             programme.Year,
	}

	metadataPath := filepath.Join(episodeDir, utils.FormatEpisodeDir(episodeNumber)+".meta.json")

	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("creating metadata file: %w", err)
	}
	defer file.Close()

	// Ensure ownership of metadata file
	if t.puid != 0 && t.pgid != 0 {
		if err := os.Chown(metadataPath, t.puid, t.pgid); err != nil {
			t.logger.Warn("failed to chown metadata file", "path", metadataPath, "error", err)
		}
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("encoding metadata: %w", err)
	}

	t.logger.Info("created metadata file",
		"episode", episodeNumber,
		"path", metadataPath,
	)

	return nil
}

// GetEpisodeDir returns the full path to an episode's directory.
func (t *Tracker) GetEpisodeDir(episodeNumber int) string {
	return filepath.Join(t.dataDir, utils.FormatEpisodeDir(episodeNumber))
}
