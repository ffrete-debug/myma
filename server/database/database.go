package database

import (
	"os"
	"strings"
	"time"

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

	// Backups, restores, status refreshes and WebSocket handlers all write
	// concurrently, so the connection needs WAL (readers never block the
	// writer) and a busy timeout (writers wait instead of failing with
	// SQLITE_BUSY). The glebarez driver takes these as _pragma DSN params.
	dsn := config.DBPath
	if !strings.Contains(dsn, "?") {
		dsn += "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	}

	// Connect to SQLite database
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})

	if err != nil {
		utils.Fatal("Database connection failed", zap.Error(err))
	}

	// SQLite allows only one writer at a time; cap the pool so concurrent
	// goroutines queue on the Go side instead of piling up on the file lock
	sqlDB, err := DB.DB()
	if err != nil {
		utils.Fatal("Database handle unavailable", zap.Error(err))
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migrate database schema
	err = DB.AutoMigrate(&models.User{}, &models.Server{}, &models.AuditLog{}, &models.Player{}, &models.Backup{}, &models.ServerMod{}, &models.BackupSchedule{}, &models.UpdateStatus{})
	if err != nil {
		utils.Fatal("Database migration failed", zap.Error(err))
	}

	utils.Info("Database initialized successfully", zap.String("db_path", config.DBPath))
}

func GetDB() *gorm.DB {
	return DB
}
