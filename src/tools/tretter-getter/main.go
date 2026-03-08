// Package main is the entry point for tretter-getter.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	dockerclient "github.com/docker/docker/client"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"

	"tretter-getter/api"
	"tretter-getter/docker"
	"tretter-getter/models"
	"tretter-getter/scheduler"
	"tretter-getter/state"
	"tretter-getter/store"
	"tretter-getter/tvp"
)

// Config holds the application configuration.
type Config struct {
	Port             string `env:"PORT" envDefault:"1945"`
	DataDir          string `env:"DATA_DIR" envDefault:"/app/data"`
	DownloadsDir     string `env:"DOWNLOADS_DIR" envDefault:"/app/downloads"`
	HostDownloadsDir string `env:"HOST_DOWNLOADS_DIR" envDefault:"/data/downloads"`
	StreamURL        string `env:"STREAM_URL" envDefault:"https://vod.tvp.pl/live,1/tvp-na-dobre-i-na-zle,1998766"`
	BufferMinutes    int    `env:"BUFFER_MINUTES" envDefault:"1"`
	EpisodeFilter    string `env:"EPISODE_FILTER"`
	LogLevel         string `env:"LOG_LEVEL" envDefault:"info"`
	DryRun           bool   `env:"DRY_RUN" envDefault:"false"`
	PUID             int    `env:"PUID" envDefault:"1000"`
	PGID             int    `env:"PGID" envDefault:"1000"`
	// Download Docker Image (optional)
	DockerImage string `env:"DOCKER_IMAGE" envDefault:"jauderho/yt-dlp:2025.01.26"`
	// Ntfy Notifications (optional)
	NtfyURL     string `env:"NTFY_URL"`
	NtfyTopic   string `env:"NTFY_TOPIC"`
	NtfyToken   string `env:"NTFY_TOKEN"`
	ExternalURL string `env:"EXTERNAL_URL"`
}

// Version is the application version, usually set at build time.
var Version = "0.1.0dev"

func main() {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	// Setup structured logging
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("starting tretter-getter", "config", cfg, "version", Version)

	// Initialize Database
	dbPath := filepath.Join(cfg.DataDir, "tretter.db")
	dbStore, err := store.InitDB(dbPath)
	if err != nil {
		logger.Error("failed to init database", "path", dbPath, "error", err)
		os.Exit(1)
	}
	defer dbStore.Close()

	// Initialize Docker client
	dockerClient, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		logger.Error("failed to create docker client", "error", err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	// Parse episode filter
	filter, err := scheduler.ParseEpisodeFilter(cfg.EpisodeFilter)
	if err != nil {
		logger.Error("failed to parse episode filter", "filter", cfg.EpisodeFilter, "error", err)
		os.Exit(1)
	}

	// Initialize components
	tvpClient := tvp.NewClient(tvp.WithLogger(logger))
	sched := scheduler.NewScheduler(
		scheduler.WithBufferMinutes(cfg.BufferMinutes),
		scheduler.WithEpisodeFilter(filter),
		scheduler.WithLogger(logger),
	)
	tracker := state.NewTracker(cfg.DownloadsDir, state.WithLogger(logger), state.WithOwnership(cfg.PUID, cfg.PGID))

	// Use HostDownloadsDir for the worker, as it needs the path from the Docker host's perspective
	worker := docker.NewWorkerManager(dockerClient, cfg.DockerImage, cfg.HostDownloadsDir, cfg.StreamURL, cfg.PUID, cfg.PGID, docker.WithLogger(logger))

	// Initialize Notification Channel and Worker
	notifyChan := make(chan Notification, 100)
	go startNotifier(logger, cfg, notifyChan)

	// --- Job Logic ---
	// Use a channel to ensure only one job runs at a time (skip if still running)
	jobChan := make(chan struct{}, 1)
	job := func() {
		select {
		case jobChan <- struct{}{}:
			defer func() { <-jobChan }()
			if err := runCheck(cfg, logger, dbStore, tvpClient, sched, tracker, worker, notifyChan); err != nil {
				logger.Error("schedule check failed", "error", err)
			}
		default:
			logger.Warn("skipping schedule check, previous run still in progress")
		}
	}

	// --- Scheduler ---
	c := cron.New(cron.WithSeconds())
	// Run every minute at 00 seconds
	c.AddFunc("0 * * * * *", job)
	c.Start()
	logger.Info("scheduler started (every minute)")

	// --- Web Server (Echo) ---
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:  true,
		LogURI:     true,
		LogMethod:  true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency", v.Latency,
			)
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// API Routes
	srv := api.New(dbStore)
	e.GET("/", func(c echo.Context) error {
		return srv.GetDashboard(c, Version)
	})
	e.GET("/api/episodes", srv.GetEpisodes)

	e.GET("/api/status", func(c echo.Context) error {
		// Detect HTMX request to return fragment
		if c.Request().Header.Get("HX-Request") == "true" {
			return srv.GetStatusFragment(c, Version)
		}

		// Legacy JSON support
		count, err := dbStore.GetRecordedCount(c.Request().Context())
		if err != nil {
			logger.Error("failed to get recorded count", "error", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"status":         "ok",
			"version":        Version,
			"recorded_count": count,
		})
	})

	e.GET("/api/active-recordings", srv.GetActiveRecordingsFragment)
	e.GET("/api/library", srv.GetLibraryFragment)

	// Serve static files from the downloads directory
	e.Static("/downloads", cfg.DownloadsDir)

	// Run initial check immediately in a goroutine
	go func() {
		logger.Info("running initial check")
		job()
	}()

	// Graceful shutdown
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			logger.Error("shutting down server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close notification channel to stop worker
	close(notifyChan)

	c.Stop()
	if err := e.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
	}
	logger.Info("tretter-getter stopped")
}

