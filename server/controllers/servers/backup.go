package servers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/middleware"
	"ark-server-commander/models"
	backupsvc "ark-server-commander/service/backup"
	arkDocker "ark-server-commander/service/docker_manager"
	arkServer "ark-server-commander/service/server"
	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
)

var backupService = backupsvc.NewBackupService()

// ListBackups Get backups for a server
// @Summary List backups
// @Description List all backups for the current user
// @Tags Backup Management
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string][]models.BackupResponse "Backup list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /backups [get]
func ListBackups(c *gin.Context) {
	userID := c.GetUint("user_id")

	backups, err := backupService.ListBackups(userID)
	if err != nil {
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    backups,
	})
}

// CreateBackup Create a backup for a server
// @Summary Create backup
// @Description Create a backup of a server's volume data
// @Tags Backup Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body models.BackupRequest true "Backup request"
// @Success 201 {object} map[string]models.BackupResponse "Backup created"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /backups [post]
func CreateBackup(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.BackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err.Error())
		return
	}

	response, err := backupService.CreateBackup(userID, req.ServerID)
	if err != nil {
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	middleware.Log.Log(userID, "backup.create", fmt.Sprintf("server:%s", req.ServerID), "", c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{
		"message": "Backup created",
		"data":    response,
	})
}

// DeleteBackup Delete a backup
// @Summary Delete backup
// @Description Delete a backup and its file
// @Tags Backup Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Backup ID"
// @Success 200 {object} map[string]string "DeleteSuccess"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /backups/{id} [delete]
func DeleteBackup(c *gin.Context) {
	userID := c.GetUint("user_id")
	backupID := c.Param("id")

	err := backupService.DeleteBackup(userID, backupID)
	if err != nil {
		if err.Error() == "backup not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	middleware.Log.Log(userID, "backup.delete", fmt.Sprintf("backup:%s", backupID), "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Backup deleted successfully",
	})
}

// DownloadBackup Download a backup file
// @Summary Download backup
// @Description Download a backup file by filename
// @Tags Backup Management
// @Accept json
// @Produce application/gzip
// @Security Bearer
// @Param filename path string true "Backup filename"
// @Success 200 {file} binary "Backup file"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not found"
// @Router /backups/download/{filename} [get]
func DownloadBackup(c *gin.Context) {
	userID := c.GetUint("user_id")
	filename := c.Param("filename")

	// Validate filename to prevent path traversal
	if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		utils.BadRequest(c, "Invalid filename", "")
		return
	}

	// Ensure the backup belongs to the current user
	var backupRecord models.Backup
	if err := database.DB.Where("filename = ? AND user_id = ?", filename, userID).First(&backupRecord).Error; err != nil {
		utils.NotFound(c, "Resource not found", "Backup not found")
		return
	}

	filePath := filepath.Join(backupsvc.BackupDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		utils.NotFound(c, "Resource not found", "Backup file not found")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/gzip")
	c.File(filePath)
}

// RestoreBackup Restore a server from a backup
// @Summary Restore backup
// @Description Restore a server's volume from a backup file
// @Tags Backup Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body models.RestoreRequest true "Restore request"
// @Success 200 {object} map[string]string "Restore started"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /backups/restore [post]
func RestoreBackup(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.RestoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err.Error())
		return
	}

	id, err := strconv.ParseUint(req.BackupID, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid backup id", "")
		return
	}

	var backup models.Backup
	if err := database.DB.Where("id = ? AND user_id = ?", uint(id), userID).First(&backup).Error; err != nil {
		utils.NotFound(c, "Resource not found", "Backup not found")
		return
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", backup.ServerID, userID).First(&server).Error; err != nil {
		utils.NotFound(c, "Resource not found", "Server not found")
		return
	}

	go func() {
		restoreBackupAsync(backup, server)
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Restore started",
	})
}

func restoreBackupAsync(backup models.Backup, server models.Server) {
	defer func() {
		if r := recover(); r != nil {
			database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
				"status": "failed",
				"error":  fmt.Sprintf("panic: %v", r),
			})
		}
	}()

	// Update status to in_progress
	database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
		"status": "in_progress",
	})

	dm, err := arkDocker.GetDockerManager()
	if err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	// 1. Stop the server if running
	if server.Status == "running" {
		svc := arkServer.NewServerService()
		if err := svc.StopServer(backup.UserID, strconv.FormatUint(uint64(server.ID), 10)); err != nil {
			database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
				"status": "failed",
				"error":  "stop server: " + err.Error(),
			})
			return
		}
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			var s models.Server
			if err := database.DB.Where("id = ?", server.ID).First(&s).Error; err == nil && s.Status == "stopped" {
				break
			}
		}
	}

	volumeName := utils.GetServerVolumeName(server.ID)

	// Open the archive from the app's own filesystem and stream it in. It used
	// to be handed to Docker as a bind-mount source, which the daemon resolves
	// on the HOST - the same path confusion that made downloads 404.
	archive, err := os.Open(filepath.Join(backupsvc.BackupDir, backup.Filename))
	if err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "open backup archive: " + err.Error(),
		})
		return
	}
	defer func() { _ = archive.Close() }()

	// 2. Remove existing volume
	if err := dm.RemoveSingleVolume(volumeName); err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "remove volume: " + err.Error(),
		})
		return
	}

	// 3. Create new volume
	if err := dm.CreateSingleVolume(volumeName); err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "create volume: " + err.Error(),
		})
		return
	}

	// 4. Extract backup tar into new volume
	if err := dm.RestoreVolume(volumeName, backup.Filename, archive); err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "restore volume: " + err.Error(),
		})
		return
	}

	// 5 & 6. Recreate container and start
	containerName := utils.GetServerContainerName(server.ID)
	if err := dm.RemoveContainer(containerName); err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "remove container: " + err.Error(),
		})
		return
	}

	_, err = dm.CreateContainer(
		server.ID,
		server.Identifier,
		server.Port,
		server.QueryPort,
		server.RCONPort,
		server.AdminPassword,
		server.Map,
		server.GameModIds,
		server.AutoRestart,
	)
	if err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "create container: " + err.Error(),
		})
		return
	}

	if err := dm.StartContainer(containerName); err != nil {
		database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
			"status": "failed",
			"error":  "start container: " + err.Error(),
		})
		return
	}

	// 7. Update backup status with result
	database.DB.Model(&models.Backup{}).Where("id = ?", backup.ID).Updates(map[string]interface{}{
		"status": "completed",
	})
	utils.Infof("Restore backup %d for server %d completed", backup.ID, server.ID)
}
