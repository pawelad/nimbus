package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestStartNotifier(t *testing.T) {
	// 1. Setup Mock Server
	var receivedPayload Notification
	var receivedAuthHeader string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// 2. Setup Config & Logger
	cfg := Config{
		NtfyURL:   ts.URL, // Use the mock server URL
		NtfyTopic: "test-topic",
		NtfyToken: "secret-token",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 3. Start Notifier
	notifyChan := make(chan Notification, 1)
	go startNotifier(logger, cfg, notifyChan)

	// 4. Send Notification
	testNotification := Notification{
		Title:   "Test Title",
		Message: "Test Message",
		Tags:    []string{"tv"},
	}
	notifyChan <- testNotification

	// 5. Wait for processing (simple sleep for this unit test)
	time.Sleep(100 * time.Millisecond)
	close(notifyChan)

	// 6. Assertions
	if receivedPayload.Topic != cfg.NtfyTopic {
		t.Errorf("Expected topic %q, got %q", cfg.NtfyTopic, receivedPayload.Topic)
	}
	if receivedPayload.Title != testNotification.Title {
		t.Errorf("Expected title %q, got %q", testNotification.Title, receivedPayload.Title)
	}
	if receivedPayload.Message != testNotification.Message {
		t.Errorf("Expected message %q, got %q", testNotification.Message, receivedPayload.Message)
	}
	if len(receivedPayload.Tags) != 1 || receivedPayload.Tags[0] != "tv" {
		t.Errorf("Expected tag 'tv', got %v", receivedPayload.Tags)
	}
	expectedAuth := "Bearer secret-token"
	if receivedAuthHeader != expectedAuth {
		t.Errorf("Expected Auth header %q, got %q", expectedAuth, receivedAuthHeader)
	}
}

func TestStartNotifier_Disabled(t *testing.T) {
	// Ensure it doesn't panic or block if config is missing
	cfg := Config{} // Empty config
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	notifyChan := make(chan Notification, 1)

	// Should drain and return immediately/async without error
	go startNotifier(logger, cfg, notifyChan)
	notifyChan <- Notification{Title: "Should be ignored"}
	close(notifyChan)
}
