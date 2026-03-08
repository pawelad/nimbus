package tvp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tretter-getter/models"
)

// TestCleanTextIntegration verifies that tab characters from the TVP API
// are cleaned when programmes are fetched.
func TestCleanTextIntegration(t *testing.T) {
	// Create a test programme with a tab character (simulating real TVP API data)
	testProgramme := struct {
		models.Programme
		Images map[string][]struct {
			URL string `json:"url"`
		} `json:"images"`
	}{
		Programme: models.Programme{
			ID:          2695283,
			Title:       "odc. 434 – Pochopna decyzja\t", // Tab character like in real data
			Slug:        "odc-434--pochopna-decyzja\t",
			Description: "Some description\twith\ttabs",
			Since:       time.Date(2026, 2, 10, 23, 0, 0, 0, time.UTC),
			Till:        time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC),
			Live:        models.LiveInfo{ID: 1998766},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]any{testProgramme})
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
	)

	programmes, err := client.FetchSchedule(
		context.Background(),
		time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 11, 23, 59, 0, 0, time.UTC),
		[]int{1998766},
	)
	if err != nil {
		t.Fatalf("FetchSchedule failed: %v", err)
	}

	if len(programmes) != 1 {
		t.Fatalf("expected 1 programme, got %d", len(programmes))
	}

	p := programmes[0]

	// Verify that tab characters have been cleaned
	expectedTitle := "odc. 434 – Pochopna decyzja"
	if p.Title != expectedTitle {
		t.Errorf("Title not cleaned: got %q, want %q", p.Title, expectedTitle)
	}

	// Verify that slug has normalized dashes (double dash -> single dash)
	expectedSlug := "odc-434-pochopna-decyzja"
	if p.Slug != expectedSlug {
		t.Errorf("Slug not cleaned: got %q, want %q", p.Slug, expectedSlug)
	}

	expectedDescription := "Some description with tabs"
	if p.Description != expectedDescription {
		t.Errorf("Description not cleaned: got %q, want %q", p.Description, expectedDescription)
	}
}
