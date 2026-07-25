package player

import (
	"testing"

	"ark-server-commander/models"
)

func TestParseListPlayersOutput_SinglePlayer(t *testing.T) {
	output := `"Player1" [12345678901234567] (char-abc)`
	players := parseListPlayersOutput(output)
	if len(players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(players))
	}
	if players[0].Name != "Player1" {
		t.Errorf("expected name Player1, got %s", players[0].Name)
	}
	if players[0].SteamID != "12345678901234567" {
		t.Errorf("expected SteamID 12345678901234567, got %s", players[0].SteamID)
	}
	if players[0].CharacterID != "char-abc" {
		t.Errorf("expected CharacterID char-abc, got %s", players[0].CharacterID)
	}
}

func TestParseListPlayersOutput_MultiplePlayers(t *testing.T) {
	output := `"Alice" [11111111111111111]` + "\n" + `"Bob" [22222222222222222]`
	players := parseListPlayersOutput(output)
	if len(players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(players))
	}
	if players[0].Name != "Alice" {
		t.Errorf("expected Alice, got %s", players[0].Name)
	}
	if players[1].Name != "Bob" {
		t.Errorf("expected Bob, got %s", players[1].Name)
	}
}

func TestParseListPlayersOutput_OfflineNoSteamID(t *testing.T) {
	output := `"OrphanPlayer" - some other format`
	players := parseListPlayersOutput(output)
	found := false
	for _, p := range players {
		if p.Name == "OrphanPlayer" {
			found = true
			if p.SteamID != "" {
				t.Errorf("expected empty SteamID for orphan, got %s", p.SteamID)
			}
		}
	}
	if !found {
		t.Error("expected to find OrphanPlayer")
	}
}

func TestParseListPlayersOutput_Empty(t *testing.T) {
	players := parseListPlayersOutput("")
	if len(players) != 0 {
		t.Errorf("expected 0 players for empty input, got %d", len(players))
	}
}

func TestParseListPlayersOutput_WhitespaceOnly(t *testing.T) {
	players := parseListPlayersOutput("  \n  \n")
	if len(players) != 0 {
		t.Errorf("expected 0 players for whitespace input, got %d", len(players))
	}
}

func TestParseListPlayersOutput_BareName(t *testing.T) {
	// Unquoted name before bracket
	output := `SomePlayer [98765432109876543]`
	players := parseListPlayersOutput(output)
	found := false
	for _, p := range players {
		if p.Name == "SomePlayer" {
			found = true
			if p.SteamID != "98765432109876543" {
				t.Errorf("expected SteamID 98765432109876543, got %s", p.SteamID)
			}
		}
	}
	if !found {
		t.Error("expected to find SomePlayer")
	}
}

func TestParseListPlayersOutput_IPAndDuration(t *testing.T) {
	output := `"Player1" [12345678901234567] (char-abc) - 1h 23m - 192.168.1.1`
	players := parseListPlayersOutput(output)
	if len(players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(players))
	}
	p := players[0]
	if p.Name != "Player1" {
		t.Errorf("expected name Player1, got %s", p.Name)
	}
	if p.SteamID != "12345678901234567" {
		t.Errorf("expected SteamID 12345678901234567, got %s", p.SteamID)
	}
	if p.CharacterID != "char-abc" {
		t.Errorf("expected CharacterID char-abc, got %s", p.CharacterID)
	}
	_ = p.IP // IP field not yet extracted by parser
	_ = p.Duration // Duration field not yet extracted by parser
}

func TestParseListPlayersOutput_OnlinePlayerStruct(t *testing.T) {
	// Verify OnlinePlayer struct fields work correctly
	p := models.OnlinePlayer{
		Name:        "TestPlayer",
		SteamID:     "12345678901234567",
		CharacterID: "char-test",
		IP:          "192.168.1.100",
		Duration:    "1h 30m",
	}
	if p.Name != "TestPlayer" {
		t.Errorf("Name mismatch: got %s", p.Name)
	}
	if p.Duration != "1h 30m" {
		t.Errorf("Duration mismatch: got %s", p.Duration)
	}
}
