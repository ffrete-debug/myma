package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
)

const BackupDir = "backups"

type BackupService struct{}

func NewBackupService() *BackupService {
	return &BackupService{}
}

func (s *BackupService) ensureBackupDir() error {
	return os.MkdirAll(BackupDir, 0755)
}

func (s *BackupService) ListBackups(userID uint) ([]models.BackupResponse, error) {
	var backups []models.Backup
	if err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&backups).Error; err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}

	var responses []models.BackupResponse
	for _, b := range backups {
		var serverName string
		var server models.Server
		if err := database.DB.Where("id = ? AND user_id = ?", b.ServerID, userID).First(&server).Error; err == nil {
			serverName = server.SessionName
		}

		responses = append(responses, models.BackupResponse{
			ID:         b.ID,
			ServerID:   b.ServerID,
			ServerName: serverName,
			Filename:   b.Filename,
			FileSize:   b.FileSize,
			Status:     b.Status,
			Error:      b.Error,
			CreatedAt:  b.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

func (s *BackupService) CreateBackup(userID uint, serverID string) (models.BackupResponse, error) {
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return models.BackupResponse{}, fmt.Errorf("invalid server id")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", uint(id), userID).First(&server).Error; err != nil {
		return models.BackupResponse{}, fmt.Errorf("server not found")
	}

	volumeName := utils.GetServerVolumeName(uint(id))

	backup := models.Backup{
		ServerID: uint(id),
		Filename: fmt.Sprintf("backup_%d_%s.tar.gz", id, time.Now().Format("20060102_150405")),
		Status:   "in_progress",
		UserID:   userID,
		FileSize: 0,
	}
	if err := database.DB.Create(&backup).Error; err != nil {
		return models.BackupResponse{}, fmt.Errorf("create backup record: %w", err)
	}

	go func() {
		s.createBackupAsync(backup.ID, volumeName, backup.Filename)
	}()

	return models.BackupResponse{
		ID:        backup.ID,
		ServerID:  backup.ServerID,
		Filename:  backup.Filename,
		Status:    backup.Status,
		CreatedAt: backup.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *BackupService) createBackupAsync(backupID uint, volumeName, filename string) {
	defer func() {
		if r := recover(); r != nil {
			database.DB.Model(&models.Backup{}).Where("id = ?", backupID).Updates(map[string]interface{}{
				"status": "failed",
				"error":  fmt.Sprintf("panic: %v", r),
			})
		}
	}()

	if err := s.ensureBackupDir(); err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backupID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	backupDirAbs, err := filepath.Abs(BackupDir)
	if err != nil {
		backupDirAbs = BackupDir
	}

	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backupID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	err = dockerManager.BackupVolume(volumeName, backupDirAbs, filename)
	if err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backupID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	// Get file size
	filePath := filepath.Join(backupDirAbs, filename)
	info, err := os.Stat(filePath)
	if err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backupID).Updates(map[string]interface{}{
			"status":    "completed",
			"file_size": 0,
		})
		return
	}

	database.DB.Model(&models.Backup{}).Where("id = ?", backupID).Updates(map[string]interface{}{
		"status":    "completed",
		"file_size": info.Size(),
	})

	utils.Infof("Backup created: %s (%d bytes)", filename, info.Size())
}

func (s *BackupService) DeleteBackup(userID uint, backupID string) error {
	id, err := strconv.ParseUint(backupID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid backup id")
	}

	var backup models.Backup
	if err := database.DB.Where("id = ? AND user_id = ?", uint(id), userID).First(&backup).Error; err != nil {
		return fmt.Errorf("backup not found")
	}

	filePath := filepath.Join(BackupDir, backup.Filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete backup file: %w", err)
	}

	if err := database.DB.Delete(&backup).Error; err != nil {
		return fmt.Errorf("delete backup record: %w", err)
	}

	return nil
}
