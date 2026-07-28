package metrics

import (
	"net/http"
	"strconv"

	metricsservice "ark-server-commander/service/metrics"

	"github.com/gin-gonic/gin"
)

var svc = metricsservice.NewService()

// GetAllMetrics godoc
// @Summary Resource and player metrics for every server the caller owns
// @Tags metrics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func GetAllMetrics(c *gin.Context) {
	userID := c.GetUint("user_id")

	all, err := svc.GetAllServerMetrics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to collect metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    all,
	})
}

// GetServerMetrics godoc
// @Summary Resource and player metrics for a single server
// @Tags metrics
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]interface{}
// @Router /metrics/{id} [get]
func GetServerMetrics(c *gin.Context) {
	userID := c.GetUint("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server id"})
		return
	}

	m, err := svc.GetServerMetrics(uint(id), userID)
	if err != nil {
		// Not-found and not-owned are deliberately the same response so the
		// endpoint cannot be used to enumerate other users' server IDs.
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    m,
	})
}
