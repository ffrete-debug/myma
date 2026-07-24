package models

import "time"

type AuditLog struct {
	ID         uint      `json:"id" gorm:"primarykey"`
	UserID     uint      `json:"user_id" gorm:"index"`
	Action     string    `json:"action"`     // server.create, server.delete, image.pull, etc
	Resource   string    `json:"resource"`   // server:1, image:tbro98/ase-server
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}
