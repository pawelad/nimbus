package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"tretter-getter/models"
)

func TestIsRecorded(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "tretter-getter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracker := NewTracker(tmpDir)

	// Test 1: Episode doesn't exist
	recorded, err := tracker.IsRecorded(999, "Fake Episode")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if recorded {
		t.Error("expected episode 999 to not be recorded")
	}

	// Test 2: Episode directory exists but no mp4
	episodeDir := filepath.Join(tmpDir, "E.0034")
	if err := os.MkdirAll(episodeDir, 0755); err != nil {
		t.Fatalf("failed to create episode dir: %v", err)
	}
	// Create metadata file only
	metaFile := filepath.Join(episodeDir, "E.0034.meta.json")
	if err := os.WriteFile(metaFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create meta file: %v", err)
	}

	recorded, err = tracker.IsRecorded(34, "odc. 34 – Wyznanie")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if recorded {
		t.Error("expected episode 34 to not be recorded (no mp4)")
	}

	// Test 3: Episode directory exists with mkv
	mkvFile := filepath.Join(episodeDir, "odc 34 – Wyznanie.mkv")
	if err := os.WriteFile(mkvFile, []byte("fake video"), 0644); err != nil {
		t.Fatalf("failed to create mkv file: %v", err)
	}

	recorded, err = tracker.IsRecorded(34, "odc. 34 – Wyznanie")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !recorded {
		t.Error("expected episode 34 to be recorded (has mkv and recorded status)")
	}

	// Remove mkv
	os.Remove(mkvFile)

	recorded, err = tracker.IsRecorded(34, "odc. 34 – Wyznanie")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if recorded {
		t.Error("expected episode 34 to NOT be recorded (only has .part)")
	}
}

func TestGetRecordingPath(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "tretter-getter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracker := NewTracker(tmpDir)

	// Episode directory exists with mkv
	episodeDir := filepath.Join(tmpDir, "E.0034")
	if err := os.MkdirAll(episodeDir, 0755); err != nil {
		t.Fatalf("failed to create episode dir: %v", err)
	}
	mkvFile := filepath.Join(episodeDir, "odc 34 – Wyznanie.mkv")
	if err := os.WriteFile(mkvFile, []byte("fake video"), 0644); err != nil {
		t.Fatalf("failed to create mkv file: %v", err)
	}

	path, err := tracker.GetRecordingPath(34, "odc. 34 – Wyznanie")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	expectedPath := "E.0034/odc 34 – Wyznanie.mkv"
	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}

	// Test missing
	_, err = tracker.GetRecordingPath(999, "Fake Episode")
	if err == nil {
		t.Error("expected error for missing episode")
	}
}

func TestCreateMetadataFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tretter-getter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tracker := NewTracker(tmpDir)

	programme := models.Programme{
		ID:          2695283,
		Title:       "odc. 34 – Wyznanie",
		Slug:        "odc-34--wyznanie",
		Description: "Test description",
		Since:       time.Date(2026, 1, 30, 23, 38, 41, 0, time.UTC),
		Till:        time.Date(2026, 1, 31, 0, 24, 26, 0, time.UTC),
		Year:        2000,
	}

	now := time.Date(2026, 1, 30, 23, 37, 0, 0, time.UTC)
	err = tracker.CreateMetadataFile(programme, 34, now)
	if err != nil {
		t.Fatalf("failed to create metadata file: %v", err)
	}

	// Check that directory was created
	episodeDir := filepath.Join(tmpDir, "E.0034")
	if _, err := os.Stat(episodeDir); os.IsNotExist(err) {
		t.Error("episode directory was not created")
	}

	// Check that metadata file exists
	metaFile := filepath.Join(episodeDir, "E.0034.meta.json")
	if _, err := os.Stat(metaFile); os.IsNotExist(err) {
		t.Error("metadata file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(metaFile)
	if err != nil {
		t.Fatalf("failed to read metadata file: %v", err)
	}

	// Basic content checks
	if len(content) == 0 {
		t.Error("metadata file is empty")
	}
}

func TestGetEpisodeDir(t *testing.T) {
	tracker := NewTracker("/data/recordings")

	dir := tracker.GetEpisodeDir(34)
	expected := "/data/recordings/E.0034"
	if dir != expected {
		t.Errorf("expected %q, got %q", expected, dir)
	}
}
