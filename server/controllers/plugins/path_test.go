package plugins

import (
	"path/filepath"
	"strings"
	"testing"
)

// pluginsRoot stands in for the container side plugins volume that cleanPath's
// result is joined onto
const pluginsRoot = "/ark/plugins"

// TestCleanPathContainsTraversal is the path traversal guard. cleanPath
// normalises rather than refuses — filepath.Clean resolves ".." away once the
// path is rooted, so it returns "/" for ".." instead of an error. The property
// that actually matters is containment: the result must stay rooted at "/",
// must not keep a ".." segment, and must not join outside the plugins root.
// Both outcomes are accepted so that hardening cleanPath into a hard reject
// does not break this test
func TestCleanPathContainsTraversal(t *testing.T) {
	cases := []string{
		"..",
		"../..",
		"/..",
		"/../etc/passwd",
		"/mods/../../etc/passwd",
		"/mods/./../../",
		"....//....//etc/passwd",
		"/etc/passwd",
		"%2e%2e/%2e%2e/etc/passwd",
		"..%2f..%2fetc%2fpasswd",
		"..\\..\\etc\\passwd",
	}

	for _, in := range cases {
		clean, err := cleanPath(in)
		if err != nil {
			// Refusing outright is also a safe answer
			continue
		}

		if !strings.HasPrefix(clean, "/") {
			t.Errorf("cleanPath(%q) = %q, want a path rooted at /", in, clean)
			continue
		}
		for _, seg := range strings.Split(clean, "/") {
			if seg == ".." {
				t.Errorf("cleanPath(%q) = %q, which still carries a traversal segment", in, clean)
			}
		}

		joined := filepath.Join(pluginsRoot, clean)
		if joined != pluginsRoot && !strings.HasPrefix(joined, pluginsRoot+"/") {
			t.Errorf("cleanPath(%q) = %q escapes the plugins root (joins to %q)", in, clean, joined)
		}
	}
}

func TestCleanPathAllowsLegitimatePaths(t *testing.T) {
	cases := map[string]string{
		"":                  "/",
		"/":                 "/",
		"mods":              "/mods",
		"/mods":             "/mods",
		"/mods/":            "/mods",
		"/mods/sub":         "/mods/sub",
		"/mods/sub/":        "/mods/sub",
		"/mods//sub":        "/mods/sub",
		"/mods/./sub":       "/mods/sub",
		"/mods/sub/../othr": "/mods/othr",
	}

	for in, want := range cases {
		got, err := cleanPath(in)
		if err != nil {
			t.Errorf("cleanPath(%q) returned an error for a legitimate path: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("cleanPath(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestSafeFileNameRejectsTraversal covers the upload file name guard. Like
// cleanPath it reduces rather than refuses, so the invariant is that whatever
// comes back is a single harmless component
func TestSafeFileNameRejectsTraversal(t *testing.T) {
	hostile := []string{
		"",
		".",
		"..",
		"../evil.dll",
		"../../etc/passwd",
		"/etc/passwd",
		"sub/evil.dll",
		"..\\..\\evil.dll",
		"C:\\Windows\\evil.dll",
	}

	for _, name := range hostile {
		got, err := safeFileName(name)
		if err != nil {
			// Refusing outright is also a safe answer
			continue
		}
		if strings.ContainsAny(got, `/\`) {
			t.Errorf("safeFileName(%q) = %q, which is not a single path component", name, got)
		}
		if got == "" || got == "." || got == ".." {
			t.Errorf("safeFileName(%q) = %q, which escapes the destination directory", name, got)
		}
	}

	if got, err := safeFileName("Plugin.dll"); err != nil || got != "Plugin.dll" {
		t.Errorf("safeFileName(%q) = %q, %v; want %q, nil", "Plugin.dll", got, err, "Plugin.dll")
	}
}
