package servers

import (
	"fmt"
	"net/http"
	"strconv"

	"ark-server-commander/middleware"
	"ark-server-commander/models"
	"ark-server-commander/service/server"
	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
)

var serverService = server.NewServerService()

// GetServers Get server list
// @Summary Get server list
// @Description Get all servers for the current user
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string][]models.ServerResponse "Server list"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers [get]
func GetServers(c *gin.Context) {
	userID := c.GetUint("user_id")

	serverResponses, err := serverService.GetServers(userID)
	if err != nil {
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    serverResponses,
	})
}

// CreateServer CreateServers
// @Summary Create a new server
// @Description Create a new ARK server configuration
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param server body models.ServerRequest true "Server configuration"
// @Success 201 {object} map[string]models.ServerResponse "Created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers [post]
func CreateServer(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req models.ServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request parameters", "")
		return
	}

	if req.GameUserSettings != "" {
		if err := utils.ValidateINIContent(req.GameUserSettings); err != nil {
			utils.BadRequest(c, "GameUserSettings.ini format error", err.Error())
			return
		}
	}
	if req.GameIni != "" {
		if err := utils.ValidateINIContent(req.GameIni); err != nil {
			utils.BadRequest(c, "Game.ini format error", err.Error())
			return
		}
	}

	response, err := serverService.CreateServer(userID, req)
	if err != nil {
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	middleware.Log.Log(userID, "server.create", fmt.Sprintf("server:%d", response.ID), response.SessionName, c.ClientIP())

	c.JSON(http.StatusCreated, gin.H{
		"message": "Server created successfully",
		"data":    response,
	})
}

// GetServer Servers
// @Summary Get server details
// @Description Get detailed server info including config files
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]models.ServerResponse "Server info (including config files)"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /servers/{id} [get]
func GetServer(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	response, err := serverService.GetServer(userID, serverID)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    response,
	})
}

// GetServerRCON ServersRCON
// @Summary Get server RCON info
// @Description Get RCON connection info including password
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]interface{} "RCON info"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /servers/{id}/rcon [get]
func GetServerRCON(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	data, err := serverService.GetServerRCON(userID, serverID)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    data,
	})
}

// GetServerLogs Get server logs
// @Summary Get server logs
// @Description Get Docker container logs for the server
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Param tail query int false "Return last N log lines, default 200"
// @Success 200 {object} map[string]string "Operation successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/logs [get]
func GetServerLogs(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")
	tailStr := c.DefaultQuery("tail", "200")
	tail, err := strconv.Atoi(tailStr)
	if err != nil || tail < 0 {
		tail = 200
	}

	data, err := serverService.GetServerLogs(userID, serverID, tail)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Operation successful",
		"data":    data,
	})
}

// UpdateServer Update Service
// @Summary Update server configuration
// @Description Update server configuration including config files
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Param server body models.ServerUpdateRequest true "Server configuration（）"
// @Success 200 {object} map[string]models.ServerResponse "Success"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id} [put]
func UpdateServer(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	var req models.ServerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request parameters", "")
		return
	}

	if req.GameUserSettings != "" {
		if err := utils.ValidateINIContent(req.GameUserSettings); err != nil {
			utils.BadRequest(c, "GameUserSettings.ini format error", err.Error())
			return
		}
	}
	if req.GameIni != "" {
		if err := utils.ValidateINIContent(req.GameIni); err != nil {
			utils.BadRequest(c, "Game.ini format error", err.Error())
			return
		}
	}

	response, argsChanged, err := serverService.UpdateServer(userID, serverID, req)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		if err.Error() == "Server identifier already exists" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	// 
	message := "Server updated successfully"
	if argsChanged && response.Status == "running" {
		message = "Server updated successfully，Start 。 Servers ， Restart server Start 。"
	}
	middleware.Log.Log(userID, "server.update", fmt.Sprintf("server:%s", serverID), "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message":      message,
		"data":         response,
		"args_changed": argsChanged,
	})
}

// DeleteServer Delete server
// @Summary Delete server
// @Description Delete server config (only stopped servers)
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]string "DeleteSuccess"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id} [delete]
func DeleteServer(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	err := serverService.DeleteServer(userID, serverID)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		if err.Error() == "None DeleteRunning server， Stop server" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Server deleted successfully",
	})
}

// StartServer Start server
// @Summary Start server
// @Description Start the specified ARK server
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]string "StartSuccess"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/start [post]
func StartServer(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	err := serverService.StartServer(userID, serverID)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		if err.Error() == "Servers " || err.Error() == "Servers Start " {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}
	middleware.Log.Log(userID, "server.start", fmt.Sprintf("server:%s", serverID), "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Server start command sent",
	})
}

// StopServer Stop server
// @Summary Stop server
// @Description Stop the specified ARK server
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]string "StopSuccess"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/stop [post]
func StopServer(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	err := serverService.StopServer(userID, serverID)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		if err.Error() == "Servers Stop" || err.Error() == "Servers Stop " {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}
	middleware.Log.Log(userID, "server.stop", fmt.Sprintf("server:%s", serverID), "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Server stop command sent",
	})
}

// RecreateContainer 
// @Summary Rebuild server container
// @Description Rebuild container for server using new image
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]string "Status"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/recreate [post]
func RecreateContainer(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	err := serverService.RecreateContainer(userID, serverID)
	if err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}
	middleware.Log.Log(userID, "server.recreate", fmt.Sprintf("server:%s", serverID), "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Container rebuild started",
	})
}

// RestartServer Restart server
// @Summary Restart server
// @Description Restart the specified ARK server (stop then start)
// @Tags Server Management
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Server ID"
// @Success 200 {object} map[string]string "RestartSuccess"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 500 {object} map[string]string "Server error"
// @Router /servers/{id}/restart [post]
func RestartServer(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	// stop first
	if err := serverService.StopServer(userID, serverID); err != nil {
		if err.Error() == "None Server ID" {
			utils.BadRequest(c, "Invalid request parameters", err.Error())
			return
		}
		if err.Error() == "Server not found" {
			utils.NotFound(c, "Resource not found", err.Error())
			return
		}
		// ignore "already stopped" errors
	}
	// then start
	if err := serverService.StartServer(userID, serverID); err != nil {
		utils.InternalError(c, "Internal server error", err.Error())
		return
	}
	middleware.Log.Log(userID, "server.restart", fmt.Sprintf("server:%s", serverID), "", c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Server restart command sent",
	})
}
