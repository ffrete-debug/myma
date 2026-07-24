package models

import "time"

type UpdateStatus struct {
	ID         uint      `json:"id" gorm:"primarykey"`
	ServerID   uint      `json:"server_id" gorm:"index"`
	Step       string    `json:"step"`
	Progress   int       `json:"progress"`
	Message    string    `json:"message"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type UpdateStep string

const (
	UpdateStepDownload  UpdateStep = "download"
	UpdateStepExtract   UpdateStep = "extract"
	UpdateStepConfigure UpdateStep = "configure"
	UpdateStepFinalize  UpdateStep = "finalize"
)
