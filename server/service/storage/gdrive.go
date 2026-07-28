package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// GDrive uploads backups to Google Drive.
//
// Drive has no long-lived token equivalent, so an OAuth2 refresh token plus
// client credentials is the only workable shape for an unattended service.
// Uploads use the resumable endpoint, which is the supported path for large
// files and avoids buffering the archive.
type GDrive struct {
	clientID     string
	clientSecret string
	refreshToken string
	folderID     string

	http *http.Client

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time
}

func NewGDrive(s Settings) *GDrive {
	return &GDrive{
		clientID:     s.GDriveClientID,
		clientSecret: s.GDriveClientSecret,
		refreshToken: s.GDriveRefreshToken,
		folderID:     s.GDriveFolderID,
		http:         &http.Client{Timeout: uploadTimeout},
	}
}

func (g *GDrive) Name() string { return "gdrive" }

func (g *GDrive) Configured() bool {
	return g.clientID != "" && g.clientSecret != "" && g.refreshToken != ""
}

func (g *GDrive) Destination() string {
	if g.folderID == "" {
		return "Google Drive (My Drive root)"
	}
	return "Google Drive folder " + g.folderID
}

// token exchanges the refresh token for an access token, cached until shortly
// before it expires.
func (g *GDrive) token() (string, error) {
	if !g.Configured() {
		return "", fmt.Errorf("google drive is not configured")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.cachedToken != "" && time.Now().Before(g.tokenExpiry) {
		return g.cachedToken, nil
	}

	form := url.Values{}
	form.Set("client_id", g.clientID)
	form.Set("client_secret", g.clientSecret)
	form.Set("refresh_token", g.refreshToken)
	form.Set("grant_type", "refresh_token")

	resp, err := g.http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return "", fmt.Errorf("google token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google token exchange returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}

	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode google token: %w", err)
	}

	g.cachedToken = out.AccessToken
	// Refresh a minute early so a long upload cannot straddle the expiry.
	g.tokenExpiry = time.Now().Add(time.Duration(out.ExpiresIn)*time.Second - time.Minute)
	return g.cachedToken, nil
}

func (g *GDrive) Upload(name string, body io.Reader, size int64, contentType string) error {
	token, err := g.token()
	if err != nil {
		return err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadURL, err := g.startResumable(token, name, contentType)
	if err != nil {
		return err
	}

	// A single PUT of the whole stream is valid for a resumable session and
	// keeps this simple; Content-Length is known, so Go streams it without
	// buffering.
	req, err := http.NewRequest(http.MethodPut, uploadURL, body)
	if err != nil {
		return fmt.Errorf("build drive upload: %w", err)
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", contentType)

	resp, err := g.http.Do(req)
	if err != nil {
		return fmt.Errorf("drive upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("drive upload returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}
	return nil
}

// startResumable creates the upload session and returns its URL.
func (g *GDrive) startResumable(token, name, contentType string) (string, error) {
	meta := map[string]any{"name": name}
	if g.folderID != "" {
		meta["parents"] = []string{g.folderID}
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("encode drive metadata: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost,
		"https://www.googleapis.com/upload/drive/v3/files?uploadType=resumable&supportsAllDrives=true",
		strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("build drive session: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Upload-Content-Type", contentType)

	resp, err := g.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("drive session: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("drive session returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("drive session did not return an upload URL")
	}
	return location, nil
}
