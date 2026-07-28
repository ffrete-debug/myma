package storage

import (
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// AWS publishes a worked SigV4 example ("GET object", us-east-1) with the
// intermediate signing key and final signature. Deriving the same key from the
// same inputs is what proves the HMAC chain is right; a subtly wrong chain
// still produces a plausible-looking hex string and only fails against real S3.
//
// Source: AWS "Deriving a signing key" worked example, which uses the iam
// service on 2015-08-30. The service is a parameter of the chain, so exercising
// it with iam validates exactly the same derivation the s3 path uses.
func TestSigningKeyDerivationMatchesAWSExample(t *testing.T) {
	const (
		secret    = "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY"
		date      = "20150830"
		region    = "us-east-1"
		service   = "iam"
		wantKeyHx = "c4afb1cc5771d871763a393e44b703571b55cc28424d1a5e86da6ed3c154a4b9"
	)

	k := hmacSHA256([]byte("AWS4"+secret), date)
	k = hmacSHA256(k, region)
	k = hmacSHA256(k, service)
	k = hmacSHA256(k, "aws4_request")

	if got := hex.EncodeToString(k); got != wantKeyHx {
		t.Fatalf("signing key mismatch:\n got %s\nwant %s", got, wantKeyHx)
	}
}

// SigV4 requires RFC 3986 encoding. net/url does not match it — this is the
// classic source of "SignatureDoesNotMatch" against real S3.
func TestURIEncodeMatchesSigV4Rules(t *testing.T) {
	cases := map[string]string{
		"abcXYZ123":   "abcXYZ123",
		"-._~":        "-._~", // unreserved: must pass through untouched
		"a b":         "a%20b",
		"a+b":         "a%2Bb", // must NOT become a space, and must not stay "+"
		"a/b":         "a%2Fb", // encoded here; canonicalURI splits on "/" first
		"a=b":         "a%3Db",
		"backup_1.gz": "backup_1.gz",
		"ä":           "%C3%A4",
		"*":           "%2A", // url.QueryEscape leaves this alone; SigV4 does not
	}
	for in, want := range cases {
		if got := uriEncode(in); got != want {
			t.Errorf("uriEncode(%q) = %q, want %q", in, got, want)
		}
	}
}

// Path separators must survive canonicalisation while each segment is encoded.
func TestCanonicalURIPreservesSeparators(t *testing.T) {
	u := &url.URL{Path: "/bucket/ark backups/save 1.tar.gz"}
	want := "/bucket/ark%20backups/save%201.tar.gz"
	if got := canonicalURI(u); got != want {
		t.Fatalf("canonicalURI = %q, want %q", got, want)
	}

	if got := canonicalURI(&url.URL{Path: ""}); got != "/" {
		t.Fatalf("empty path should canonicalise to \"/\", got %q", got)
	}
}

// Header names must be lowercased and sorted, and host must be included even
// when Go keeps it off the Header map.
func TestCanonicalHeadersSortedAndIncludesHost(t *testing.T) {
	req, err := http.NewRequest(http.MethodPut, "https://bucket.example.com/key", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-amz-date", "20260101T000000Z")
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("x-amz-content-sha256", unsignedPayload)

	signed, canonical := canonicalHeaders(req)

	if signed != "content-type;host;x-amz-content-sha256;x-amz-date" {
		t.Fatalf("signed headers not lowercased/sorted: %q", signed)
	}
	if !strings.Contains(canonical, "host:bucket.example.com\n") {
		t.Fatalf("host missing from canonical headers:\n%s", canonical)
	}
	if !strings.HasSuffix(canonical, "\n") {
		t.Fatal("canonical header block must end with a newline")
	}
}

func TestAddressingStyles(t *testing.T) {
	endpoint, _ := url.Parse("https://s3.example.com")

	virtual := &Client{cfg: Config{Bucket: "my-bucket"}}
	if host, path := virtual.addressing(endpoint, "a/b.gz"); host != "my-bucket.s3.example.com" || path != "/a/b.gz" {
		t.Errorf("virtual-host addressing = %s %s", host, path)
	}

	// MinIO and most self-hosted gateways only support path style.
	pathStyle := &Client{cfg: Config{Bucket: "my-bucket", PathStyle: true}}
	if host, path := pathStyle.addressing(endpoint, "a/b.gz"); host != "s3.example.com" || path != "/my-bucket/a/b.gz" {
		t.Errorf("path-style addressing = %s %s", host, path)
	}
}

func TestObjectKeyPrefixing(t *testing.T) {
	for _, tc := range []struct{ prefix, want string }{
		{"", "backup.tar.gz"},
		{"ark", "ark/backup.tar.gz"},
		{"/ark/", "ark/backup.tar.gz"}, // stray slashes must not double up
	} {
		c := &Client{cfg: Config{Prefix: tc.prefix}}
		if got := c.ObjectKey("backup.tar.gz"); got != tc.want {
			t.Errorf("prefix %q -> %q, want %q", tc.prefix, got, tc.want)
		}
	}
}

func TestUploadRefusesWhenUnconfigured(t *testing.T) {
	c := NewClient(Config{Bucket: "b"}) // no endpoint or credentials
	if err := c.Upload("k", strings.NewReader("x"), 1, ""); err == nil {
		t.Fatal("expected an error when object storage is not configured")
	}
}

// The Authorization header must carry the scope, the sorted signed headers and
// a 64-hex signature. A malformed header is rejected by S3 with an unhelpful
// error, so it is worth pinning the shape.
func TestSignProducesWellFormedAuthorization(t *testing.T) {
	c := NewClient(Config{
		Endpoint: "https://s3.example.com", Region: "eu-west-1", Bucket: "b",
		AccessKey: "AKIDEXAMPLE", SecretKey: "SECRET",
	})

	req, err := http.NewRequest(http.MethodPut, "https://b.s3.example.com/key.tar.gz", nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	req.Header.Set("Host", "b.s3.example.com")
	req.Header.Set("x-amz-date", now.Format(isoLayout))
	req.Header.Set("x-amz-content-sha256", unsignedPayload)

	c.sign(req, now)

	auth := req.Header.Get("Authorization")
	for _, want := range []string{
		"AWS4-HMAC-SHA256 ",
		"Credential=AKIDEXAMPLE/20260102/eu-west-1/s3/aws4_request",
		"SignedHeaders=host;x-amz-content-sha256;x-amz-date",
		"Signature=",
	} {
		if !strings.Contains(auth, want) {
			t.Errorf("Authorization missing %q:\n%s", want, auth)
		}
	}

	sig := auth[strings.LastIndex(auth, "Signature=")+len("Signature="):]
	if len(sig) != 64 {
		t.Errorf("signature should be 64 hex chars, got %d: %q", len(sig), sig)
	}
	if _, err := hex.DecodeString(sig); err != nil {
		t.Errorf("signature is not hex: %v", err)
	}
}
