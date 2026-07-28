package models

import "gorm.io/gorm"

// ServerMod is a Steam Workshop item attached to a server.
//
// Load order matters in ARK: later mods override earlier ones, so Position is
// part of the record rather than an incidental sort. WorkshopID is a string
// because Steam publishedfileids are 64-bit and are handled as strings
// everywhere in Steam's own API.
type ServerMod struct {
	gorm.Model
	ServerID   uint   `json:"server_id" gorm:"not null;index;uniqueIndex:idx_server_workshop"`
	WorkshopID string `json:"workshop_id" gorm:"not null;uniqueIndex:idx_server_workshop"`
	Name       string `json:"name"`
	PreviewURL string `json:"preview_url"`
	// Position is 0-based and dense within a server. The service renumbers on
	// every mutation so gaps cannot accumulate.
	Position int  `json:"position" gorm:"not null;default:0"`
	Enabled  bool `json:"enabled" gorm:"not null;default:true"`
}

// ServerModRequest is the payload for attaching a mod to a server.
type ServerModRequest struct {
	WorkshopID string `json:"workshop_id" binding:"required"`
}

// ServerModReorderRequest carries the full desired order. Sending the complete
// list rather than a move instruction keeps the operation idempotent and avoids
// the client and server disagreeing about intermediate states.
type ServerModReorderRequest struct {
	WorkshopIDs []string `json:"workshop_ids" binding:"required"`
}

// WorkshopItem is a Steam Workshop entry as returned to the UI.
type WorkshopItem struct {
	WorkshopID    string `json:"workshop_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	PreviewURL    string `json:"preview_url"`
	FileSizeBytes int64  `json:"file_size_bytes"`
	Subscriptions int64  `json:"subscriptions"`
	TimeUpdated   int64  `json:"time_updated"`
}
