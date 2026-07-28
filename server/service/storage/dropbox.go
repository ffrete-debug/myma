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

// Dropbox uploads backups to a Dropbox app folder.
//
// Dropbox refuses a single-shot upload above 150 MB, and ARK saves routinely
// exceed that, so anything large goes through an upload session in chunks.
type Dropbox struct {
	accessToken  string
	refreshToken string
	appKey       string
	appSecret    string
	path         string

	http *http.Client

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time
}

const (
	// dropboxSingleShotLimit is Dropbox's hard cap for /files/upload. We switch
	// to a session well below it to leave headroom.
	dropboxSingleShotLimit = 100 << 20 // 100 MiB
	// dropboxChunkSize balances request count against retry cost.
	dropboxChunkSize = 32 << 20 // 32 MiB
)

func NewDropbox(s Settings) *Dropbox {
	return &Dropbox{
		accessToken:  s.DropboxAccessToken,
		refreshToken: s.DropboxRefreshToken,
		appKey:       s.DropboxAppKey,
		appSecret:    s.DropboxAppSecret,
		path:         s.DropboxPath,
		http:         &http.Client{Timeout: uploadTimeout},
	}
}

func (d *Dropbox) Name() string { return "dropbox" }

func (d *Dropbox) Configured() bool {
	if d.accessToken != "" {
		return true
	}
	return d.refreshToken != "" && d.appKey != "" && d.appSecret != ""
}

func (d *Dropbox) Destination() string {
	return "Dropbox:" + "/" + strings.Trim(d.path, "/")
}

// token returns a usable access token, refreshing if necessary.
//
// Dropbox access tokens are short-lived now, so a deployment configured with a
// refresh token must exchange it before each upload run; caching avoids doing
// that on every chunk.
func (d *Dropbox) token() (string, error) {
	if d.refreshToken == "" {
		if d.accessToken == "" {
			return "", fmt.Errorf("dropbox is not configured")
		}
		return d.accessToken, nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cachedToken != "" && time.Now().Before(d.tokenExpiry) {
		return d.cachedToken, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", d.refreshToken)

	req, err := http.NewRequest(http.MethodPost,
		"https://api.dropbox.com/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build dropbox token request: %w", err)
	}
	req.SetBasicAuth(d.appKey, d.appSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("dropbox token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dropbox token exchange returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}

	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode dropbox token: %w", err)
	}

	d.cachedToken = out.AccessToken
	// Refresh a minute early so a long upload cannot straddle the expiry.
	d.tokenExpiry = time.Now().Add(time.Duration(out.ExpiresIn)*time.Second - time.Minute)
	return d.cachedToken, nil
}

func (d *Dropbox) Upload(name string, body io.Reader, size int64, _ string) error {
	if !d.Configured() {
		return fmt.Errorf("dropbox is not configured")
	}

	token, err := d.token()
	if err != nil {
		return err
	}

	remote := "/" + joinPath(d.path, name)

	if size <= dropboxSingleShotLimit {
		return d.uploadSingle(token, remote, body, size)
	}
	return d.uploadSession(token, remote, body)
}

func (d *Dropbox) uploadSingle(token, remote string, body io.Reader, size int64) error {
	arg, _ := json.Marshal(map[string]any{
		"path": remote, "mode": "overwrite", "mute": true,
	})

	req, err := http.NewRequest(http.MethodPost, "https://content.dropboxapi.com/2/files/upload", body)
	if err != nil {
		return fmt.Errorf("build dropbox upload: %w", err)
	}
	req.ContentLength = size
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(arg))

	resp, err := d.http.Do(req)
	if err != nil {
		return fmt.Errorf("dropbox upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dropbox upload returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}
	return nil
}

// uploadSession streams a large archive in chunks.
func (d *Dropbox) uploadSession(token, remote string, body io.Reader) error {
	sessionID, written, err := d.startSession(token, body)
	if err != nil {
		return err
	}

	buf := make([]byte, dropboxChunkSize)
	for {
		n, readErr := io.ReadFull(body, buf)
		if n > 0 {
			if err := d.appendChunk(token, sessionID, written, buf[:n]); err != nil {
				return err
			}
			written += int64(n)
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read backup for dropbox upload: %w", readErr)
		}
	}

	return d.finishSession(token, sessionID, written, remote)
}

func (d *Dropbox) startSession(token string, body io.Reader) (string, int64, error) {
	buf := make([]byte, dropboxChunkSize)
	n, err := io.ReadFull(body, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", 0, fmt.Errorf("read backup for dropbox upload: %w", err)
	}

	req, reqErr := http.NewRequest(http.MethodPost,
		"https://content.dropboxapi.com/2/files/upload_session/start", strings.NewReader(string(buf[:n])))
	if reqErr != nil {
		return "", 0, fmt.Errorf("build dropbox session: %w", reqErr)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", `{"close":false}`)

	resp, err := d.http.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("dropbox session start: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("dropbox session start returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}

	var out struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", 0, fmt.Errorf("decode dropbox session: %w", err)
	}
	return out.SessionID, int64(n), nil
}

func (d *Dropbox) appendChunk(token, sessionID string, offset int64, chunk []byte) error {
	arg, _ := json.Marshal(map[string]any{
		"cursor": map[string]any{"session_id": sessionID, "offset": offset},
		"close":  false,
	})

	req, err := http.NewRequest(http.MethodPost,
		"https://content.dropboxapi.com/2/files/upload_session/append_v2", strings.NewReader(string(chunk)))
	if err != nil {
		return fmt.Errorf("build dropbox append: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(arg))

	resp, err := d.http.Do(req)
	if err != nil {
		return fmt.Errorf("dropbox append: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dropbox append returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}
	return nil
}

func (d *Dropbox) finishSession(token, sessionID string, offset int64, remote string) error {
	arg, _ := json.Marshal(map[string]any{
		"cursor": map[string]any{"session_id": sessionID, "offset": offset},
		"commit": map[string]any{"path": remote, "mode": "overwrite", "mute": true},
	})

	req, err := http.NewRequest(http.MethodPost,
		"https://content.dropboxapi.com/2/files/upload_session/finish", nil)
	if err != nil {
		return fmt.Errorf("build dropbox finish: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Dropbox-API-Arg", string(arg))

	resp, err := d.http.Do(req)
	if err != nil {
		return fmt.Errorf("dropbox finish: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dropbox finish returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}
	return nil
}

// errorSnippet reads a bounded excerpt of an error body. Provider errors can be
// long HTML pages; a short excerpt is enough to tell credentials from quota.
func errorSnippet(r io.Reader) string {
	b, _ := io.ReadAll(io.LimitReader(r, 512))
	return strings.TrimSpace(string(b))
}
