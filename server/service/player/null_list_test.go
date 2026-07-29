package player

import (
	"encoding/json"
	"strings"
	"testing"
)

// A nil slice marshals to JSON null, not []. The players page read .length off
// that null and threw, crashing the entire page as soon as a server with nobody
// online was selected - which is every freshly started server.
func TestEmptyPlayerListMarshalsAsArrayNotNull(t *testing.T) {
	for _, reply := range []string{
		"No Players Connected",
		"",
		"   ",
		"Server received, But no response!!",
	} {
		players := ParseListPlayersOutput(reply)

		if players == nil {
			t.Errorf("ParseListPlayersOutput(%q) returned a nil slice", reply)
			continue
		}

		encoded, err := json.Marshal(players)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(encoded) != "[]" {
			t.Errorf("ParseListPlayersOutput(%q) marshalled to %s, want []", reply, encoded)
		}
	}
}

// The populated case must still marshal as a normal array.
func TestPopulatedPlayerListMarshalsAsArray(t *testing.T) {
	players := ParseListPlayersOutput("0. Rockwell, 76561198000000001")

	encoded, err := json.Marshal(players)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.HasPrefix(string(encoded), "[{") {
		t.Fatalf("expected a JSON array of objects, got %s", encoded)
	}
	if len(players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(players))
	}
}
