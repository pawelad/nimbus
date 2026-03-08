package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Notification represents a message to be sent via ntfy.
type Notification struct {
	Topic    string               `json:"topic"`
	Title    string               `json:"title"`
	Message  string               `json:"message"`
	Priority int                  `json:"priority,omitempty"`
	Tags     []string             `json:"tags,omitempty"`
	Actions  []NotificationAction `json:"actions,omitempty"`
}

// NotificationAction represents an action button in the notification.
type NotificationAction struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url"`
	Clear  bool   `json:"clear,omitempty"`
}

// startNotifier starts a background worker that listens for notifications and sends them.
func startNotifier(logger *slog.Logger, cfg Config, notifyChan <-chan Notification) {
	if cfg.NtfyURL == "" || cfg.NtfyTopic == "" {
		logger.Info("ntfy not configured, notification worker disabled")
		// Drain the channel to prevent blocking if something sends to it
		for range notifyChan {
		}
		return
	}

	logger.Info("starting notification worker", "url", cfg.NtfyURL, "topic", cfg.NtfyTopic)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for n := range notifyChan {
		if err := sendNotification(client, cfg, n); err != nil {
			logger.Error("failed to send notification", "title", n.Title, "error", err)
		} else {
			logger.Info("notification sent", "title", n.Title)
		}
	}
}

func sendNotification(client *http.Client, cfg Config, n Notification) error {
	n.Topic = cfg.NtfyTopic
	payload, err := json.Marshal(n)
	if err != nil {
		return fmt.Errorf("marshaling notification: %w", err)
	}

	// Publish to the base URL when using JSON API
	url := cfg.NtfyURL
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.NtfyToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.NtfyToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