func runCheck(
	cfg Config,
	logger *slog.Logger,
	db *store.Store,
	tvpClient *tvp.Client,
	sched *scheduler.Scheduler,
	tracker *state.Tracker,
	worker *docker.WorkerManager,
	notifyChan chan<- Notification,
) error {
	ctx := context.Background()
	now := time.Now()

	// Fetch schedule window
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())

	since := now.Add(-1 * time.Hour)
	if since.Before(startOfDay) {
		since = startOfDay
	}
	till := now.Add(1 * time.Hour)
	if till.After(endOfDay) {
		till = endOfDay
	}

	// Check ALL currently active recordings in the DB to see if they finished
	// This ensures we catch episodes that finished outside of their scheduled window
	activeEpisodes, err := db.GetActiveRecordings(ctx)
	if err != nil {
		logger.Error("failed to get active recordings", "error", err)
	} else {
		for _, ep := range activeEpisodes {
			checkActiveRecording(ctx, cfg, logger, db, tracker, worker, ep, notifyChan)
		}
	}

	targetStationID, err := tvp.ExtractIDFromURL(cfg.StreamURL)
	if err != nil {
		return fmt.Errorf("parsing station ID from URL: %w", err)
	}

	// Taken from a real API request when going to the website
	stationIDs := []int{399697, 399698, 399699, 399700, 399701, 399702, 1998766, 2543014, 2543049, 2543050}

	// Ensure target station is in the list
	if !slices.Contains(stationIDs, targetStationID) {
		stationIDs = append(stationIDs, targetStationID)
	}

	programmes, err := tvpClient.FetchSchedule(ctx, since, till, stationIDs)
	if err != nil {
		return fmt.Errorf("fetching schedule: %w", err)
	}

	programmes = tvp.FilterByStation(programmes, targetStationID)
	recordings := sched.FindRecordableEpisodes(programmes, now)

	if len(recordings) == 0 {
		return nil
	}

	for _, rec := range recordings {
		if err := processRecording(ctx, cfg, logger, db, tracker, worker, rec, now, notifyChan); err != nil {
			logger.Error("failed to process recording",
				"episode", rec.EpisodeNumber,
				"error", err,
			)
		}
	}

	return nil
}

