package config

import (
	"os"
	"testing"
)

// The RCON host must never resolve to "localhost". Game servers run in their
// own containers and publish RCON on the Docker host, so "localhost" from
// inside this process points at the manager itself and every dial fails.
// That was a real bug: the RCON console could not connect at all.
func TestRCONHostIsNeverLocalhost(t *testing.T) {
	t.Setenv("JWT_SECRET", "Zx9Kq2Lp7Rn4Tv8Wm3Yb6Jd1Fh5Gs0Ac")
	t.Setenv("RCON_HOST", "")

	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig: %v", err)
	}

	if RCONHost == "localhost" {
		t.Fatal(`RCONHost resolved to "localhost", which can never reach a game server container`)
	}

	// Whichever branch was taken, it must be one of the two valid defaults.
	if RCONHost != "127.0.0.1" && RCONHost != dockerHostGateway {
		t.Fatalf("unexpected default RCONHost %q", RCONHost)
	}
}

func TestRCONHostEnvOverride(t *testing.T) {
	t.Setenv("JWT_SECRET", "Zx9Kq2Lp7Rn4Tv8Wm3Yb6Jd1Fh5Gs0Ac")
	t.Setenv("RCON_HOST", "10.0.0.42")

	if err := InitConfig(); err != nil {
		t.Fatalf("InitConfig: %v", err)
	}

	if RCONHost != "10.0.0.42" {
		t.Fatalf("RCON_HOST override ignored: got %q", RCONHost)
	}
}

// The auto-detect branch keys off /.dockerenv, which the daemon creates in
// every container it starts. Assert the helper agrees with the filesystem so a
// refactor cannot silently invert it.
func TestRunningInContainerMatchesDockerenv(t *testing.T) {
	_, err := os.Stat("/.dockerenv")
	if got, want := runningInContainer(), err == nil; got != want {
		t.Fatalf("runningInContainer() = %v, want %v", got, want)
	}
}
