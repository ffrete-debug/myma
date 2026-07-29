package models

import (
	"time"

	"gorm.io/gorm"
)

type Player struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	ServerID    uint           `json:"server_id" gorm:"index:idx_server_id;"`
	Name        string         `json:"name"`
	SteamID     string         `json:"steam_id" gorm:"index:idx_steam_id;"`
	CharacterID string         `json:"character_id"`
	Status      string         `json:"status" gorm:"default:'online'"`
	IP          string         `json:"ip"`
	JoinedAt    time.Time      `json:"joined_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	Server Server `json:"server" gorm:"foreignKey:ServerID"`
}

type PlayerListResponse struct {
	ServerID    uint           `json:"server_id"`
	Identifier  string         `json:"identifier"`
	SessionName string         `json:"session_name"`
	Online      []OnlinePlayer `json:"online"`
	TotalOnline int            `json:"total_online"`
	MaxPlayers  int            `json:"max_players"`

	// The server is stopped: an empty list is expected, not an error.
	NotRunning bool `json:"not_running,omitempty"`
	// The server is marked running but RCON did not answer - usually still
	// booting.
	Unreachable bool `json:"unreachable,omitempty"`
}

type OnlinePlayer struct {
	Name        string `json:"name"`
	SteamID     string `json:"steam_id"`
	CharacterID string `json:"character_id,omitempty"`
	IP          string `json:"ip,omitempty"`
	Duration    string `json:"duration,omitempty"`
}
