// Package steam is a thin client for the public Steam Workshop endpoints used
// by the mod browser.
package steam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ark-server-commander/config"
	"ark-server-commander/models"
)

// ARKAppID is ARK: Survival Evolved on Steam. Workshop queries are scoped to it
// so the browser cannot surface items from unrelated games.
const ARKAppID = 346110

// ErrSearchUnavailable is returned when Workshop search is requested without a
// Steam Web API key. Item lookup by ID works without one; only search does not.
var ErrSearchUnavailable = errors.New("workshop search requires STEAM_API_KEY")

const (
	detailsEndpoint = "https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/"
	queryEndpoint   = "https://api.steampowered.com/IPublishedFileService/QueryFiles/v1/"
	requestTimeout  = 15 * time.Second
	maxSearchLimit  = 50
)

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: requestTimeout}}
}

// GetItems looks up Workshop items by ID. This endpoint is unauthenticated, so
// it works whether or not a Steam API key is configured.
//
// Steam returns HTTP 200 with a per-item result code even for IDs that do not
// exist, so missing items are simply absent from the returned slice rather than
// being an error.
func (c *Client) GetItems(ctx context.Context, workshopIDs []string) ([]models.WorkshopItem, error) {
	if len(workshopIDs) == 0 {
		return nil, nil
	}

	form := url.Values{}
	form.Set("itemcount", strconv.Itoa(len(workshopIDs)))
	for i, id := range workshopIDs {
		form.Set(fmt.Sprintf("publishedfileids[%d]", i), id)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, detailsEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build workshop request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workshop request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workshop returned status %d", resp.StatusCode)
	}

	var payload struct {
		Response struct {
			PublishedFileDetails []struct {
				PublishedFileID string `json:"publishedfileid"`
				Result          int    `json:"result"`
				Title           string `json:"title"`
				Description     string `json:"description"`
				PreviewURL      string `json:"preview_url"`
				FileSize        any    `json:"file_size"`
				Subscriptions   int64  `json:"subscriptions"`
				TimeUpdated     int64  `json:"time_updated"`
				ConsumerAppID   int    `json:"consumer_app_id"`
			} `json:"publishedfiledetails"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode workshop response: %w", err)
	}

	items := make([]models.WorkshopItem, 0, len(payload.Response.PublishedFileDetails))
	for _, d := range payload.Response.PublishedFileDetails {
		// result 1 is success; anything else means the ID is missing, private
		// or deleted.
		if d.Result != 1 {
			continue
		}
		// Guard against a caller passing an ID from another game.
		if d.ConsumerAppID != 0 && d.ConsumerAppID != ARKAppID {
			continue
		}
		items = append(items, models.WorkshopItem{
			WorkshopID:    d.PublishedFileID,
			Title:         d.Title,
			Description:   d.Description,
			PreviewURL:    d.PreviewURL,
			FileSizeBytes: parseFileSize(d.FileSize),
			Subscriptions: d.Subscriptions,
			TimeUpdated:   d.TimeUpdated,
		})
	}

	return items, nil
}

// Search queries the Workshop for ARK items matching text.
//
// Unlike GetItems this endpoint requires a Steam Web API key. Rather than
// returning an opaque 403 from Steam, we fail fast with ErrSearchUnavailable so
// the UI can tell the operator exactly what to configure.
func (c *Client) Search(ctx context.Context, text string, page, limit int) ([]models.WorkshopItem, error) {
	if config.SteamAPIKey == "" {
		return nil, ErrSearchUnavailable
	}
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > maxSearchLimit {
		limit = 20
	}

	q := url.Values{}
	q.Set("key", config.SteamAPIKey)
	q.Set("appid", strconv.Itoa(ARKAppID))
	q.Set("search_text", text)
	q.Set("page", strconv.Itoa(page))
	q.Set("numperpage", strconv.Itoa(limit))
	q.Set("return_details", "true")
	q.Set("return_previews", "true")
	// query_type 12 is "by text match relevance" when search_text is set, and
	// degrades to trending when it is empty, which is a sensible browse default.
	q.Set("query_type", "12")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryEndpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("build workshop search: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workshop search: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrSearchUnavailable
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workshop search returned status %d", resp.StatusCode)
	}

	var payload struct {
		Response struct {
			PublishedFileDetails []struct {
				PublishedFileID string `json:"publishedfileid"`
				Result          int    `json:"result"`
				Title           string `json:"title"`
				Description     string `json:"file_description"`
				PreviewURL      string `json:"preview_url"`
				FileSize        any    `json:"file_size"`
				Subscriptions   int64  `json:"subscriptions"`
				TimeUpdated     int64  `json:"time_updated"`
			} `json:"publishedfiledetails"`
		} `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode workshop search: %w", err)
	}

	items := make([]models.WorkshopItem, 0, len(payload.Response.PublishedFileDetails))
	for _, d := range payload.Response.PublishedFileDetails {
		if d.Result != 0 && d.Result != 1 {
			continue
		}
		items = append(items, models.WorkshopItem{
			WorkshopID:    d.PublishedFileID,
			Title:         d.Title,
			Description:   d.Description,
			PreviewURL:    d.PreviewURL,
			FileSizeBytes: parseFileSize(d.FileSize),
			Subscriptions: d.Subscriptions,
			TimeUpdated:   d.TimeUpdated,
		})
	}

	return items, nil
}

// parseFileSize normalises Steam's inconsistent file_size field, which comes
// back as a JSON number on some endpoints and a quoted string on others.
func parseFileSize(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case string:
		n, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}
