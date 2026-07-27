package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFolderExists_Directory(t *testing.T) {
	if !FolderExists(t.TempDir()) {
		t.Error("FolderExists(temp dir) should be true")
	}
}

func TestFolderExists_Missing(t *testing.T) {
	if FolderExists(filepath.Join(t.TempDir(), "missing")) {
		t.Error("FolderExists(missing) should be false")
	}
}

func TestFolderExists_RegularFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "regular.txt")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if FolderExists(path) {
		t.Error("FolderExists(regular file) should be false")
	}
}

// A path whose parent component is a regular file makes os.Stat fail with
// ENOTDIR, an error os.IsNotExist does not match. The nil os.FileInfo must
// not be dereferenced.
func TestFolderExists_StatError(t *testing.T) {
	file := filepath.Join(t.TempDir(), "regular.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	path := filepath.Join(file, "child")
	if _, err := os.Stat(path); err == nil || os.IsNotExist(err) {
		t.Skipf("stat(%s) did not produce a non-IsNotExist error", path)
	}

	if FolderExists(path) {
		t.Error("FolderExists(stat error) should be false")
	}
}
