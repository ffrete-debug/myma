package player

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"ark-server-commander/config"
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/rcon"
	"ark-server-commander/service/server"
	"ark-server-commander/utils"
)

// PlayerService manages player data for ARK servers.
type PlayerService struct {
	serverService *server.ServerService
}

// NewPlayerService creates a PlayerService with the default ServerService.
func NewPlayerService() *PlayerService {
	return &PlayerService{serverService: server.NewServerService()}
}

// GetPlayers fetches the list of online players for a server via RCON.
// It also persists the player records in the DB so they can be queried later.
func (s *PlayerService) GetPlayers(userID uint, serverID string) (models.PlayerListResponse, error) {
	id, err := utils.ParseUint(serverID)
	if err != nil {
		return models.PlayerListResponse{}, fmt.Errorf("invalid server ID: %w", err)
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return models.PlayerListResponse{}, fmt.Errorf("server not found: %w", err)
	}

	// A stopped server has no RCON listener, which is expected - not a server
	// fault. Returning 500 with a raw dial error made the players page look
	// broken whenever the selected server happened to be stopped, which is the
	// default when the first server in the list is not running.
	if server.Status != "running" {
		return models.PlayerListResponse{
			ServerID:    server.ID,
			Identifier:  server.Identifier,
			SessionName: server.SessionName,
			Online:      []models.OnlinePlayer{},
			TotalOnline: 0,
			MaxPlayers:  server.MaxPlayers,
			NotRunning:  true,
		}, nil
	}

	// Use RCON to get the player list from the running server
	output, err := rcon.ExecuteCommand(config.RCONHost, server.RCONPort, server.AdminPassword, "listplayers")
	if err != nil {
		// The server is marked running but RCON is unreachable - usually still
		// booting. Report it as a state, not a failure, so the page can say so.
		return models.PlayerListResponse{
			ServerID:    server.ID,
			Identifier:  server.Identifier,
			SessionName: server.SessionName,
			Online:      []models.OnlinePlayer{},
			TotalOnline: 0,
			MaxPlayers:  server.MaxPlayers,
			Unreachable: true,
		}, nil
	}

	online := ParseListPlayersOutput(output)

	// Persist online players in DB (upsert by steam_id + server_id)
	for _, p := range online {
		var existing models.Player
		err := database.DB.Where("server_id = ? AND steam_id = ?", server.ID, p.SteamID).First(&existing).Error
		if err != nil {
			// New player — create
			player := models.Player{
				ServerID:    server.ID,
				Name:        p.Name,
				SteamID:     p.SteamID,
				CharacterID: p.CharacterID,
				Status:      "online",
				JoinedAt:    time.Now().UTC(),
			}
			database.DB.Create(&player)
		} else {
			// Existing player — update status and timestamps
			existing.Name = p.Name
			existing.CharacterID = p.CharacterID
			existing.Status = "online"
			existing.JoinedAt = time.Now().UTC()
			database.DB.Save(&existing)
		}
	}

	// Mark any previously online players for this server as offline if not in current list
	var allServerPlayers []models.Player
	database.DB.Where("server_id = ?", server.ID).Find(&allServerPlayers)
	onlineSteamIDs := make(map[string]bool)
	for _, p := range online {
		onlineSteamIDs[p.SteamID] = true
	}
	for _, ap := range allServerPlayers {
		if !onlineSteamIDs[ap.SteamID] && ap.Status == "online" {
			ap.Status = "offline"
			database.DB.Save(&ap)
		}
	}

	return models.PlayerListResponse{
		ServerID:    server.ID,
		Identifier:  server.Identifier,
		SessionName: server.SessionName,
		Online:      online,
		TotalOnline: len(online),
		MaxPlayers:  server.MaxPlayers,
	}, nil
}

