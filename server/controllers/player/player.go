package player

import (
	"net/http"

	"ark-server-commander/models"
	"ark-server-commander/service/player"
	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
)

var playerService = player.NewPlayerService()

// GetPlayers returns the online player list for a server, fetched via RCON.
// @Summary Get online players for a server
// @Description Fetches the online player list via RCON listplayers and persists records in DB
// @Tags Player Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Server ID"
// @Success 200 {object} map[string]interface{} "Player list"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/players [get]
func GetPlayers(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	result, err := playerService.GetPlayers(userID, serverID)
	if err != nil {
		utils.InternalError(c, "Failed to fetch players", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    result,
	})
}

// GetPlayersHistory returns the DB-stored player history for a server.
// @Summary Get player history for a server
// @Description Returns recent player records (online + offline) from the database
// @Tags Player Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Server ID"
// @Success 200 {object} map[string]interface{} "Player history"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/players/history [get]
func GetPlayersHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	players, err := playerService.GetPlayersHistory(userID, serverID)
	if err != nil {
		utils.InternalError(c, "Failed to fetch player history", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    players,
	})
}

// GetPlayerSummary returns a summary of online/offline player counts.
// @Summary Get player summary for a server
// @Description Returns online count, total tracked players, and recent offline players
// @Tags Player Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Server ID"
// @Success 200 {object} map[string]interface{} "Player summary"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/players/summary [get]
func GetPlayerSummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	result, err := playerService.GetPlayers(userID, serverID)
	if err != nil {
		utils.InternalError(c, "Failed to fetch player summary", err.Error())
		return
	}

	history, err := playerService.GetPlayersHistory(userID, serverID)
	if err != nil {
		utils.InternalError(c, "Failed to fetch player history", err.Error())
		return
	}

	var offlineCount int
	for _, p := range history {
		if p.Status == "offline" {
			offlineCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data": gin.H{
			"server_id":     result.ServerID,
			"identifier":    result.Identifier,
			"online_count":  result.TotalOnline,
			"max_players":   result.MaxPlayers,
			"tracked_total": len(history) + result.TotalOnline,
			"offline_count": offlineCount,
		},
	})
}

// PlayersResponse is a convenience alias for the standard API envelope
// when returning a slice of players.
type PlayersResponse struct {
	Players []models.Player `json:"players"`
	Total   int             `json:"total"`
}
