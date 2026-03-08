package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"tretter-getter/api"
	"tretter-getter/store"

	"github.com/labstack/echo/v4"
)

func TestGetEpisodes_Pagination(t *testing.T) {
	// 1. Setup in-memory DB
	dbStore, err := store.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer dbStore.Close()

	// 2. Seed data
	// We insert 20 episodes. Episode numbers 1 to 20.
	ctx := context.Background()
	for i := 1; i <= 20; i++ {
		_, err := dbStore.UpsertEpisode(ctx, store.UpsertEpisodeParams{
			EpisodeNumber:    int64(i),
			ProgrammeID:      100 + int64(i),
			Title:            fmt.Sprintf("Episode %d", i),
			Description:      "Desc",
			WebUrl:           "http://example.com",
			ImageUrl:         "",
			Since:            time.Now(),
			Till:             time.Now().Add(1 * time.Hour),
			RecordingStarted: time.Now(),
			DurationSeconds:  3600,
			Year:             2024,
		})
		if err != nil {
			t.Fatalf("failed to insert episode %d: %v", i, err)
		}
	}

	// 3. Setup Server
	srv := api.New(dbStore)
	e := echo.New()

	// 4. Test Case: Limit 5, Offset 5
	// The list is sorted by episode_number DESC by default (from query.sql).
	// Sorted: 20, 19, ..., 1.
	// Offset 0 (Limit 5) -> 20, 19, 18, 17, 16
	// Offset 5 (Limit 5) -> 15, 14, 13, 12, 11

	req := httptest.NewRequest(http.MethodGet, "/api/episodes?limit=5&offset=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := srv.GetEpisodes(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var episodes []store.Episode
	if err := json.Unmarshal(rec.Body.Bytes(), &episodes); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(episodes) != 5 {
		t.Errorf("expected 5 episodes, got %d", len(episodes))
	}

	// Verify expected IDs
	expectedEpisodeNumbers := []int64{15, 14, 13, 12, 11}
	var gotEpisodeNumbers []int64
	for _, ep := range episodes {
		gotEpisodeNumbers = append(gotEpisodeNumbers, ep.EpisodeNumber)
	}

	if !reflect.DeepEqual(gotEpisodeNumbers, expectedEpisodeNumbers) {
		t.Errorf("expected episode numbers %v, got %v", expectedEpisodeNumbers, gotEpisodeNumbers)
	}
}