// GetPlayersHistory returns the DB-stored player records for a server (online + recent offline).
func (s *PlayerService) GetPlayersHistory(userID uint, serverID string) ([]models.Player, error) {
	id, err := utils.ParseUint(serverID)
	if err != nil {
		return nil, fmt.Errorf("invalid server ID: %w", err)
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	var players []models.Player
	database.DB.Where("server_id = ?", server.ID).Order("updated_at DESC").Limit(200).Find(&players)
	return players, nil
}

// ParseListPlayersOutput parses the raw RCON output of the `listplayers` command.
// ParseListPlayersOutput turns an ARK "listplayers" RCON response into player
// records. Exported so the metrics service can reuse it for population counts
// without duplicating the parsing rules.
//
// ARK is not consistent across versions/forks. Both of these are seen:
//
//  0. PlayerName, 76561198000000000
//  1. "PlayerName" [76561198000000000] (charid) - Duration: 1h 23m
//
// and an empty server answers with a prose line such as "No Players Connected".
//
// A line is only accepted as a player if it carries at least one positive
// signal: a 17-digit SteamID64, a "N." list index, or a quoted name. Without
// that check the prose empty-server reply was parsed as a player named
// "No Players Connected", so an empty server reported one player online.
func ParseListPlayersOutput(output string) []models.OnlinePlayer {
	// Non-nil so an empty result marshals to [] rather than null. A nil slice
	// becomes JSON null, and the UI read .length off it and crashed the whole
	// players page as soon as a server with nobody online was selected.
	players := make([]models.OnlinePlayer, 0)

	steamIDRe := regexp.MustCompile(`(\d{17})`)
	charIDRe := regexp.MustCompile(`\(([^)]+)\)`)
	nameRe := regexp.MustCompile(`"([^"]+)"`)
	indexRe := regexp.MustCompile(`^(\d+)[.)]\s*`)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || isNoPlayersLine(line) {
			continue
		}

		steamID := ""
		if m := steamIDRe.FindStringSubmatch(line); m != nil {
			steamID = m[1]
		}

		indexed := indexRe.MatchString(line)
		quoted := nameRe.MatchString(line)
		// A player line carries at least one positive signal: a SteamID, a "N."
		// list index, or a quoted name. Prose such as "No Players Connected"
		// has none of the three, and used to be parsed as a player.
		if steamID == "" && !indexed && !quoted {
			continue
		}

		body := indexRe.ReplaceAllString(line, "")

		name := ""
		if m := nameRe.FindStringSubmatch(body); m != nil {
			name = m[1]
		} else if steamID != "" {
			// "Name, <id>" or "Name [<id>]": the name is whatever precedes the
			// id, minus the separator.
			if idx := strings.Index(body, steamID); idx > 0 {
				name = strings.TrimRight(strings.TrimSpace(body[:idx]), ",[ \t")
				name = strings.TrimSpace(name)
			}
		} else {
			name = strings.TrimSpace(strings.Split(body, ",")[0])
		}

		charID := ""
		if m := charIDRe.FindStringSubmatch(body); m != nil {
			charID = m[1]
		}

		// An indexed line with neither a name nor an id carries no information.
		if name == "" && steamID == "" {
			continue
		}

		players = append(players, models.OnlinePlayer{
			Name:        name,
			SteamID:     steamID,
			CharacterID: charID,
		})
	}

	return players
}

// noPlayersMarkers are the prose replies ARK gives for an empty server. They
// vary by version, so match on substrings rather than exact strings.
var noPlayersMarkers = []string{
	"no players connected",
	"no players online",
	"there are no players",
}

func isNoPlayersLine(line string) bool {
	lower := strings.ToLower(line)
	for _, m := range noPlayersMarkers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

// StoreDBPlayers persists a batch of player records for a server (used by the RCON polling service).
func (s *PlayerService) StoreDBPlayers(serverID uint, online []models.OnlinePlayer) error {
	for _, p := range online {
		var existing models.Player
		err := database.DB.Where("server_id = ? AND steam_id = ?", serverID, p.SteamID).First(&existing).Error
		if err != nil {
			player := models.Player{
				ServerID:    serverID,
				Name:        p.Name,
				SteamID:     p.SteamID,
				CharacterID: p.CharacterID,
				Status:      "online",
				JoinedAt:    time.Now().UTC(),
			}
			database.DB.Create(&player)
		} else {
			existing.Name = p.Name
			existing.CharacterID = p.CharacterID
			existing.Status = "online"
			existing.JoinedAt = time.Now().UTC()
			database.DB.Save(&existing)
		}
	}
	return nil
}
