package config

import (
	"os"
	"testing"
)

// Limits are opt-in: unset must stay 0 (unlimited), which is Docker's own
// default and what every existing deployment already relies on.
func TestResourceLimitsDefaultToUnlimited(t *testing.T) {
	withEnv(t, map[string]string{
		"SERVER_MEMORY_LIMIT_MB": "",
		"SERVER_CPU_LIMIT":       "",
	}, func() {
		ServerMemoryLimitMB, ServerCPULimit = 0, 0
		loadResourceLimits()

		if ServerMemoryLimitMB != 0 {
			t.Errorf("memory limit = %d, want 0 (unlimited)", ServerMemoryLimitMB)
		}
		if ServerCPULimit != 0 {
			t.Errorf("cpu limit = %v, want 0 (unlimited)", ServerCPULimit)
		}
	})
}

func TestResourceLimitsReadFromEnv(t *testing.T) {
	withEnv(t, map[string]string{
		"SERVER_MEMORY_LIMIT_MB": "4096",
		"SERVER_CPU_LIMIT":       "2.5",
	}, func() {
		ServerMemoryLimitMB, ServerCPULimit = 0, 0
		loadResourceLimits()

		if ServerMemoryLimitMB != 4096 {
			t.Errorf("memory limit = %d, want 4096", ServerMemoryLimitMB)
		}
		if ServerCPULimit != 2.5 {
			t.Errorf("cpu limit = %v, want 2.5", ServerCPULimit)
		}
	})
}

// A typo in one of these must not stop the manager from booting - it falls back
// to unlimited rather than failing startup.
func TestInvalidResourceLimitsAreIgnored(t *testing.T) {
	for _, bad := range []string{"abc", "-1", "0", "4GB"} {
		withEnv(t, map[string]string{
			"SERVER_MEMORY_LIMIT_MB": bad,
			"SERVER_CPU_LIMIT":       bad,
		}, func() {
			ServerMemoryLimitMB, ServerCPULimit = 0, 0
			loadResourceLimits()

			if ServerMemoryLimitMB != 0 {
				t.Errorf("memory limit %q accepted as %d, want ignored", bad, ServerMemoryLimitMB)
			}
			if ServerCPULimit != 0 {
				t.Errorf("cpu limit %q accepted as %v, want ignored", bad, ServerCPULimit)
			}
		})
	}
}

func withEnv(t *testing.T, env map[string]string, body func()) {
	t.Helper()
	saved := map[string]string{}
	for k, v := range env {
		saved[k] = os.Getenv(k)
		if v == "" {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v)
		}
	}
	defer func() {
		for k, v := range saved {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}()
	body()
}
