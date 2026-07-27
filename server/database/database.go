package database

import (
	"os"
	"strings"

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

	// SQL statements carry credentials (admin passwords, password hashes),
	// so only log them when LOG_LEVEL explicitly asks for debug output
	gormLogLevel := logger.Warn
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		gormLogLevel = logger.Info
	}

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open(config.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})

	if err != nil {
		utils.Fatal("Database connection failed", zap.Error(err))
	}

	// Auto-migrate database schema
	err = DB.AutoMigrate(&models.User{}, &models.Server{}, &models.AuditLog{}, &models.Player{}, &models.Backup{})
	if err != nil {
		utils.Fatal("Database migration failed", zap.Error(err))
	}

	utils.Info("Database initialized successfully", zap.String("db_path", config.DBPath))
}

func GetDB() *gorm.DB {
	return DB
}
