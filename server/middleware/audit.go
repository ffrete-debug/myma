package middleware

import (
	"ark-server-commander/models"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Log *AuditLogger

func InitAudit(db *gorm.DB) {
	Log = NewAuditLogger(db)
}

// AuditLogger logs sensitive operations to AuditLog table
type AuditLogger struct {
	db *gorm.DB
}

func NewAuditLogger(db *gorm.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

func (a *AuditLogger) Log(userID uint, action, resource, detail, ip string) {
	entry := models.AuditLog{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Detail:    detail,
		IP:        ip,
		CreatedAt: time.Now(),
	}
	if err := a.db.Create(&entry).Error; err != nil {
		zap.L().Error("Audit log write failed", zap.Error(err))
	}
}
