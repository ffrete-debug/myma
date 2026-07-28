package player

import "testing"

// The reported bug: the metrics dashboard showed 1 player on an empty server.
// ARK answers "listplayers" with a prose line when nobody is connected, and the
// old parser turned that line into a player named "No Players Connected".
func TestParseListPlayersOutput_EmptyServerReportsZero(t *testing.T) {
	replies := []string{
		"No Players Connected",
		"no players connected",
		"No Players Connected\n",
		" There are no players connected. ",
		"No Players Online",
	}

	for _, reply := range replies {
		if got := ParseListPlayersOutput(reply); len(got) != 0 {
			t.Errorf("ParseListPlayersOutput(%q) = %d players, want 0 (%+v)", reply, len(got), got)
		}
	}
}

func TestParseListPlayersOutput_EmptyAndWhitespaceInput(t *testing.T) {
	for _, reply := range []string{"", "   ", "\n\n", "\t\n  \n"} {
		if got := ParseListPlayersOutput(reply); len(got) != 0 {
			t.Errorf("ParseListPlayersOutput(%q) = %d players, want 0", reply, len(got))
		}
	}
}

// The comma-separated form is what ASE dedicated servers commonly return.
func TestParseListPlayersOutput_CommaFormat(t *testing.T) {
	out := ParseListPlayersOutput("0. Rockwell, 76561198000000001\n1. Helena, 76561198000000002")

	if len(out) != 2 {
		t.Fatalf("got %d players, want 2: %+v", len(out), out)
	}
	if out[0].Name != "Rockwell" || out[0].SteamID != "76561198000000001" {
		t.Errorf("player 0 = %+v", out[0])
	}
	if out[1].Name != "Helena" || out[1].SteamID != "76561198000000002" {
		t.Errorf("player 1 = %+v", out[1])
	}
}

// A real player list must still parse when mixed with banner/prompt noise,
// which RCON consoles often prepend.
func TestParseListPlayersOutput_IgnoresSurroundingNoise(t *testing.T) {
	out := ParseListPlayersOutput(
		"Server received, But no response!!\n" +
			"0. Rockwell, 76561198000000001\n" +
			"Command executed\n",
	)

	if len(out) != 1 {
		t.Fatalf("got %d players, want 1: %+v", len(out), out)
	}
	if out[0].SteamID != "76561198000000001" {
		t.Errorf("unexpected player %+v", out[0])
	}
}

// Guards the specific arithmetic the dashboard does: an empty server must be
// distinguishable from an unreachable one, and it counts as zero, not one.
func TestParseListPlayersOutput_CountIsUsableForMetrics(t *testing.T) {
	if n := len(ParseListPlayersOutput("No Players Connected")); n != 0 {
		t.Fatalf("empty server counted as %d players", n)
	}
	if n := len(ParseListPlayersOutput("0. Rockwell, 76561198000000001")); n != 1 {
		t.Fatalf("one player counted as %d", n)
	}
}
