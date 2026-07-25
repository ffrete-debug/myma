package player

import (
	"fmt"
	"regexp"
	"strings"
	"time"

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

	// Use RCON to get the player list from the running server
	output, err := rcon.ExecuteCommand("localhost", server.RCONPort, server.AdminPassword, "listplayers")
	if err != nil {
		return models.PlayerListResponse{}, fmt.Errorf("rcon listplayers failed: %w", err)
	}

	online := parseListPlayersOutput(output)

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

// parseListPlayersOutput parses the raw RCON output of the `listplayers` command.
// ARK RCON returns output in a line-based format like:
//   1. "PlayerName" (<steamid>) - Duration: 1h 23m 45s
// or similar variations with different delimiters.
func parseListPlayersOutput(output string) []models.OnlinePlayer {
	var players []models.OnlinePlayer

	lines := strings.Split(output, "\n")
	steamIDRe := regexp.MustCompile(`\[(\d{17})\]`)
	charIDRe := regexp.MustCompile(`\(([^)]+)\)`)
	nameRe := regexp.MustCompile(`"([^"]+)"`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to extract name (quoted or unquoted)
		name := ""
		if m := nameRe.FindStringSubmatch(line); m != nil {
			name = m[1]
		} else {
			// Fallback: take text before bracket or dash
			parts := strings.Split(line, "[")
			if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
				name = strings.TrimSpace(parts[0])
			}
		}

		// Try to extract SteamID
		steamID := ""
		if m := steamIDRe.FindStringSubmatch(line); m != nil {
			steamID = m[1]
		}

		// Try to extract CharacterID (from parens)
		charID := ""
		if m := charIDRe.FindStringSubmatch(line); m != nil {
			charID = m[1]
		}

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
