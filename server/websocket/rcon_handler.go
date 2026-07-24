package websocket

import (
	"net/http"
	"strconv"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/rcon"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var rconUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func HandleRCONWebSocket(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	conn, err := rconUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("invalid server id"))
		return
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("server not found"))
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		output, err := rcon.ExecuteCommand("localhost", server.RCONPort, server.AdminPassword, string(msg))
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("error: " + err.Error()))
			continue
		}

		if len(output) > 65536 {
			output = output[:65536] + "\n... output truncated ..."
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(output))
		if err != nil {
			break
		}
	}
}