func processRecording(
	ctx context.Context,
	cfg Config,
	logger *slog.Logger,
	db *store.Store,
	tracker *state.Tracker,
	worker *docker.WorkerManager,
	rec models.ScheduledRecording,
	now time.Time,
	notifyChan chan<- Notification,
) error {
	// 1. Get current state from Database
	ep, err := db.GetEpisode(ctx, int64(rec.EpisodeNumber))
	status := ""
	if err == nil {
		status = ep.Status
	}

	// 2. Check file existence (source of truth for "is it done?")
	recorded, err := tracker.IsRecorded(rec.EpisodeNumber, rec.Programme.Title)
	if err != nil {
		return fmt.Errorf("checking if recorded: %w", err)
	}
	if recorded {
		// If it's recorded but status is recording, checkActiveRecording might have missed it or we need to sync
		if status == "recording" && !cfg.DryRun {
			// It was caught here
			logger.Info("syncing status to recorded in processRecording", "episode", rec.EpisodeNumber)
			if syncErr := db.UpdateEpisodeStatus(ctx, store.UpdateEpisodeStatusParams{
				Status:        "recorded",
				EpisodeNumber: int64(rec.EpisodeNumber),
			}); syncErr == nil {
				// Send notification if we are the ones marking it done
				sendRecordingNotification(cfg, logger, tracker, int64(rec.EpisodeNumber), rec.Programme.Title, rec.Programme.Description, notifyChan)
			}
		}
		logger.Info("skipping, episode already recorded", "episode", rec.EpisodeNumber, "title", rec.Programme.Title)
		return nil
	}

	// 3. Check worker status and handle transitions
	running, err := worker.IsWorkerRunning(ctx, rec.EpisodeNumber)
	if err != nil {
		return fmt.Errorf("checking worker: %w", err)
	}

	if running {
		// If DB says it's not recording but worker is running, sync it
		if status != "recording" {
			logger.Warn("worker is running but DB says it's not recording", "episode", rec.EpisodeNumber, "status", status)
			if !cfg.DryRun {
				logger.Info("syncing status to recording", "episode", rec.EpisodeNumber)
				if err := db.UpdateEpisodeStatus(ctx, store.UpdateEpisodeStatusParams{
					Status:        "recording",
					EpisodeNumber: int64(rec.EpisodeNumber),
				}); err != nil {
					logger.Error("failed to sync status to recording", "episode", rec.EpisodeNumber, "error", err)
				}
			}
		}

		logger.Info("skipping, episode already being recorded", "episode", rec.EpisodeNumber, "title", rec.Programme.Title)
		return nil
	}

	// If we got here, the worker is NOT running.
	// We no longer handle finalizing here if it was "recording" because either it was
	// already finalized by the db.GetActiveRecordings() loop above,
	// or it crashed. If it crashed and file isn't there, it will fall through to spawn.
	if status == "recording" {
		logger.Warn("worker stopped but file not found or incomplete, will attempt to respawn", "episode", rec.EpisodeNumber)
	}

	if cfg.DryRun {
		logger.Info("DRY RUN: skipping state persistence and worker spawn", "episode", rec.EpisodeNumber)
		return nil
	}

	// 4. Upsert to DB and spawn worker
	_, err = db.UpsertEpisode(ctx, store.UpsertEpisodeParams{
		EpisodeNumber:    int64(rec.EpisodeNumber),
		ProgrammeID:      int64(rec.Programme.ID),
		Title:            rec.Programme.Title,
		Description:      rec.Programme.Description,
		WebUrl:           rec.Programme.WebURL,
		ImageUrl:         rec.Programme.ImageURL,
		Since:            rec.Programme.Since,
		Till:             rec.Programme.Till,
		RecordingStarted: now,
		DurationSeconds:  int64(rec.Programme.Till.Sub(rec.Programme.Since).Seconds()),
		Year:             int64(rec.Programme.Year),
		Status:           "recording",
	})
	if err != nil {
		logger.Error("failed to upsert episode to db", "episode", rec.EpisodeNumber, "error", err)
	}

	// Create redundant metadata file (keep legacy behavior for now as backup)
	if err := tracker.CreateMetadataFile(rec.Programme, rec.EpisodeNumber, now); err != nil {
		logger.Warn("failed to create metadata file", "episode", rec.EpisodeNumber, "error", err)
	}

	if err := worker.SpawnWorker(ctx, rec, now); err != nil {
		// If spawning fails, mark as failed in DB
		if updateErr := db.UpdateEpisodeStatus(ctx, store.UpdateEpisodeStatusParams{
			Status:        "failed",
			EpisodeNumber: int64(rec.EpisodeNumber),
		}); updateErr != nil {
			logger.Error("failed to mark episode as failed", "episode", rec.EpisodeNumber, "error", updateErr)
		}
		return fmt.Errorf("spawning worker: %w", err)
	}

	return nil
}

