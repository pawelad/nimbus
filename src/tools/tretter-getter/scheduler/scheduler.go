// Package scheduler provides logic for finding episodes to record.
package scheduler

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"tretter-getter/models"
	"tretter-getter/utils"
)

// Scheduler handles finding episodes that need to be recorded.
type Scheduler struct {
	bufferMinutes int
	filter        func(int) bool
	logger        *slog.Logger
}

// Option is a function that configures the Scheduler.
type Option func(*Scheduler)

// WithBufferMinutes sets the buffer minutes for recording.
func WithBufferMinutes(minutes int) Option {
	return func(s *Scheduler) {
		s.bufferMinutes = minutes
	}
}

// WithLogger sets the logger for the Scheduler.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Scheduler) {
		s.logger = logger
	}
}

// WithEpisodeFilter sets the episode filter for the Scheduler.
func WithEpisodeFilter(filter func(int) bool) Option {
	return func(s *Scheduler) {
		s.filter = filter
	}
}

// NewScheduler creates a new Scheduler.
func NewScheduler(opts ...Option) *Scheduler {
	s := &Scheduler{
		bufferMinutes: 0,
		filter:        func(int) bool { return true }, // Default: allow all
		logger:        slog.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// FindRecordableEpisodes finds episodes that should be recorded right now.
// An episode is recordable if:
// - We can extract an episode number from the title
// - The recording window (since - buffer, till + buffer) includes the current time
// - The episode hasn't finished yet (till + buffer > now)
func (s *Scheduler) FindRecordableEpisodes(programmes []models.Programme, now time.Time) []models.ScheduledRecording {
	buffer := time.Duration(s.bufferMinutes) * time.Minute
	var recordings []models.ScheduledRecording

	for _, p := range programmes {
		episodeNum, err := utils.ExtractEpisodeNumber(p.Title)
		if err != nil {
			s.logger.Debug("skipping programme without episode number",
				"title", p.Title,
				"error", err,
			)
			continue
		}

		// Skip if episode doesn't match the filter
		if s.filter != nil && !s.filter(episodeNum) {
			s.logger.Debug("skipping episode filtered out by expression",
				"title", p.Title,
				"episode", episodeNum,
			)
			continue
		}

		recordStart := p.Since.Add(-buffer)
		recordEnd := p.Till.Add(buffer)

		// Skip if episode has already finished (including buffer)
		if recordEnd.Before(now) {
			continue
		}

		// Skip if episode hasn't started yet (we'll catch it in a future run)
		if recordStart.After(now) {
			continue
		}

		// Only start recordings near the beginning of the scheduled time.
		// If we are more than 5 minutes into the actual airing, skip it.
		// It will be recorded later during a rerun.
		// This prevents creating download workers in the middle of an episode.
		startWindowEnd := p.Since.Add(5 * time.Minute)
		if now.After(startWindowEnd) {
			s.logger.Debug("skipping episode, too late to start full recording",
				"title", p.Title,
				"episode", episodeNum,
				"since", p.Since,
				"now", now,
			)
			continue
		}

		// This episode should be recording now
		recording := models.ScheduledRecording{
			Programme:     p,
			EpisodeNumber: episodeNum,
			RecordStart:   recordStart,
			RecordEnd:     recordEnd,
			ContainerName: fmt.Sprintf("tretter-getter-download-%04d", episodeNum),
		}

		s.logger.Debug("found recordable episode",
			"title", p.Title,
			"episode", episodeNum,
			"programmeId", p.ID,
			"recordStart", recordStart,
			"recordEnd", recordEnd,
		)

		recordings = append(recordings, recording)
	}

	return recordings
}

// ParseEpisodeFilter parses an episode filter expression (e.g., "1-10, 15, 20-25").
// Returns a function that returns true if an episode number matches the expression.
func ParseEpisodeFilter(expr string) (func(int) bool, error) {
	if strings.TrimSpace(expr) == "" {
		return func(int) bool { return true }, nil
	}

	parts := strings.Split(expr, ",")
	var filters []func(int) bool

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Range: X-Y
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range expression: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid range start: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid range end: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
			}

			filters = append(filters, func(n int) bool {
				return n >= start && n <= end
			})
		} else {
			// Single number: X
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid episode number: %s", part)
			}

			filters = append(filters, func(n int) bool {
				return n == num
			})
		}
	}

	if len(filters) == 0 {
		return func(int) bool { return true }, nil
	}

	return func(n int) bool {
		for _, f := range filters {
			if f(n) {
				return true
			}
		}
		return false
	}, nil
}
