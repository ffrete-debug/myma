package models

import (
	"strings"
	"testing"
)

// Mods must reach the launch line as a single comma-joined -mods= argument in
// the order given, because ARK resolves overrides by load order.
func TestGenerateArgsStringWithModsPreservesOrder(t *testing.T) {
	sa := NewServerArgs()
	server := Server{Map: "TheIsland", Port: 7777, QueryPort: 27015, RCONPort: 27020}

	got := sa.GenerateArgsStringWithMods(server, []string{"111", "222", "333"})

	if !strings.Contains(got, "-mods=111,222,333") {
		t.Fatalf("expected -mods=111,222,333 in args, got:\n%s", got)
	}
}

// No mods must not emit an empty -mods=, which ARK treats as a malformed
// argument rather than "no mods".
func TestGenerateArgsStringWithNoModsOmitsFlag(t *testing.T) {
	sa := NewServerArgs()
	server := Server{Map: "TheIsland", Port: 7777, QueryPort: 27015, RCONPort: 27020}

	for name, ids := range map[string][]string{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			if got := sa.GenerateArgsStringWithMods(server, ids); strings.Contains(got, "-mods=") {
				t.Fatalf("expected no -mods= flag, got:\n%s", got)
			}
		})
	}
}

// The plain builder must stay equivalent to the mod-aware one with no mods, so
// existing callers are unaffected.
func TestGenerateArgsStringMatchesNoModsVariant(t *testing.T) {
	sa := NewServerArgs()
	server := Server{Map: "Ragnarok", Port: 7777, QueryPort: 27015, RCONPort: 27020}

	if a, b := sa.GenerateArgsString(server), sa.GenerateArgsStringWithMods(server, nil); a != b {
		t.Fatalf("builders diverged:\n plain: %s\n mods : %s", a, b)
	}
}

// A user's explicit -mods= in CustomArgs must come after the generated one, so
// it still wins — matching how the rest of the builder treats custom args.
func TestCustomArgsOverrideGeneratedMods(t *testing.T) {
	sa := NewServerArgs()
	sa.CustomArgs = []string{"-mods=999"}
	server := Server{Map: "TheIsland", Port: 7777, QueryPort: 27015, RCONPort: 27020}

	got := sa.GenerateArgsStringWithMods(server, []string{"111"})

	generated := strings.Index(got, "-mods=111")
	custom := strings.Index(got, "-mods=999")
	if generated == -1 || custom == -1 {
		t.Fatalf("expected both -mods= arguments present, got:\n%s", got)
	}
	if custom < generated {
		t.Fatalf("custom -mods= must come last so it wins, got:\n%s", got)
	}
}
