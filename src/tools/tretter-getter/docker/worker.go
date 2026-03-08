// Package docker provides functionality for spawning and managing yt-dlp Docker containers.
package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"syscall"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerfilters "github.com/docker/docker/api/types/filters"
	dockerimage "github.com/docker/docker/api/types/image"
	dockerclient "github.com/docker/docker/client"

	"tretter-getter/models"
	"tretter-getter/utils"
)

const (
	// ContainerPrefix is the prefix for container names.
	ContainerPrefix = "tretter-getter-download-"
)

// WorkerManager manages yt-dlp Docker containers.
type WorkerManager struct {
	client    *dockerclient.Client
	image     string
	dataDir   string
	streamURL string
	puid      int
	pgid      int
	logger    *slog.Logger
}

// Option is a function that configures the WorkerManager.
type Option func(*WorkerManager)

// WithLogger sets the logger for the WorkerManager.
func WithLogger(logger *slog.Logger) Option {
	return func(w *WorkerManager) {
		w.logger = logger
	}
}

// NewWorkerManager creates a new WorkerManager.
func NewWorkerManager(dockerClient *dockerclient.Client, image, dataDir, streamURL string, puid, pgid int, opts ...Option) *WorkerManager {
	w := &WorkerManager{
		client:    dockerClient,
		image:     image,
		dataDir:   dataDir,
		streamURL: streamURL,
		puid:      puid,
		pgid:      pgid,
		logger:    slog.Default(),
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// IsWorkerRunning checks if a worker container is already running for the given episode number.
// If a container exists but is not running (e.g. exited), it is removed.
func (w *WorkerManager) IsWorkerRunning(ctx context.Context, episodeNumber int) (bool, error) {
	containerName := fmt.Sprintf("%s%04d", ContainerPrefix, episodeNumber)

	filterArgs := dockerfilters.NewArgs()
	filterArgs.Add("name", containerName)

	// List all containers, including stopped ones, to find zombies
	containers, err := w.client.ContainerList(ctx, dockercontainer.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return false, fmt.Errorf("listing containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			// Container names from Docker API include leading slash
			if strings.TrimPrefix(name, "/") == containerName {
				if c.State == "running" {
					w.logger.Debug("found running worker",
						"episode", episodeNumber,
						"containerName", containerName,
						"state", c.State,
					)
					return true, nil
				}

				// Found a container with the name, but it's not running (zombie/stopped)
				w.logger.Info("removing stale worker container",
					"containerName", containerName,
					"state", c.State,
				)
				if err := w.client.ContainerRemove(ctx, containerName, dockercontainer.RemoveOptions{Force: true}); err != nil {
					return false, fmt.Errorf("removing stale container: %w", err)
				}
				return false, nil
			}
		}
	}

	return false, nil
}

// SpawnWorker starts a new yt-dlp container to record an episode.
func (w *WorkerManager) SpawnWorker(ctx context.Context, recording models.ScheduledRecording, now time.Time) error {
	containerName := recording.ContainerName
	episodeDir := utils.FormatEpisodeDir(recording.EpisodeNumber)

	// Calculate duration: from now until RecordEnd.
	// This ensures the worker stops exactly when the scheduled window closes.
	duration := recording.RecordEnd.Sub(now)
	durationSeconds := int(duration.Seconds())
	if durationSeconds <= 0 {
		return fmt.Errorf("recording has already ended")
	}

	// Sanitize title for filename (remove/replace problematic characters)
	safeTitle := utils.SanitizeFilename(recording.Programme.Title)

	// Build yt-dlp command
	cmd := []string{
		"--output", fmt.Sprintf("/downloads/%s/%s.%%(ext)s", episodeDir, safeTitle),
		// Use specific format ID 4399 for 1280x720 50fps
		"-f", "4399/bestvideo[height<=720]+bestaudio/best[height<=720]",
		"--downloader", "ffmpeg",
		"--downloader-args", fmt.Sprintf("ffmpeg:-t %d", durationSeconds),
		// Force recode to MKV to trigger VideoConvertor and apply compression
		"--recode-video", "mkv",
		"--postprocessor-args", "VideoConvertor:-c:v hevc_qsv -global_quality 28 -look_ahead 1 -c:a aac -b:a 128k",
		"--retries", "10",
		"--fragment-retries", "10",
		w.streamURL,
	}

	w.logger.Info("spawning worker",
		"programmeId", recording.Programme.ID,
		"episode", recording.EpisodeNumber,
		"title", recording.Programme.Title,
		"durationSeconds", durationSeconds,
		"containerName", containerName,
	)

	// Ensure image exists locally
	if err := w.ensureImage(ctx, w.image); err != nil {
		return fmt.Errorf("ensuring image: %w", err)
	}

	// Create container config
	config := &dockercontainer.Config{
		Image: w.image,
		Cmd:   cmd,
		User:  fmt.Sprintf("%d:%d", w.puid, w.pgid),
	}

	// Host config with volume mount and hardware acceleration
	hostConfig := &dockercontainer.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/downloads", w.dataDir)},
		Resources: dockercontainer.Resources{
			Devices: []dockercontainer.DeviceMapping{
				{
					PathOnHost:        "/dev/dri",
					PathInContainer:   "/dev/dri",
					CgroupPermissions: "rwm",
				},
			},
		},
		AutoRemove: true,
		LogConfig: dockercontainer.LogConfig{
			Type: "journald",
			Config: map[string]string{
				"tag": fmt.Sprintf("tretter-worker-%d-%d", recording.Programme.ID, recording.EpisodeNumber),
			},
		},
	}

	// Attempt to determine render group ID to allow hardware acceleration access
	// This enables the non-root user to access /dev/dri/renderD128
	// TODO: Is `/dev/dri/renderD128` to specific?
	if gid, err := getPathGroupID("/dev/dri/renderD128"); err == nil {
		hostConfig.GroupAdd = []string{fmt.Sprintf("%d", gid)}
		w.logger.Debug("added render group to container", "gid", gid)
	} else {
		w.logger.Warn("failed to determine render group ID, hardware acceleration might fail", "path", "/dev/dri/renderD128", "error", err)
	}

	// Create the container
	resp, err := w.client.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return fmt.Errorf("creating container: %w", err)
	}

	w.logger.Debug("container created",
		"containerId", resp.ID,
		"containerName", containerName,
	)

	// Start the container
	if err := w.client.ContainerStart(ctx, resp.ID, dockercontainer.StartOptions{}); err != nil {
		return fmt.Errorf("starting container: %w", err)
	}

	w.logger.Info("worker started successfully",
		"containerId", resp.ID[:12],
		"containerName", containerName,
		"episode", recording.EpisodeNumber,
		"title", recording.Programme.Title,
	)

	return nil
}

// ensureImage checks if the image exists locally and pulls it if not.
func (w *WorkerManager) ensureImage(ctx context.Context, image string) error {
	_, err := w.client.ImageInspect(ctx, image)
	if err == nil {
		return nil
	}
	if !cerrdefs.IsNotFound(err) {
		return fmt.Errorf("inspecting image: %w", err)
	}

	w.logger.Info("pulling image", "image", image)
	reader, err := w.client.ImagePull(ctx, image, dockerimage.PullOptions{})
	if err != nil {
		return fmt.Errorf("pulling image: %w", err)
	}
	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("reading pull output: %w", err)
	}

	return nil
}

// getPathGroupID returns the GID of the given path.
func getPathGroupID(path string) (int, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get sys stats for %s", path)
	}
	return int(stat.Gid), nil
}
