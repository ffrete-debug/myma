package audit

import (
	"net/http"
	"strconv"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
)

// GetAuditLogs returns paginated audit logs with optional filters.
// @Summary Get audit logs
// @Description List audit log entries with pagination and filtering
// @Tags Audit
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param user_id query int false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param start_date query string false "Filter by start date (RFC3339)"
// @Param end_date query string false "Filter by end date (RFC3339)"
// @Success 200 {object} map[string]interface{} "Audit log list with pagination"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /audit-logs [get]
func GetAuditLogs(c *gin.Context) {
	userID := c.GetUint("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	db := database.DB.Model(&models.AuditLog{})

	// Only admins can see all users' logs; regular users see their own
	role := c.GetString("role")
	if role != "admin" {
		db = db.Where("user_id = ?", userID)
	}

	if v := c.Query("user_id"); v != "" {
		if uid, err := strconv.ParseUint(v, 10, 32); err == nil {
			db = db.Where("user_id = ?", uid)
		}
	}
	if v := c.Query("action"); v != "" {
		db = db.Where("action = ?", v)
	}
	if v := c.Query("start_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if v := c.Query("end_date"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			db = db.Where("created_at <= ?", t)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		utils.InternalError(c, "Failed to count audit logs", err.Error())
		return
	}

	var logs []models.AuditLog
	if err := db.Order("created_at desc").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		utils.InternalError(c, "Failed to fetch audit logs", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    logs,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
