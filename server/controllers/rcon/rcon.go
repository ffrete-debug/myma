package rcon

import (
	"fmt"
	"net/http"

	"ark-server-commander/middleware"
	rconservice "ark-server-commander/service/rcon"
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

// ActionRequest is a structured admin operation from the UI.
//
// The command string is built server-side from the action and its parameters
// rather than being supplied by the caller: ARK's RCON is line-oriented, so a
// newline inside a message or player name would otherwise let a caller append
// a second, unintended command.
type ActionRequest struct {
	Action string            `json:"action" binding:"required"`
	Params map[string]string `json:"params"`
}

// ExecuteRCONAction handles POST /servers/:id/rcon/action.
func ExecuteRCONAction(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	var req ActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request", err.Error())
		return
	}

	command, err := rconservice.BuildCommand(rconservice.Action(req.Action), req.Params)
	if err != nil {
		utils.BadRequest(c, "Invalid action", err.Error())
		return
	}

	output, execErr := executor.ExecuteRCONCommand(userID, serverID, command)

	// The resolved command is audited rather than the raw request, so the log
	// records exactly what the server was asked to do.
	middleware.Log.Log(userID, "rcon.action", fmt.Sprintf("server:%s", serverID), command, c.ClientIP())

	if execErr != nil {
		utils.InternalError(c, "RCON execution failed", execErr.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Command executed",
		"data":    gin.H{"command": command, "output": output},
	})
}
