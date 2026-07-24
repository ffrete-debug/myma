package update

import (
	"ark-server-commander/models"
	"ark-server-commander/websocket"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UpdateService struct {
	db  *gorm.DB
	hub *websocket.Hub
}

func NewUpdateService(db *gorm.DB, hub *websocket.Hub) *UpdateService {
	return &UpdateService{db: db, hub: hub}
}

func (s *UpdateService) StartUpdate(serverID uint, step models.UpdateStep, message string) {
	status := models.UpdateStatus{
		ServerID:  serverID,
		Step:      string(step),
		Progress:  0,
		Message:   message,
		StartedAt: time.Now(),
	}
	if err := s.db.Create(&status).Error; err != nil {
		zap.L().Error("Failed to create update status", zap.Error(err))
		return
	}
	s.hub.BroadcastToServer(serverID, status)
}

func (s *UpdateService) UpdateProgress(serverID uint, progress int, message string) {
	var status models.UpdateStatus
	if err := s.db.Where("server_id = ?", serverID).Last(&status).Error; err != nil {
		return
	}
	status.Progress = progress
	status.Message = message
	s.db.Save(&status)
	s.hub.BroadcastToServer(serverID, status)
}

func (s *UpdateService) CompleteUpdate(serverID uint) {
	var status models.UpdateStatus
	if err := s.db.Where("server_id = ?", serverID).Last(&status).Error; err != nil {
		return
	}
	now := time.Now()
	status.Progress = 100
	status.CompletedAt = &now
	s.db.Save(&status)
	s.hub.BroadcastToServer(serverID, status)
}

func (s *UpdateService) GetUpdateStatus(serverID uint) (*models.UpdateStatus, error) {
	var status models.UpdateStatus
	if err := s.db.Where("server_id = ?", serverID).Last(&status).Error; err != nil {
		return nil, err
	}
	return &status, nil
}
