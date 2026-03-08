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

func TestFetchSchedule(t *testing.T) {
	testProgrammes := []models.Programme{
		{
			ID:    2695283,
			Title: "odc. 34 – Wyznanie",
			Slug:  "odc-34--wyznanie",
			Since: time.Date(2026, 1, 30, 23, 38, 41, 0, time.FixedZone("CET", 3600)),
			Till:  time.Date(2026, 1, 31, 0, 24, 26, 0, time.FixedZone("CET", 3600)),
			Live:  models.LiveInfo{ID: 1998766},
		},
		{
			ID:    2670375,
			Title: "Polowanie",
			Slug:  "polowanie",
			Since: time.Date(2026, 1, 30, 23, 15, 0, 0, time.FixedZone("CET", 3600)),
			Till:  time.Date(2026, 1, 31, 1, 20, 0, 0, time.FixedZone("CET", 3600)),
			Live:  models.LiveInfo{ID: 399697},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		// Check query parameters
		q := r.URL.Query()
		if q.Get("lang") != "PL" {
			t.Errorf("expected lang=PL, got %s", q.Get("lang"))
		}
		if q.Get("platform") != "BROWSER" {
			t.Errorf("expected platform=BROWSER, got %s", q.Get("platform"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testProgrammes)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
	)

	programmes, err := client.FetchSchedule(
		context.Background(),
		time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 23, 59, 0, 0, time.UTC),
		[]int{1998766, 399697},
	)
	if err != nil {
		t.Fatalf("FetchSchedule failed: %v", err)
	}

	if len(programmes) != 2 {
		t.Errorf("expected 2 programmes, got %d", len(programmes))
	}
}

func TestFilterByStation(t *testing.T) {
	programmes := []models.Programme{
		{ID: 1, Title: "Programme 1", Live: models.LiveInfo{ID: 1998766}},
		{ID: 2, Title: "Programme 2", Live: models.LiveInfo{ID: 399697}},
		{ID: 3, Title: "Programme 3", Live: models.LiveInfo{ID: 1998766}},
	}

	filtered := FilterByStation(programmes, 1998766)

	if len(filtered) != 2 {
		t.Errorf("expected 2 programmes, got %d", len(filtered))
	}

	for _, p := range filtered {
		if p.Live.ID != 1998766 {
			t.Errorf("expected station ID 1998766, got %d", p.Live.ID)
		}
	}
}

func TestFetchSchedule_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
	)

	_, err := client.FetchSchedule(
		context.Background(),
		time.Now(),
		time.Now().Add(24*time.Hour),
		[]int{1998766},
	)

	if err == nil {
		t.Error("expected error for HTTP 500, got nil")
	}
}
