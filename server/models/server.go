package models

import (
	"time"

	"gorm.io/gorm"
)

type Server struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	Identifier    string         `json:"identifier" gorm:"not null"`
	SessionName   string         `json:"session_name" gorm:"default:'ARK Server'"` // Servers
	ClusterID     string         `json:"cluster_id" gorm:"default:''"`             // ID
	Port          int            `json:"port" gorm:"not null;default:7777"`
	QueryPort     int            `json:"query_port" gorm:"not null;default:27015"`
	RCONPort      int            `json:"rcon_port" gorm:"not null;default:32330"`
	AdminPassword string         `json:"admin_password" gorm:"not null;default:password"`
	Map           string         `json:"map" gorm:"default:'TheIsland'"`
	MaxPlayers    int            `json:"max_players" gorm:"not null;default:70"` // Max Players
	GameModIds    string         `json:"game_mod_ids" gorm:"default:''"`         // ID，
	Status        string         `json:"status" gorm:"default:'stopped'"`
	AutoRestart   bool           `json:"auto_restart" gorm:"default:true"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	User          User           `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	// Start（JSON）
	ServerArgsJSON string `json:"server_args_json" gorm:"default:'{}'"` // StartJSON
}

type ServerRequest struct {
	Identifier    string `json:"identifier" binding:"required"`
	SessionName   string `json:"session_name"` // Servers
	ClusterID     string `json:"cluster_id"`   // ID
	Port          int    `json:"port" binding:"required,min=1,max=65535"`
	QueryPort     int    `json:"query_port" binding:"required,min=1,max=65535"`
	RCONPort      int    `json:"rcon_port" binding:"required,min=1,max=65535"`
	AdminPassword string `json:"admin_password" binding:"required"`
	Map           string `json:"map"`
	MaxPlayers    int    `json:"max_players" binding:"min=1,max=200"` // Max Players
	GameModIds    string `json:"game_mod_ids"`                        // ID，
	AutoRestart   *bool  `json:"auto_restart"`                        // YesNoRestart（）
	// （）
	GameUserSettings string `json:"game_user_settings,omitempty"` // GameUserSettings.ini 
	GameIni          string `json:"game_ini,omitempty"`           // Game.ini 
	// Start（）
	ServerArgs *ServerArgsRequest `json:"server_args,omitempty"`
}

type ServerResponse struct {
	ID            uint   `json:"id"`
	Identifier    string `json:"identifier"`
	SessionName   string `json:"session_name"` // Servers
	ClusterID     string `json:"cluster_id"`   // ID
	Port          int    `json:"port"`
	QueryPort     int    `json:"query_port"`
	RCONPort      int    `json:"rcon_port"`
	AdminPassword string `json:"admin_password"`
	Map           string `json:"map"`
	MaxPlayers    int    `json:"max_players"` // Max Players
	GameModIds    string `json:"game_mod_ids"`
	Status        string `json:"status"`
	AutoRestart   bool   `json:"auto_restart"`
	UserID        uint   `json:"user_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	// 
	GameUserSettings string `json:"game_user_settings,omitempty"` // GameUserSettings.ini 
	GameIni          string `json:"game_ini,omitempty"`           // Game.ini 
	// Start
	ServerArgs    *ServerArgs `json:"server_args,omitempty"`    // Start
	GeneratedArgs string      `json:"generated_args,omitempty"` // Start
}

type ServerUpdateRequest struct {
	Identifier    string `json:"identifier"`
	SessionName   string `json:"session_name"` // Servers
	ClusterID     string `json:"cluster_id"`   // ID
	Port          int    `json:"port" binding:"min=1,max=65535"`
	QueryPort     int    `json:"query_port" binding:"min=1,max=65535"`
	RCONPort      int    `json:"rcon_port" binding:"min=1,max=65535"`
	AdminPassword string `json:"admin_password"`
	Map           string `json:"map"`
	MaxPlayers    int    `json:"max_players" binding:"min=1,max=200"` // Max Players
	GameModIds    string `json:"game_mod_ids"`                        // ID，
	AutoRestart   *bool  `json:"auto_restart"`
	// （）
	GameUserSettings string `json:"game_user_settings,omitempty"` // GameUserSettings.ini 
	GameIni          string `json:"game_ini,omitempty"`           // Game.ini 
	// Start（）
	ServerArgs *ServerArgsRequest `json:"server_args,omitempty"` // Start
}
