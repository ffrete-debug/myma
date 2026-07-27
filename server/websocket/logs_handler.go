package websocket

import (
	"bufio"
	"io"
	"net/http"
	"strconv"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var logsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const (
	logsReadTimeout    = 60 * time.Second
	logsWriteTimeout   = 10 * time.Second
	logsPingPeriod     = (logsReadTimeout * 9) / 10
	logsMaxMessageSize = 512
	// ARK writes very long single lines (mod lists, stack traces). The Scanner
	// default of 64 KB would silently end the stream on the first one.
	logsMaxLineSize = 1 << 20
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

	// Non-TTY containers multiplex stdout/stderr with an 8-byte frame header per
	// message; without demuxing, every line reaches the browser prefixed with
	// binary garbage. GetContainerLogs does the same via stdcopy.StdCopy — here
	// it feeds a pipe so the stream keeps flowing line by line.
	pr, pw := io.Pipe()
	defer pr.Close()
	go func() {
		_, err := stdcopy.StdCopy(pw, pw, reader)
		_ = pw.CloseWithError(err)
	}()

	// The client never sends anything, but an active reader is still required:
	// it processes pongs and notices a disconnect straight away, instead of
	// leaving this handler and the Docker API connection blocked forever on a
	// quiet container.
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn.SetReadLimit(logsMaxMessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(logsReadTimeout))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(logsReadTimeout))
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	lines := make(chan []byte, 256)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 0, 64*1024), logsMaxLineSize)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			// scanner.Bytes() is only valid until the next Scan, so hand the
			// writer its own copy.
			line := make([]byte, len(scanner.Bytes()))
			copy(line, scanner.Bytes())
			select {
			case lines <- line:
			case <-done:
				return
			}
		}
		if err := scanner.Err(); err != nil {
			utils.Warn("log stream ended with error", zap.String("container", containerName), zap.Error(err))
		}
	}()

	// Only this goroutine writes to the connection. Pings keep the peer's read
	// deadline refreshed while a container is silent.
	ticker := time.NewTicker(logsPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return
			}
			_ = conn.SetWriteDeadline(time.Now().Add(logsWriteTimeout))
			if err := conn.WriteMessage(websocket.TextMessage, line); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(logsWriteTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
