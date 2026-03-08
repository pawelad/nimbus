package api

import (
	"net/http"

	"tretter-getter/store"
	"tretter-getter/web/templates"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// Server holds the dependencies for the web server handlers.
type Server struct {
	store *store.Store
}

// New creates a new Server instance.
func New(store *store.Store) *Server {
	return &Server{store: store}
}

// Render is a helper to render Templ components.
func (s *Server) Render(c echo.Context, status int, t templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	c.Response().WriteHeader(status)
	return t.Render(c.Request().Context(), c.Response().Writer)
}

// GetEpisodes handles GET /api/episodes
func (s *Server) GetEpisodes(c echo.Context) error {
	var params struct {
		Limit  int `query:"limit"`
		Offset int `query:"offset"`
	}
	// Set defaults
	params.Limit = 50
	params.Offset = 0

	if err := c.Bind(&params); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
	}

	// Enforce max limit to prevent abuse
	if params.Limit > 100 {
		params.Limit = 100
	}

	episodes, err := s.store.GetEpisodes(c.Request().Context(), store.GetEpisodesParams{
		Limit:  int64(params.Limit),
		Offset: int64(params.Offset),
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, episodes)
}

// GetDashboard handles GET /
func (s *Server) GetDashboard(c echo.Context, version string) error {
	ctx := c.Request().Context()
	count, _ := s.store.GetRecordedCount(ctx)

	// Fetch recent episodes
	library, _ := s.store.GetEpisodes(ctx, store.GetEpisodesParams{Limit: 50, Offset: 0})

	// Filter active ones for the top section
	// In a real app, we might have a separate query for this
	var active []store.Episode
	for _, ep := range library {
		if ep.Status == "recording" {
			active = append(active, ep)
		}
	}

	// We'll assume it's online if we reached this handler
	return s.Render(c, http.StatusOK, templates.Dashboard(true, version, count, active, library))
}

// GetStatusFragment handles GET /api/status (returning HTML)
func (s *Server) GetStatusFragment(c echo.Context, version string) error {
	count, _ := s.store.GetRecordedCount(c.Request().Context())
	return s.Render(c, http.StatusOK, templates.Status(true, version, count))
}

// GetActiveRecordingsFragment handles GET /api/active-recordings
func (s *Server) GetActiveRecordingsFragment(c echo.Context) error {
	// For now, just reuse GetEpisodes logic
	library, _ := s.store.GetEpisodes(c.Request().Context(), store.GetEpisodesParams{Limit: 20, Offset: 0})
	var active []store.Episode
	for _, ep := range library {
		if ep.Status == "recording" {
			active = append(active, ep)
		}
	}
	return s.Render(c, http.StatusOK, templates.ActiveRecordings(active))
}

// GetLibraryFragment handles GET /api/library
func (s *Server) GetLibraryFragment(c echo.Context) error {
	library, _ := s.store.GetEpisodes(c.Request().Context(), store.GetEpisodesParams{Limit: 50, Offset: 0})
	return s.Render(c, http.StatusOK, templates.EpisodeGrid(library))
}
