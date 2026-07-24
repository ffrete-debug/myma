package rcon

import (
	"fmt"
	"net/http"

	"ark-server-commander/middleware"
	"ark-server-commander/service/server"
	"ark-server-commander/utils"
	"github.com/gin-gonic/gin"
)

var serverService = server.NewServerService()

type RCONRequest struct {
	Command string `json:"command" binding:"required"`
}

func ExecuteRCON(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	var req RCONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err.Error())
		return
	}

	output, err := serverService.ExecuteRCONCommand(userID, serverID, req.Command)
	if err != nil {
		middleware.Log.Log(userID, "rcon.execute", fmt.Sprintf("server:%s", serverID), req.Command, c.ClientIP())
		utils.InternalError(c, "RCON execution failed", err.Error())
		return
	}

	middleware.Log.Log(userID, "rcon.execute", fmt.Sprintf("server:%s", serverID), req.Command, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Command executed",
		"data":    gin.H{"output": output},
	})
}
