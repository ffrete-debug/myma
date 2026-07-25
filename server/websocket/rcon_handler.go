package websocket

import (
	"net/http"
	"strconv"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/rcon"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var rconUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Limits applied to the RCON WebSocket session.
const (
	rconMaxCommandSize = 1 << 10          // 1 KB per inbound command (RCON commands are short shell-like strings)
	rconReadTimeout    = 30 * time.Second  // idle read deadline before we drop the session
	rconWriteTimeout   = 10 * time.Second  // write deadline for responses
	rconMinInterval    = 250 * time.Millisecond // rate limit between commands per session
	rconMaxOutput      = 65536             // 64 KB — bounce brutal responses (server logs, dump commands)
)

// HandleRCONWebSocket upgrades the connection to a WebSocket and bridges
// inbound RCON commands to the ARK server's RCON endpoint. Auth is enforced
// by AuthMiddleware on the route (token via Authorization header or ?token=).
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
		safeWriteText(conn, "error: invalid server id")
		return
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		safeWriteText(conn, "error: server not found")
		return
	}

	// Per-connection rate limiter: a session that issues commands too fast is
	// most likely a runaway script. Forcing a small floor (~250 ms) keeps the
	// terminal responsive without hammering the RCON socket on the game server.
	lastCommand := time.Time{}

	for {
		// Refresh the read deadline every iteration — a single endless command
		// should not keep the session dead.
		if err := conn.SetReadDeadline(time.Now().Add(rconReadTimeout)); err != nil {
			break
		}

		// Apply a write deadline so a stalled client can't accumulate buffered
		// responses on the server side.
		_ = conn.SetWriteDeadline(time.Now().Add(rconWriteTimeout))

		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if len(msg) == 0 {
			continue
		}

		if len(msg) > rconMaxCommandSize {
			safeWriteText(conn, "error: command too long (max 1KB)")
			continue
		}

		if now := time.Now(); now.Sub(lastCommand) < rconMinInterval {
			// Stripped-down backpressure: respond with a flood-guard notice.
			// Throwing in a short sleep would block the reader loop, so instead
			// we bounce the command and let the client decide.
			safeWriteText(conn, "error: rate limited — slow down")
			continue
		}

		lastCommand = time.Now()

		output, err := rcon.ExecuteCommand("localhost", server.RCONPort, server.AdminPassword, string(msg))
		if err != nil {
			safeWriteText(conn, "error: "+err.Error())
			continue
		}

		if len(output) > rconMaxOutput {
			output = output[:rconMaxOutput] + "\n... output truncated ..."
		}

		if err := conn.WriteMessage(websocket.TextMessage, []byte(output)); err != nil {
			break
		}
	}
}

// safeWriteText is a best-effort text write used for control/protocol errors.
// The WS connection may already be closing when these run, so failures are
// swallowed.
func safeWriteText(conn *websocket.Conn, msg string) {
	_ = conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
