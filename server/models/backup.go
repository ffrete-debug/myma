package models

import (
	"time"
)

type Backup struct {
	ID          uint      `json:"id" gorm:"primarykey"`
	ServerID    uint      `json:"server_id" gorm:"not null;index"`
	Server      Server    `json:"server" gorm:"foreignKey:ServerID"`
	Filename    string    `json:"filename" gorm:"not null"`
	FileSize    int64     `json:"file_size" gorm:"default:0"`
	Status      string    `json:"status" gorm:"default:'completed'"` // completed, in_progress, failed
	Error       string    `json:"error,omitempty"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
}

type BackupRequest struct {
	ServerID string `json:"server_id" binding:"required"`
}

type BackupResponse struct {
	ID         uint   `json:"id"`
	ServerID   uint   `json:"server_id"`
	ServerName string `json:"server_name"`
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type RestoreRequest struct {
	BackupID string `json:"backup_id" binding:"required"`
}
