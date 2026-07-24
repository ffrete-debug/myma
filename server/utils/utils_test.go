package utils

import "testing"

func TestValidateINIContent(t *testing.T) {
	err := ValidateINIContent("[Section]\nkey=value\n")
	if err != nil {
		t.Errorf("good INI should pass: %v", err)
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("test123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if !CheckPassword("test123", hash) {
		t.Error("should match")
	}
	if CheckPassword("wrong", hash) {
		t.Error("should not match")
	}
}

func TestGetServerContainerName(t *testing.T) {
	name := GetServerContainerName(1)
	if name != "ase-server-1" {
		t.Errorf("expected ase-server-1, got %s", name)
	}
}

func TestGetServerPluginsVolumeName(t *testing.T) {
	name := GetServerPluginsVolumeName(1)
	if name != "ase-server-plugins-1" {
		t.Errorf("expected ase-server-plugins-1, got %s", name)
	}
}
