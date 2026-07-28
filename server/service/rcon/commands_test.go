package rcon

import (
	"strings"
	"testing"
)

func TestBuildCommandSimpleActions(t *testing.T) {
	cases := map[Action]string{
		ActionSaveWorld:        "SaveWorld",
		ActionListPlayers:      "ListPlayers",
		ActionDestroyWildDinos: "DestroyWildDinos",
		ActionDoExit:           "DoExit",
	}
	for action, want := range cases {
		got, err := BuildCommand(action, nil)
		if err != nil {
			t.Errorf("%s: unexpected error %v", action, err)
			continue
		}
		if got != want {
			t.Errorf("%s = %q, want %q", action, got, want)
		}
	}
}

// An unknown action must never fall through to something executable.
func TestBuildCommandRejectsUnknownAction(t *testing.T) {
	for _, a := range []Action{"", "cheat", "DoExit; DestroyWildDinos", "SaveWorld"} {
		if _, err := BuildCommand(a, nil); err == nil {
			t.Errorf("expected %q to be rejected", a)
		}
	}
}

// ARK's RCON is line-oriented, so a newline in a broadcast message would end
// the Broadcast command and start a second one. This is the injection guard.
func TestBroadcastRejectsCommandInjection(t *testing.T) {
	injections := []string{
		"hello\nDestroyWildDinos",
		"hello\rDoExit",
		"hello\x00DoExit",
		"hello\x1bDoExit",
		"hello\x7f",
	}
	for _, msg := range injections {
		if _, err := BuildCommand(ActionBroadcast, map[string]string{"message": msg}); err == nil {
			t.Errorf("expected %q to be rejected as injection", msg)
		}
	}
}

func TestBroadcastAcceptsOrdinaryText(t *testing.T) {
	got, err := BuildCommand(ActionBroadcast, map[string]string{
		"message": "Server restarting in 5 minutes!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Broadcast Server restarting in 5 minutes!" {
		t.Fatalf("got %q", got)
	}
}

func TestBroadcastRequiresNonEmptyBoundedMessage(t *testing.T) {
	if _, err := BuildCommand(ActionBroadcast, map[string]string{"message": "   "}); err == nil {
		t.Error("expected empty message to be rejected")
	}
	long := strings.Repeat("a", maxBroadcastLen+1)
	if _, err := BuildCommand(ActionBroadcast, map[string]string{"message": long}); err == nil {
		t.Error("expected over-long message to be rejected")
	}
}

// Player identifiers reach a command line, so anything non-numeric — notably a
// space, which would append an extra argument — must be rejected.
func TestPlayerActionsValidateSteamID(t *testing.T) {
	valid := "76561198000000000"
	for _, action := range []Action{ActionKickPlayer, ActionBanPlayer, ActionUnbanPlayer} {
		got, err := BuildCommand(action, map[string]string{"steam_id": valid})
		if err != nil {
			t.Errorf("%s: unexpected error %v", action, err)
			continue
		}
		if !strings.HasSuffix(got, " "+valid) {
			t.Errorf("%s = %q, expected it to end with the id", action, got)
		}
	}

	invalid := []string{
		"", "   ", "abc", "765 DoExit", "765\nDoExit", "765;DoExit",
		"-1", "76561198000000000000000000", "7.6e10",
	}
	for _, id := range invalid {
		if _, err := BuildCommand(ActionKickPlayer, map[string]string{"steam_id": id}); err == nil {
			t.Errorf("expected steam_id %q to be rejected", id)
		}
	}
}

func TestSetTimeOfDayValidatesFormat(t *testing.T) {
	if got, err := BuildCommand(ActionSetTimeOfDay, map[string]string{"time": "13:45"}); err != nil || got != "SetTimeOfDay 13:45" {
		t.Fatalf("got %q, err %v", got, err)
	}

	for _, bad := range []string{"", "25:00", "12:60", "1:00", "12-00", "12:00; DoExit", "noon"} {
		if _, err := BuildCommand(ActionSetTimeOfDay, map[string]string{"time": bad}); err == nil {
			t.Errorf("expected time %q to be rejected", bad)
		}
	}
}

// The UI gates a confirmation step on this, so the classification matters.
func TestDestructiveClassification(t *testing.T) {
	destructive := []Action{ActionDestroyWildDinos, ActionDoExit, ActionBanPlayer}
	for _, a := range destructive {
		if !a.Destructive() {
			t.Errorf("%s should be marked destructive", a)
		}
	}

	safe := []Action{ActionSaveWorld, ActionListPlayers, ActionBroadcast, ActionKickPlayer, ActionSetTimeOfDay}
	for _, a := range safe {
		if a.Destructive() {
			t.Errorf("%s should not be marked destructive", a)
		}
	}
}
