package images

import (
	"net/http"

	"ark-server-commander/service/server"

	"github.com/gin-gonic/gin"
)

var serverService = server.NewServerService()

// PullImage Pull Docker image
// @Summary Docker
// @Description User
// @Tags Image Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body map[string]string true "Image information {\"image_name\": \"tbro98/ase-server:latest\"}"
// @Success 200 {object} map[string]interface{} "Status"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /images/pull [post]
func PullImage(c *gin.Context) {
	var req struct {
		ImageName string `json:"image_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	err := serverService.PullImage(req.ImageName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": " On ",
		"data": map[string]interface{}{
			"image_name": req.ImageName,
			"status":     "pulling",
		},
	})
}

// Check ImageUpdates Check for image updates
// @Summary YesNo
// @Description Check all managed images for new versions
// @Tags Image Management
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]bool "Image update status map"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /images/check-updates [get]
func CheckImageUpdates(c *gin.Context) {
	data, err := serverService.CheckImageUpdates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": " ",
		"data":    data,
	})
}

// UpdateImage updates a Docker image
// @Summary Update Docker image
// @Description Update the specified image and handle affected containers
// @Tags Image Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body map[string]string true "Image information {\"image_name\": \"tbro98/ase-server:latest\"}"
// @Success 200 {object} map[string]interface{} "Update status"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /images/update [post]
func UpdateImage(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		ImageName string `json:"image_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	affectedServers, err := serverService.UpdateImage(req.ImageName, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": " On ",
		"data": map[string]interface{}{
			"image_name":       req.ImageName,
			"affected_servers": affectedServers,
			"status":           "updating",
		},
	})
}

// GetAffectedServers returns servers using the given image
// @Summary Server list
// @Description Server list
// @Tags Image Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param image_name query string true "Image name"
// @Success 200 {object} map[string]interface{} "Server list"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /images/affected [get]
func GetAffectedServers(c *gin.Context) {
	userID := c.GetUint("user_id")
	imageName := c.Query("image_name")

	if imageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image name "})
		return
	}

	servers, err := serverService.GetAffectedServers(imageName, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data": map[string]interface{}{
			"image_name": imageName,
			"servers":    servers,
		},
	})
}

// GetImageStatus Get image status
// @Summary Get image status
// @Description Get image status (images pulled asynchronously)
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Image status info"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /images/status [get]
func GetImageStatus(c *gin.Context) {
	data, err := serverService.GetImageStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    data,
	})
}
