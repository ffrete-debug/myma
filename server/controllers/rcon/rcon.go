package rcon

import (
	"fmt"
	"net/http"

	"ark-server-commander/middleware"
	"ark-server-commander/service/server"
	"ark-server-commander/utils"
	"github.com/gin-gonic/gin"
)

// RCONExecutor is the subset of the ServerService the HTTP handler needs.
// Declared as an interface so tests can inject a fake without spinning up
// a database or hitting the Docker daemon.
type RCONExecutor interface {
	ExecuteRCONCommand(userID uint, serverID string, command string) (string, error)
}

// executor is the RCONExecutor used by the handler in production. Tests can
// swap it via setExecutorForTest (see rcon_test.go).
var executor RCONExecutor = server.NewServerService()

// setExecutorForTest replaces the package-level executor. Intended only for
// tests; not safe for concurrent use — keep tests sequential.
func setExecutorForTest(e RCONExecutor) {
	executor = e
}

type RCONRequest struct {
	Command string `json:"command" binding:"required"`
}

// ExecuteRCON handles POST /servers/:id/rcon/execute.
//
// Auth is enforced by AuthMiddleware on the route. The handler is intentionally
// thin: argument binding, service dispatch and audit logging. Command-content
// validation/rate-limiting lives in the WebSocket path; the HTTP endpoint is
// for occasional ad-hoc commands.
func ExecuteRCON(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	var req RCONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err.Error())
		return
	}

	output, err := executor.ExecuteRCONCommand(userID, serverID, req.Command)
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
