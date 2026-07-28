package servers

import (
	"net/http"
	"strconv"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/middleware"
	"ark-server-commander/models"
	backupservice "ark-server-commander/service/backup"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ownsServer reports whether the caller owns the server. Ownership is checked
// on every schedule operation, since a schedule drives work against a volume.
func ownsServer(serverID, userID uint) bool {
	var count int64
	database.DB.Model(&models.Server{}).
		Where("id = ? AND user_id = ?", serverID, userID).
		Count(&count)
	return count > 0
}

func scheduleServerID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return 0, false
	}
	return uint(id), true
}

// GetBackupSchedule godoc
// @Summary Get a server's automated backup schedule
// @Tags backups
// @Produce json
// @Router /servers/{id}/backup-schedule [get]
func GetBackupSchedule(c *gin.Context) {
	serverID, ok := scheduleServerID(c)
	if !ok {
		return
	}
	userID := c.GetUint("user_id")

	if !ownsServer(serverID, userID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	var sched models.BackupSchedule
	err := database.DB.Where("server_id = ?", serverID).First(&sched).Error
	if err == gorm.ErrRecordNotFound {
		// No schedule yet is a normal state, not an error. Returning the
		// disabled defaults lets the UI render its form without special-casing.
		c.JSON(http.StatusOK, gin.H{
			"message": "Operation successful",
			"data": models.BackupSchedule{
				ServerID: serverID, UserID: userID,
				Enabled: false, IntervalHours: 24, RetainCount: 7,
			},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Operation successful", "data": sched})
}

// UpsertBackupSchedule godoc
// @Summary Create or update a server's automated backup schedule
// @Tags backups
// @Produce json
// @Router /servers/{id}/backup-schedule [put]
func UpsertBackupSchedule(c *gin.Context) {
	serverID, ok := scheduleServerID(c)
	if !ok {
		return
	}
	userID := c.GetUint("user_id")

	if !ownsServer(serverID, userID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	var req models.BackupScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interval_hours must be between 1 and 168"})
		return
	}

	// Cloud upload cannot be enabled without somewhere to upload to; failing
	// here is clearer than accepting the setting and failing silently at 3am.
	if req.UploadToCloud && !backupservice.CloudConfigured() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cloud backup is not configured (set BACKUP_PROVIDER and its credentials)",
		})
		return
	}

	var sched models.BackupSchedule
	err := database.DB.Where("server_id = ?", serverID).First(&sched).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		sched = models.BackupSchedule{ServerID: serverID, UserID: userID}
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load schedule"})
		return
	}

	sched.Enabled = req.Enabled
	sched.IntervalHours = req.IntervalHours
	sched.RetainCount = req.RetainCount
	sched.UploadToCloud = req.UploadToCloud

	// Project the next run so the UI can show it immediately rather than
	// waiting for the scheduler's first tick.
	if sched.Enabled {
		base := time.Now()
		if sched.LastRunAt != nil {
			base = *sched.LastRunAt
		}
		next := base.Add(time.Duration(sched.IntervalHours) * time.Hour)
		sched.NextRunAt = &next
	} else {
		sched.NextRunAt = nil
	}

	if err := database.DB.Save(&sched).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save schedule"})
		return
	}

	middleware.Log.Log(userID, "backup.schedule", "server", c.Param("id"), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Operation successful", "data": sched})
}

// UploadBackupToCloud godoc
// @Summary Upload an existing backup to object storage on demand
// @Tags backups
// @Produce json
// @Router /backups/{id}/upload [post]
func UploadBackupToCloud(c *gin.Context) {
	userID := c.GetUint("user_id")

	backupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid backup id"})
		return
	}

	var backup models.Backup
	if err := database.DB.Where("id = ? AND user_id = ?", backupID, userID).First(&backup).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "backup not found"})
		return
	}
	if backup.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backup is not complete"})
		return
	}

	if err := backupservice.UploadBackupToCloud(backup.Filename); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	middleware.Log.Log(userID, "backup.upload", "backup", backup.Filename, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Operation successful"})
}

// GetCloudStorageStatus godoc
// @Summary Report whether object storage is configured
// @Tags backups
// @Produce json
// @Router /backups/cloud-status [get]
func GetCloudStorageStatus(c *gin.Context) {
	// Credentials are deliberately not echoed back — only whether a destination
	// is usable, which provider it is, and a non-secret description of it.
	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data": gin.H{
			"configured":  backupservice.CloudConfigured(),
			"provider":    backupservice.CloudProviderName(),
			"destination": backupservice.CloudDestination(),
		},
	})
}
