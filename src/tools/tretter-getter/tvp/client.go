// Package tvp provides a client for fetching TV schedules from the TVP VOD API.
package tvp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"tretter-getter/models"
	"tretter-getter/utils"
)

const (
	// DefaultBaseURL is the base URL for the TVP VOD API.
	DefaultBaseURL = "https://vod.tvp.pl/api/"
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
)

// Client is an HTTP client for the TVP VOD API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// Option is a function that configures the Client.
type Option func(*Client)

// WithBaseURL sets the base URL for the client.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets the HTTP client to use.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithLogger sets the logger for the client.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient creates a new TVP API client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// FetchSchedule fetches the TV schedule for the given time range and station IDs.
func (c *Client) FetchSchedule(ctx context.Context, since, till time.Time, stationIDs []int) ([]models.Programme, error) {
	endpoint := strings.TrimSuffix(c.baseURL, "/") + "/products/lives/programmes"
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint URL: %w", err)
	}

	q := u.Query()
	for _, id := range stationIDs {
		q.Add("liveId[]", fmt.Sprintf("%d", id))
	}
	q.Set("since", since.Format("2006-01-02T15:04-0700"))
	q.Set("till", till.Format("2006-01-02T15:04-0700"))
	q.Set("lang", "PL")
	q.Set("platform", "BROWSER")
	u.RawQuery = q.Encode()

	c.logger.Debug("fetching schedule", "url", u.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set browser-like headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:147.0) Gecko/20100101 Firefox/147.0")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://vod.tvp.pl/na-zywo?full-epg=true")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse raw response to extract nested images
	var rawProgrammes []struct {
		models.Programme
		Images map[string][]struct {
			URL string `json:"url"`
		} `json:"images"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawProgrammes); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	var programmes []models.Programme
	for _, rp := range rawProgrammes {
		p := rp.Programme

		// Clean text fields from API (remove tabs, normalize whitespace)
		p.Title = utils.CleanText(p.Title)
		p.Description = utils.CleanText(p.Description)
		p.Slug = utils.CleanText(p.Slug)

		// Extract 16x9 image if available
		if imgs, ok := rp.Images["16x9"]; ok && len(imgs) > 0 {
			p.ImageURL = imgs[0].URL
			// Ensure protocol-relative URLs have https:
			if strings.HasPrefix(p.ImageURL, "//") {
				p.ImageURL = "https:" + p.ImageURL
			}
		}
		programmes = append(programmes, p)
	}

	c.logger.Debug("fetched programmes", "count", len(programmes))
	return programmes, nil
}

// FilterByStation filters programmes to only include those from the specified station.
func FilterByStation(programmes []models.Programme, stationID int) []models.Programme {
	var filtered []models.Programme
	for _, p := range programmes {
		if p.Live.ID == stationID {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// ExtractIDFromURL extracts the terminal station ID from a VOD URL.
// Example: https://vod.tvp.pl/live,1/tvp-na-dobre-i-na-zle,1998766 -> 1998766
func ExtractIDFromURL(u string) (int, error) {
	idx := strings.LastIndex(u, ",")
	if idx == -1 || idx == len(u)-1 {
		return 0, fmt.Errorf("invalid stream URL format: %s", u)
	}

	idStr := u[idx+1:]
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		return 0, fmt.Errorf("failed to parse station ID from %s: %w", idStr, err)
	}

	return id, nil
}
