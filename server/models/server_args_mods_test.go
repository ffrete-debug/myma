package models

import (
	"strings"
	"testing"
)

func baseServer() Server {
	return Server{Map: "TheIsland", Port: 7777, QueryPort: 27015, RCONPort: 27020}
}

// -mods= is an ASA/CurseForge flag. ASE ignores it entirely (docs/wiki.md:358,
// "inASE = No"), so emitting it did nothing at all — mods configured in the UI
// never reached the server. The game image installs Workshop mods itself from
// the GameModIds environment variable instead.
func TestNoModsFlagIsEmitted(t *testing.T) {
	sa := NewServerArgs()

	for _, ids := range [][]string{nil, {}, {"111", "222"}} {
		got := sa.GenerateArgsStringWithMods(baseServer(), ids)
		if strings.Contains(got, "-mods=") {
			t.Errorf("-mods= must never be emitted (ASE ignores it), got:\n%s", got)
		}
	}
}

// When the server record already carries GameModIds, that is what travels — the
// resolved list must not be duplicated onto the launch line.
func TestGameModIdsFromServerRecordWins(t *testing.T) {
	sa := NewServerArgs()
	server := baseServer()
	server.GameModIds = "900,901"

	got := sa.GenerateArgsStringWithMods(server, []string{"111", "222"})

	if !strings.Contains(got, "?GameModIds=900,901") {
		t.Errorf("expected the server's GameModIds, got:\n%s", got)
	}
	if strings.Contains(got, "111") {
		t.Errorf("resolved list must not override the stored field, got:\n%s", got)
	}
}

// Fallback: if the field has not been synced yet, the resolved list still
// reaches the server through the correct ASE channel rather than being lost.
func TestResolvedModsFallBackToGameModIdsQueryParam(t *testing.T) {
	sa := NewServerArgs()

	got := sa.GenerateArgsStringWithMods(baseServer(), []string{"111", "222", "333"})

	if !strings.Contains(got, "?GameModIds=111,222,333") {
		t.Fatalf("expected ?GameModIds= with the resolved list, got:\n%s", got)
	}
}

// SessionName must NOT appear on the command line. The image runs
// `$PROTON run ShooterGameServer.exe ${SERVER_ARGS}` with SERVER_ARGS expanded
// UNQUOTED, so a name containing a space word-splits the launch string: the name
// is truncated at the first space and every query parameter after it is dropped.
// It lives in GameUserSettings.ini instead.
func TestSessionNameIsNotOnTheCommandLine(t *testing.T) {
	sa := NewServerArgs()
	server := baseServer()
	server.SessionName = "My Cool Server"

	got := sa.GenerateArgsStringWithMods(server, nil)

	if strings.Contains(got, "SessionName") {
		t.Errorf("SessionName must not reach SERVER_ARGS, got:\n%s", got)
	}
	if strings.Contains(got, "My Cool Server") {
		t.Errorf("session name text leaked into the launch string:\n%s", got)
	}
}

// The generated string must be byte-stable. Go randomises map iteration, which
// made SERVER_ARGS differ on every call: the container drift check compared
// unequal strings and rebuilt the container on every single start, and which
// parameters survived was non-deterministic.
func TestGeneratedArgsAreDeterministic(t *testing.T) {
	sa := NewServerArgs()
	sa.QueryParams["ServerPVE"] = "True"
	sa.QueryParams["AllowFlyerCarry"] = "False"
	sa.CommandLineArgs["useexclusivelist"] = true
	sa.CommandLineArgs["crossplay"] = true

	first := sa.GenerateArgsStringWithMods(baseServer(), []string{"111"})
	for i := 0; i < 25; i++ {
		if got := sa.GenerateArgsStringWithMods(baseServer(), []string{"111"}); got != first {
			t.Fatalf("args are not deterministic:\n run 0: %s\n run %d: %s", first, i+1, got)
		}
	}
}

// No argument may contain a space, for the same word-splitting reason.
func TestNoGeneratedArgumentContainsASpace(t *testing.T) {
	sa := NewServerArgs()
	server := baseServer()
	server.SessionName = "Name With Spaces"
	server.AdminPassword = "hunter2"

	got := sa.GenerateArgsStringWithMods(server, []string{"111"})

	// The map/query section is the first whitespace-delimited token; every
	// following token must be a flag.
	for _, field := range strings.Fields(got) {
		if strings.HasPrefix(field, "-") || strings.Contains(field, "?") || field == server.Map {
			continue
		}
		t.Errorf("unexpected bare token %q in launch string — a value contains a space:\n%s", field, got)
	}
}

func TestPlainBuilderMatchesNoModsVariant(t *testing.T) {
	sa := NewServerArgs()
	server := baseServer()

	if a, b := sa.GenerateArgsString(server), sa.GenerateArgsStringWithMods(server, nil); a != b {
		t.Fatalf("builders diverged:\n plain: %s\n mods : %s", a, b)
	}
}
