package database

import (
	"ark-server-commander/config"
	"ark-server-commander/models"
	"ark-server-commander/utils"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		utils.Fatal("Database connection failed", zap.Error(err))
	}

	// Auto-migrate database schema
	err = DB.AutoMigrate(&models.User{}, &models.Server{}, &models.AuditLog{})
	if err != nil {
		utils.Fatal("Database migration failed", zap.Error(err))
	}

	utils.Info("Database initialized successfully", zap.String("db_path", config.DBPath))
}

func GetDB() *gorm.DB {
	return DB
}
