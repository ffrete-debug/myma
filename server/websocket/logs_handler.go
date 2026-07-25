package websocket

import (
	"bufio"
	"net/http"
	"strconv"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var logsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const (
	logsReadTimeout  = 60 * time.Second
	logsWriteTimeout = 10 * time.Second
)

// HandleLogsWebSocket upgrades the connection to a WebSocket and streams
// Docker container logs in real-time. Auth is enforced by AuthMiddleware
// on the route (token via Authorization header or ?token=).
func HandleLogsWebSocket(c *gin.Context) {
	userID := c.GetUint("user_id")
	serverID := c.Param("id")

	conn, err := logsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		safeWriteText(conn, "error: invalid server id")
		return
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		safeWriteText(conn, "error: server not found")
		return
	}

	dm, err := docker_manager.GetDockerManager()
	if err != nil {
		safeWriteText(conn, "error: docker manager unavailable")
		return
	}

	containerName := utils.GetServerContainerName(server.ID)

	reader, err := dm.StreamContainerLogs(containerName, true, "")
	if err != nil {
		safeWriteText(conn, "error: unable to open log stream")
		return
	}
	defer reader.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(logsWriteTimeout))

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			break
		}
		_ = conn.SetWriteDeadline(time.Now().Add(logsWriteTimeout))
	}
}
