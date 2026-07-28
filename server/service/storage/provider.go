package storage

import (
	"fmt"
	"io"
	"strings"
)

// Provider is a destination that a backup archive can be uploaded to.
//
// Uploads stream: ARK saves routinely run to gigabytes, so no implementation
// may buffer the whole archive in memory. Each provider is given the exact size
// up front, because every one of these APIs needs it (S3 for the signature,
// Dropbox and Drive to choose between single-shot and chunked upload).
type Provider interface {
	// Name identifies the provider in logs and in the UI.
	Name() string
	// Configured reports whether enough settings are present to try an upload.
	Configured() bool
	// Upload stores body under the given file name.
	Upload(name string, body io.Reader, size int64, contentType string) error
	// Destination is a human-readable, non-secret description of where uploads
	// go. It is shown in the UI, so it must never include credentials.
	Destination() string
}

// Kind selects a provider implementation.
type Kind string

const (
	KindNone    Kind = ""
	KindS3      Kind = "s3"
	KindDropbox Kind = "dropbox"
	KindGDrive  Kind = "gdrive"
	KindWebDAV  Kind = "webdav"
)

// ParseKind maps the BACKUP_PROVIDER setting to a Kind.
//
// Unknown values are an error rather than a silent fallback: an operator who
// typed "dropbx" should be told, not left believing backups are being uploaded.
func ParseKind(s string) (Kind, error) {
	switch Kind(strings.ToLower(strings.TrimSpace(s))) {
	case KindNone:
		return KindNone, nil
	case KindS3:
		return KindS3, nil
	case KindDropbox:
		return KindDropbox, nil
	case KindGDrive:
		return KindGDrive, nil
	case KindWebDAV:
		return KindWebDAV, nil
	default:
		return KindNone, fmt.Errorf("unknown backup provider %q (expected s3, dropbox, gdrive or webdav)", s)
	}
}

// Settings carries every provider's configuration. Only the fields relevant to
// the selected Kind are read.
type Settings struct {
	Kind Kind

	// S3-compatible (AWS S3, MinIO, Backblaze B2, Wasabi, Cloudflare R2).
	S3 Config

	// Dropbox: a long-lived access token, or a refresh token plus app
	// credentials. Refresh tokens are preferred - Dropbox access tokens now
	// expire after a few hours, so a token-only setup stops working silently.
	DropboxAccessToken  string
	DropboxRefreshToken string
	DropboxAppKey       string
	DropboxAppSecret    string
	DropboxPath         string

	// Google Drive: OAuth2 refresh token plus client credentials. Drive has no
	// long-lived token equivalent, so this is the only supported shape.
	GDriveClientID     string
	GDriveClientSecret string
	GDriveRefreshToken string
	GDriveFolderID     string

	// WebDAV (Nextcloud, ownCloud, or any WebDAV server).
	WebDAVURL      string
	WebDAVUsername string
	WebDAVPassword string
}

// New builds the configured provider. It returns nil when no provider is
// selected, which callers treat as "cloud upload disabled".
func New(s Settings) Provider {
	switch s.Kind {
	case KindS3:
		return NewClient(s.S3)
	case KindDropbox:
		return NewDropbox(s)
	case KindGDrive:
		return NewGDrive(s)
	case KindWebDAV:
		return NewWebDAV(s)
	default:
		return nil
	}
}

// joinPath builds a clean remote path from a folder and a file name. Providers
// differ on leading slashes, so the caller normalises.
func joinPath(folder, name string) string {
	folder = strings.Trim(folder, "/")
	if folder == "" {
		return name
	}
	return folder + "/" + name
}