func checkActiveRecording(
	ctx context.Context,
	cfg Config,
	logger *slog.Logger,
	db *store.Store,
	tracker *state.Tracker,
	worker *docker.WorkerManager,
	ep store.Episode,
	notifyChan chan<- Notification,
) {
	running, err := worker.IsWorkerRunning(ctx, int(ep.EpisodeNumber))
	if err != nil {
		logger.Error("failed to check worker", "episode", ep.EpisodeNumber, "error", err)
		return
	}

	if !running {
		if isActuallyFinished, _ := tracker.IsRecorded(int(ep.EpisodeNumber), ep.Title); isActuallyFinished {
			logger.Info("worker finished successfully", "episode", ep.EpisodeNumber)
			if !cfg.DryRun {
				logger.Info("syncing status to recorded", "episode", ep.EpisodeNumber)
				err = db.UpdateEpisodeStatus(ctx, store.UpdateEpisodeStatusParams{
					Status:        "recorded",
					EpisodeNumber: ep.EpisodeNumber,
				})
				if err != nil {
					logger.Error("failed to sync status to recorded", "episode", ep.EpisodeNumber, "error", err)
				}

				if err == nil {
					sendRecordingNotification(cfg, logger, tracker, ep.EpisodeNumber, ep.Title, ep.Description, notifyChan)
				}
			}
		} else {
			// It's checked as "recording" in DB, but worker isn't running and file isn't complete.
			// If we are past the start window (5 minutes into the episode), we won't try to restart it.
			// Mark it as failed so it doesn't stay stuck in the 'recording' state until a rerun.
			if time.Now().After(ep.Since.Add(5 * time.Minute)) {
				logger.Warn("active recording worker stopped and too late to restart, marking as failed", "episode", ep.EpisodeNumber)
				if !cfg.DryRun {
					err = db.UpdateEpisodeStatus(ctx, store.UpdateEpisodeStatusParams{
						Status:        "failed",
						EpisodeNumber: ep.EpisodeNumber,
					})
					if err != nil {
						logger.Error("failed to sync status to failed", "episode", ep.EpisodeNumber, "error", err)
					}
				}
			} else {
				// The runCheck loop might respawn it if it's still in the start window.
				logger.Warn("active recording worker stopped but file incomplete", "episode", ep.EpisodeNumber)
			}
		}
	}
}

func sendRecordingNotification(
	cfg Config,
	logger *slog.Logger,
	tracker *state.Tracker,
	episodeNumber int64,
	title string,
	description string,
	notifyChan chan<- Notification,
) {
	var downloadLink string
	filename, err := tracker.GetRecordingPath(int(episodeNumber), title)
	if err == nil && cfg.ExternalURL != "" && filename != "" {
		var joinErr error
		downloadLink, joinErr = url.JoinPath(cfg.ExternalURL, "downloads", filename)
		if joinErr != nil {
			logger.Error("failed to construct download link", "error", joinErr)
		}
	}

	notifyChan <- Notification{
		Title:   fmt.Sprintf("Recorded: Na dobre i na złe (Ep %d)", episodeNumber),
		Message: fmt.Sprintf("%s\n\n%s", title, description),
		Tags:    []string{"tv"},
		Actions: []NotificationAction{
			{
				Action: "view",
				Label:  "Download / Watch",
				URL:    downloadLink,
			},
		},
	}
}
