// Package storage uploads backup archives to S3-compatible object storage.
//
// This is a deliberately small client rather than the AWS SDK. It performs one
// operation — PUT an object — and pulling in aws-sdk-go-v2 for that would add a
// large dependency tree to a project whose CI audits dependencies. Signature
// correctness is covered by unit tests against AWS's own published test vectors.
//
// Works with any S3-compatible endpoint: AWS S3, MinIO, Backblaze B2, Wasabi,
// Cloudflare R2.
package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// unsignedPayload lets us stream a multi-gigabyte archive without buffering
	// it or reading it twice to compute a body hash. S3 permits this over
	// HTTPS, where TLS already protects integrity in transit.
	unsignedPayload = "UNSIGNED-PAYLOAD"
	isoLayout       = "20060102T150405Z"
	dateLayout      = "20060102"
	uploadTimeout   = 30 * time.Minute
)

// Config describes the destination bucket.
type Config struct {
	Endpoint  string // e.g. https://s3.eu-west-1.amazonaws.com or http://minio:9000
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	Prefix    string // optional key prefix, e.g. "ark-backups"
	// PathStyle addresses the bucket as <endpoint>/<bucket>/<key> instead of
	// <bucket>.<endpoint>/<key>. Required by MinIO and most self-hosted
	// gateways, which do not implement virtual-host addressing.
	PathStyle bool
}

// Valid reports whether enough is configured to attempt an upload.
func (c Config) Valid() bool {
	return c.Endpoint != "" && c.Bucket != "" && c.AccessKey != "" && c.SecretKey != ""
}

type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{Timeout: uploadTimeout}}
}

// ObjectKey returns the full key an object will be stored under.
func (c *Client) ObjectKey(name string) string {
	prefix := strings.Trim(c.cfg.Prefix, "/")
	if prefix == "" {
		return name
	}
	return prefix + "/" + name
}

// Upload streams body to the bucket under key.
//
// size must be the exact content length: S3 requires it, and a streaming body
// of unknown length would otherwise be sent chunked, which the signature does
// not cover here.
func (c *Client) Name() string { return "s3" }

// Configured satisfies Provider; Valid is kept for existing callers.
func (c *Client) Configured() bool { return c.cfg.Valid() }

// Destination is a non-secret description shown in the UI.
func (c *Client) Destination() string {
	if c.cfg.Prefix != "" {
		return fmt.Sprintf("%s/%s (%s)", c.cfg.Bucket, strings.Trim(c.cfg.Prefix, "/"), c.cfg.Endpoint)
	}
	return fmt.Sprintf("%s (%s)", c.cfg.Bucket, c.cfg.Endpoint)
}

// Upload stores body under the configured prefix. `name` is a bare file name;
// the prefix is applied here so every provider takes the same argument.
func (c *Client) Upload(name string, body io.Reader, size int64, contentType string) error {
	if !c.cfg.Valid() {
		return fmt.Errorf("object storage is not configured")
	}

	key := c.ObjectKey(name)

	endpoint, err := url.Parse(c.cfg.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %w", err)
	}

	host, path := c.addressing(endpoint, key)
	target := *endpoint
	target.Host = host
	target.Path = path

	req, err := http.NewRequest(http.MethodPut, target.String(), body)
	if err != nil {
		return fmt.Errorf("build upload request: %w", err)
	}
	req.ContentLength = size
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	now := time.Now().UTC()
	req.Header.Set("Host", host)
	req.Header.Set("x-amz-date", now.Format(isoLayout))
	req.Header.Set("x-amz-content-sha256", unsignedPayload)

	c.sign(req, now)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("upload to object storage: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// S3 error bodies are XML and can be long; a bounded excerpt is enough
		// to tell an operator whether it is credentials, a missing bucket or a
		// clock-skew problem.
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("object storage returned %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	return nil
}

// addressing picks between path-style and virtual-host-style URLs.
func (c *Client) addressing(endpoint *url.URL, key string) (host, path string) {
	base := strings.TrimSuffix(endpoint.Path, "/")
	if c.cfg.PathStyle {
		return endpoint.Host, base + "/" + c.cfg.Bucket + "/" + key
	}
	return c.cfg.Bucket + "." + endpoint.Host, base + "/" + key
}

// sign applies AWS Signature Version 4 to req.
func (c *Client) sign(req *http.Request, now time.Time) {
	amzDate := now.Format(isoLayout)
	shortDate := now.Format(dateLayout)
	scope := strings.Join([]string{shortDate, c.cfg.Region, "s3", "aws4_request"}, "/")

	signedHeaders, canonicalHeaders := canonicalHeaders(req)

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI(req.URL),
		canonicalQuery(req.URL),
		canonicalHeaders,
		signedHeaders,
		unsignedPayload,
	}, "\n")

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		hashHex([]byte(canonicalRequest)),
	}, "\n")

	// The signing key is derived by chaining HMACs over date, region, service
	// and the terminator — this is what makes a leaked signature useless
	// outside its day, region and service.
	k := hmacSHA256([]byte("AWS4"+c.cfg.SecretKey), shortDate)
	k = hmacSHA256(k, c.cfg.Region)
	k = hmacSHA256(k, "s3")
	k = hmacSHA256(k, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(k, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.cfg.AccessKey, scope, signedHeaders, signature,
	))
}

// canonicalHeaders returns the semicolon-joined signed header names and the
// canonical header block, both lowercased and sorted as SigV4 requires.
func canonicalHeaders(req *http.Request) (signed, canonical string) {
	names := make([]string, 0, len(req.Header)+1)
	values := map[string]string{}

	for name, vals := range req.Header {
		lower := strings.ToLower(name)
		names = append(names, lower)
		values[lower] = strings.TrimSpace(strings.Join(vals, ","))
	}
	if _, ok := values["host"]; !ok {
		names = append(names, "host")
		values["host"] = req.URL.Host
	}

	sortStrings(names)

	var b strings.Builder
	for _, n := range names {
		b.WriteString(n)
		b.WriteString(":")
		b.WriteString(values[n])
		b.WriteString("\n")
	}
	return strings.Join(names, ";"), b.String()
}

// canonicalURI percent-encodes each path segment. Object keys legitimately
// contain characters that must be escaped, and "/" must survive as a separator.
func canonicalURI(u *url.URL) string {
	if u.Path == "" {
		return "/"
	}
	segments := strings.Split(u.Path, "/")
	for i, s := range segments {
		segments[i] = uriEncode(s)
	}
	return strings.Join(segments, "/")
}

func canonicalQuery(u *url.URL) string {
	q := u.Query()
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sortStrings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		vals := q[k]
		sortStrings(vals)
		for _, v := range vals {
			parts = append(parts, uriEncode(k)+"="+uriEncode(v))
		}
	}
	return strings.Join(parts, "&")
}

// uriEncode implements the RFC 3986 encoding SigV4 requires. net/url's escapers
// do not match it: they leave some reserved characters alone and encode a space
// as "+" in query context, either of which breaks the signature.
func uriEncode(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= 'A' && ch <= 'Z',
			ch >= 'a' && ch <= 'z',
			ch >= '0' && ch <= '9',
			ch == '-', ch == '.', ch == '_', ch == '~':
			b.WriteByte(ch)
		default:
			fmt.Fprintf(&b, "%%%02X", ch)
		}
	}
	return b.String()
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func hashHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// sortStrings is a tiny insertion sort, used to avoid importing sort for the
// handful of short slices involved in signing.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
