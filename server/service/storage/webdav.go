package storage

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// WebDAV uploads backups to any WebDAV server — Nextcloud, ownCloud, or a plain
// Apache/nginx DAV mount.
//
// Included because it covers the self-hosted case that S3, Dropbox and Drive do
// not, and it is the simplest of the four: an authenticated PUT that streams.
type WebDAV struct {
	baseURL  string
	username string
	password string
	http     *http.Client
}

func NewWebDAV(s Settings) *WebDAV {
	return &WebDAV{
		baseURL:  strings.TrimSuffix(s.WebDAVURL, "/"),
		username: s.WebDAVUsername,
		password: s.WebDAVPassword,
		http:     &http.Client{Timeout: uploadTimeout},
	}
}

func (w *WebDAV) Name() string { return "webdav" }

func (w *WebDAV) Configured() bool {
	return w.baseURL != "" && w.username != "" && w.password != ""
}

func (w *WebDAV) Destination() string { return w.baseURL }

func (w *WebDAV) Upload(name string, body io.Reader, size int64, contentType string) error {
	if !w.Configured() {
		return fmt.Errorf("webdav is not configured")
	}

	// Each path segment is escaped individually so "/" keeps working as a
	// separator while spaces and other characters in the file name do not
	// produce a malformed request line.
	segments := strings.Split(strings.Trim(name, "/"), "/")
	for i, s := range segments {
		segments[i] = uriEncode(s)
	}
	target := w.baseURL + "/" + strings.Join(segments, "/")

	req, err := http.NewRequest(http.MethodPut, target, body)
	if err != nil {
		return fmt.Errorf("build webdav upload: %w", err)
	}
	req.ContentLength = size
	req.SetBasicAuth(w.username, w.password)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := w.http.Do(req)
	if err != nil {
		return fmt.Errorf("webdav upload: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// WebDAV answers 201 for a new resource and 204 for an overwrite; both are
	// success, as is any other 2xx.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webdav returned %d: %s", resp.StatusCode, errorSnippet(resp.Body))
	}
	return nil
}
