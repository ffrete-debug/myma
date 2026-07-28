package utils

import (
	"strings"
	"testing"
)

// SessionName moved off the command line because the game image expands
// SERVER_ARGS unquoted, so a space truncated the name and dropped every query
// parameter after it. The INI is now authoritative, which makes this upsert the
// thing that keeps the Steam browser name correct.
func TestEnsureSessionNameReplacesExistingInBothSections(t *testing.T) {
	ini := `[ServerSettings]
SessionName=Old Name
MaxPlayers=70

[SessionSettings]
SessionName=Old Name
`
	out := EnsureSessionName(ini, "My Cool Server")

	if strings.Contains(out, "Old Name") {
		t.Errorf("old name survived:\n%s", out)
	}
	if got := strings.Count(out, "SessionName=My Cool Server"); got != 2 {
		t.Errorf("expected the name in both sections, found %d:\n%s", got, out)
	}
	if !strings.Contains(out, "MaxPlayers=70") {
		t.Error("unrelated keys must be preserved")
	}
}

// A name with spaces is the whole point — it must round-trip intact.
func TestEnsureSessionNamePreservesSpaces(t *testing.T) {
	out := EnsureSessionName("[SessionSettings]\nSessionName=x\n", "Ark  Of  Spaces")
	if !strings.Contains(out, "SessionName=Ark  Of  Spaces") {
		t.Fatalf("spaces not preserved:\n%s", out)
	}
}

func TestEnsureSessionNameAddsKeyWhenSectionPresentButKeyMissing(t *testing.T) {
	ini := "[ServerSettings]\nMaxPlayers=70\n\n[SessionSettings]\nOtherKey=1\n"
	out := EnsureSessionName(ini, "Newly Named")

	if got := strings.Count(out, "SessionName=Newly Named"); got != 2 {
		t.Fatalf("expected 2 insertions, got %d:\n%s", got, out)
	}
	if !strings.Contains(out, "OtherKey=1") || !strings.Contains(out, "MaxPlayers=70") {
		t.Errorf("existing keys lost:\n%s", out)
	}
}

func TestEnsureSessionNameAddsMissingSections(t *testing.T) {
	out := EnsureSessionName("[/Script/Engine.GameSession]\nMaxPlayers=10\n", "Fresh")

	for _, section := range []string{"[ServerSettings]", "[SessionSettings]"} {
		if !strings.Contains(out, section) {
			t.Errorf("missing section %s:\n%s", section, out)
		}
	}
	if got := strings.Count(out, "SessionName=Fresh"); got != 2 {
		t.Errorf("expected 2 occurrences, got %d:\n%s", got, out)
	}
}

// A key in some unrelated section must not be touched.
func TestEnsureSessionNameIgnoresOtherSections(t *testing.T) {
	ini := "[SomethingElse]\nSessionName=Leave Me\n\n[SessionSettings]\nSessionName=x\n"
	out := EnsureSessionName(ini, "Renamed")

	if !strings.Contains(out, "SessionName=Leave Me") {
		t.Errorf("unrelated section was modified:\n%s", out)
	}
}

func TestEnsureSessionNameEmptyIsNoOp(t *testing.T) {
	ini := "[SessionSettings]\nSessionName=Keep\n"
	if out := EnsureSessionName(ini, ""); out != ini {
		t.Fatalf("empty name should be a no-op, got:\n%s", out)
	}
}
