package scheduler

import (
	"testing"
	"time"

	"tretter-getter/models"
)

func TestFindRecordableEpisodes(t *testing.T) {
	// Fixed "now" for testing
	now := time.Date(2026, 1, 30, 23, 40, 0, 0, time.FixedZone("CET", 3600))

	programmes := []models.Programme{
		{
			// Currently airing - should be included
			ID:    2695283,
			Title: "odc. 34 – Wyznanie",
			Since: time.Date(2026, 1, 30, 23, 38, 41, 0, time.FixedZone("CET", 3600)),
			Till:  time.Date(2026, 1, 31, 0, 24, 26, 0, time.FixedZone("CET", 3600)),
			Live:  models.LiveInfo{ID: 1998766},
		},
		{
			// Already finished - should be excluded
			ID:    2695280,
			Title: "odc. 33 – Poprzedni",
			Since: time.Date(2026, 1, 30, 22, 0, 0, 0, time.FixedZone("CET", 3600)),
			Till:  time.Date(2026, 1, 30, 22, 45, 0, 0, time.FixedZone("CET", 3600)),
			Live:  models.LiveInfo{ID: 1998766},
		},
		{
			// Future episode - should be excluded
			ID:    2695290,
			Title: "odc. 35 – Następny",
			Since: time.Date(2026, 1, 31, 0, 30, 0, 0, time.FixedZone("CET", 3600)),
			Till:  time.Date(2026, 1, 31, 1, 15, 0, 0, time.FixedZone("CET", 3600)),
			Live:  models.LiveInfo{ID: 1998766},
		},
		{
			// No episode number - should be excluded
			ID:    2670375,
			Title: "Polowanie",
			Since: time.Date(2026, 1, 30, 23, 15, 0, 0, time.FixedZone("CET", 3600)),
			Till:  time.Date(2026, 1, 31, 1, 20, 0, 0, time.FixedZone("CET", 3600)),
			Live:  models.LiveInfo{ID: 399697},
		},
	}

	scheduler := NewScheduler(WithBufferMinutes(1))
	recordings := scheduler.FindRecordableEpisodes(programmes, now)

	if len(recordings) != 1 {
		t.Fatalf("expected 1 recording, got %d", len(recordings))
	}

	r := recordings[0]
	if r.EpisodeNumber != 34 {
		t.Errorf("expected episode 34, got %d", r.EpisodeNumber)
	}
	if r.ContainerName != "tretter-getter-download-0034" {
		t.Errorf("expected container name tretter-getter-download-0034, got %s", r.ContainerName)
	}
}

func TestFindRecordableEpisodes_InBufferWindow(t *testing.T) {
	// Test that episodes are caught during the buffer window before they start
	// Episode starts at 23:38:41, with 1 min buffer, recording starts at 23:37:41
	// If now is 23:37:45, episode should be recordable (within buffer, before episode starts)
	now := time.Date(2026, 1, 30, 23, 37, 45, 0, time.FixedZone("CET", 3600))

	programmes := []models.Programme{
		{
			ID:    2695283,
			Title: "odc. 34 – Wyznanie",
			Since: time.Date(2026, 1, 30, 23, 38, 41, 0, time.FixedZone("CET", 3600)),
			Till:  time.Date(2026, 1, 31, 0, 24, 26, 0, time.FixedZone("CET", 3600)),
			Live:  models.LiveInfo{ID: 1998766},
		},
	}

	scheduler := NewScheduler(WithBufferMinutes(1))
	recordings := scheduler.FindRecordableEpisodes(programmes, now)

	if len(recordings) != 1 {
		t.Fatalf("expected 1 recording in buffer window, got %d", len(recordings))
	}
}

func TestParseEpisodeFilter(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		episodes  []int
		wantMatch []bool
		wantErr   bool
	}{
		{
			name:      "Empty expression (allow all)",
			expr:      "",
			episodes:  []int{1, 10, 100},
			wantMatch: []bool{true, true, true},
		},
		{
			name:      "Single number",
			expr:      "15",
			episodes:  []int{14, 15, 16},
			wantMatch: []bool{false, true, false},
		},
		{
			name:      "Range",
			expr:      "10-20",
			episodes:  []int{9, 10, 15, 20, 21},
			wantMatch: []bool{false, true, true, true, false},
		},
		{
			name:      "Multiple items",
			expr:      "1, 5, 10-12",
			episodes:  []int{1, 2, 5, 9, 10, 11, 12, 13},
			wantMatch: []bool{true, false, true, false, true, true, true, false},
		},
		{
			name:      "Complex expression with whitespace",
			expr:      " 1 - 3 , 10 , 20 - 22 ",
			episodes:  []int{1, 2, 3, 4, 10, 11, 20, 21, 22, 23},
			wantMatch: []bool{true, true, true, false, true, false, true, true, true, false},
		},
		{
			name:    "Invalid number",
			expr:    "abc",
			wantErr: true,
		},
		{
			name:    "Invalid range (not numbers)",
			expr:    "10-abc",
			wantErr: true,
		},
		{
			name:    "Invalid range (reversed)",
			expr:    "20-10",
			wantErr: true,
		},
		{
			name:    "Invalid range format",
			expr:    "1-2-3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := ParseEpisodeFilter(tt.expr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for i, ep := range tt.episodes {
				if got := filter(ep); got != tt.wantMatch[i] {
					t.Errorf("filter(%d) = %v, want %v", ep, got, tt.wantMatch[i])
				}
			}
		})
	}
}

func TestFindRecordableEpisodes_WithFilter(t *testing.T) {
	now := time.Date(2026, 1, 30, 23, 40, 0, 0, time.FixedZone("CET", 3600))
	programmes := []models.Programme{
		{
			ID:    1,
			Title: "odc. 10",
			Since: now.Add(-2 * time.Minute),
			Till:  now.Add(10 * time.Minute),
		},
		{
			ID:    2,
			Title: "odc. 20",
			Since: now.Add(-2 * time.Minute),
			Till:  now.Add(10 * time.Minute),
		},
	}

	filter, _ := ParseEpisodeFilter("10")
	s := NewScheduler(WithEpisodeFilter(filter))
	recordings := s.FindRecordableEpisodes(programmes, now)

	if len(recordings) != 1 {
		t.Fatalf("expected 1 recording, got %d", len(recordings))
	}
	if recordings[0].EpisodeNumber != 10 {
		t.Errorf("expected episode 10, got %d", recordings[0].EpisodeNumber)
	}
}

func TestFindRecordableEpisodes_SkipsLateStart(t *testing.T) {
	now := time.Date(2026, 1, 30, 23, 40, 0, 0, time.FixedZone("CET", 3600))
	programmes := []models.Programme{
		{
			ID:    1,
			Title: "odc. 10",
			Since: now.Add(-6 * time.Minute), // More than 5 minutes past start
			Till:  now.Add(10 * time.Minute),
		},
		{
			ID:    2,
			Title: "odc. 20",
			Since: now.Add(-4 * time.Minute), // Less than 5 minutes past start
			Till:  now.Add(10 * time.Minute),
		},
	}

	s := NewScheduler()
	recordings := s.FindRecordableEpisodes(programmes, now)

	if len(recordings) != 1 {
		t.Fatalf("expected 1 recording, got %d", len(recordings))
	}
	if recordings[0].EpisodeNumber != 20 {
		t.Errorf("expected episode 20, got %d", recordings[0].EpisodeNumber)
	}
}
