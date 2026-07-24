package rcon

import (
	"testing"
)

func TestExecuteCommand_InvalidHost(t *testing.T) {
	_, err := ExecuteCommand("invalid-host-nonexistent.example.com", 32330, "password", "status")
	if err == nil {
		t.Error("expected error for invalid host, got nil")
	}
}
