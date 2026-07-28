package storage

import (
	"strings"
	"testing"
)

func TestParseKind(t *testing.T) {
	valid := map[string]Kind{
		"":        KindNone,
		"s3":      KindS3,
		"S3":      KindS3,
		"  s3  ":  KindS3,
		"dropbox": KindDropbox,
		"Dropbox": KindDropbox,
		"gdrive":  KindGDrive,
		"webdav":  KindWebDAV,
	}
	for in, want := range valid {
		got, err := ParseKind(in)
		if err != nil {
			t.Errorf("ParseKind(%q) errored: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseKind(%q) = %q, want %q", in, got, want)
		}
	}

	// A typo must be an error, not a silent fallback to "disabled" — an
	// operator who mistyped should be told, not left believing backups upload.
	for _, in := range []string{"dropbx", "google", "s4", "aws", "drive"} {
		if _, err := ParseKind(in); err == nil {
			t.Errorf("ParseKind(%q) should have errored", in)
		}
	}
}

func TestNewReturnsNilWhenDisabled(t *testing.T) {
	if p := New(Settings{Kind: KindNone}); p != nil {
		t.Fatalf("expected nil provider when disabled, got %T", p)
	}
}

func TestProvidersReportUnconfiguredWhenEmpty(t *testing.T) {
	for _, kind := range []Kind{KindS3, KindDropbox, KindGDrive, KindWebDAV} {
		p := New(Settings{Kind: kind})
		if p == nil {
			t.Errorf("%s: expected a provider instance", kind)
			continue
		}
		if p.Configured() {
			t.Errorf("%s: reported configured with no settings", kind)
		}
		if p.Name() == "" {
			t.Errorf("%s: empty Name()", kind)
		}
	}
}

func TestProvidersReportConfiguredWhenComplete(t *testing.T) {
	cases := map[Kind]Settings{
		KindS3: {Kind: KindS3, S3: Config{
			Endpoint: "https://s3.test", Bucket: "b", AccessKey: "k", SecretKey: "s",
		}},
		// A bare access token is enough for Dropbox, though a refresh token is
		// preferred since access tokens now expire.
		KindDropbox: {Kind: KindDropbox, DropboxAccessToken: "tok"},
		KindGDrive: {Kind: KindGDrive,
			GDriveClientID: "id", GDriveClientSecret: "sec", GDriveRefreshToken: "ref"},
		KindWebDAV: {Kind: KindWebDAV,
			WebDAVURL: "https://dav.test/backups", WebDAVUsername: "u", WebDAVPassword: "p"},
	}
	for kind, settings := range cases {
		p := New(settings)
		if !p.Configured() {
			t.Errorf("%s: expected configured", kind)
		}
		if p.Destination() == "" {
			t.Errorf("%s: expected a non-empty destination description", kind)
		}
	}
}

// Dropbox refresh-token auth needs all three parts; a partial setup must not
// report itself ready.
func TestDropboxRefreshTokenNeedsAppCredentials(t *testing.T) {
	partial := New(Settings{Kind: KindDropbox, DropboxRefreshToken: "ref"})
	if partial.Configured() {
		t.Error("refresh token alone should not count as configured")
	}

	full := New(Settings{Kind: KindDropbox,
		DropboxRefreshToken: "ref", DropboxAppKey: "key", DropboxAppSecret: "sec"})
	if !full.Configured() {
		t.Error("refresh token with app credentials should be configured")
	}
}

// Destination strings are shown in the UI, so they must never leak secrets.
func TestDestinationsExcludeCredentials(t *testing.T) {
	secrets := []string{"SUPERSECRET", "tok-abc", "refresh-xyz"}

	dests := []string{
		New(Settings{Kind: KindS3, S3: Config{
			Endpoint: "https://s3.test", Bucket: "b", AccessKey: "AKIA", SecretKey: "SUPERSECRET",
		}}).Destination(),
		New(Settings{Kind: KindDropbox, DropboxAccessToken: "tok-abc", DropboxPath: "/ark"}).Destination(),
		New(Settings{Kind: KindGDrive,
			GDriveClientID: "id", GDriveClientSecret: "SUPERSECRET",
			GDriveRefreshToken: "refresh-xyz", GDriveFolderID: "folder1"}).Destination(),
		New(Settings{Kind: KindWebDAV,
			WebDAVURL: "https://dav.test/backups", WebDAVUsername: "u", WebDAVPassword: "SUPERSECRET",
		}).Destination(),
	}

	for _, d := range dests {
		for _, secret := range secrets {
			if strings.Contains(d, secret) {
				t.Errorf("destination %q leaks a credential", d)
			}
		}
	}
}

func TestJoinPath(t *testing.T) {
	cases := []struct{ folder, name, want string }{
		{"", "b.tar.gz", "b.tar.gz"},
		{"ark", "b.tar.gz", "ark/b.tar.gz"},
		{"/ark/", "b.tar.gz", "ark/b.tar.gz"},
		{"ark/saves", "b.tar.gz", "ark/saves/b.tar.gz"},
	}
	for _, c := range cases {
		if got := joinPath(c.folder, c.name); got != c.want {
			t.Errorf("joinPath(%q,%q) = %q, want %q", c.folder, c.name, got, c.want)
		}
	}
}
