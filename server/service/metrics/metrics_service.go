// Package metrics exposes per-server resource and population figures for the
// dashboard: CPU and memory from the Docker daemon, player count over RCON.
package metrics

import (
	"fmt"
	"sync"
	"time"

	"ark-server-commander/config"
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/service/player"
	"ark-server-commander/service/rcon"
	"ark-server-commander/utils"

	"go.uber.org/zap"
)

// ServerMetrics is one dashboard row: what the container is consuming and how
// many players are on it.
//
// PlayersOnline is -1 when the count could not be determined (server stopped,
// RCON unreachable, bad admin password). That is deliberately distinct from 0,
// which means the server answered and nobody is connected — the dashboard must
// not render an unreachable server as empty.
type ServerMetrics struct {
	ServerID      uint    `json:"server_id"`
	Identifier    string  `json:"server_identifier"`
	SessionName   string  `json:"session_name"`
	Status        string  `json:"status"`
	CPUPercent    float64 `json:"cpu_percent"`
	CPUCores      int     `json:"cpu_cores"`
	MemoryUsageMB float64 `json:"memory_usage_mb"`
	MemoryLimitMB float64 `json:"memory_limit_mb"`
	MemoryPercent float64 `json:"memory_percent"`
	NetworkRxMB   float64 `json:"network_rx_mb"`
	NetworkTxMB   float64 `json:"network_tx_mb"`
	PlayersOnline int     `json:"players_online"`
	MaxPlayers    int     `json:"max_players"`
	SampledAt     int64   `json:"sampled_at"`
	Error         string  `json:"error,omitempty"`
}

// PlayersUnknown marks a player count that could not be read.
const PlayersUnknown = -1

// rconTimeout bounds a single player-count probe. RCON on a loaded ARK server
// can be slow; anything past this is treated as unreachable so one bad server
// cannot stall the whole dashboard.
const rconTimeout = 5 * time.Second

type Service struct{}

func NewService() *Service { return &Service{} }

// GetServerMetrics samples one server the caller owns.
func (s *Service) GetServerMetrics(serverID, userID uint) (*ServerMetrics, error) {
	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		return nil, fmt.Errorf("server not found")
	}
	return s.sample(&server), nil
}

// GetAllServerMetrics samples every server the caller owns.
//
// Servers are sampled concurrently: each one costs a Docker round-trip plus an
// RCON round-trip, so doing this serially would make the dashboard latency the
// sum of every server's worst case. Concurrency is bounded by the number of
// servers the user owns, which is small and already capped by the list handler.
func (s *Service) GetAllServerMetrics(userID uint) ([]ServerMetrics, error) {
	var servers []models.Server
	if err := database.DB.Where("user_id = ?", userID).Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}

	out := make([]ServerMetrics, len(servers))
	var wg sync.WaitGroup
	for i := range servers {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			out[idx] = *s.sample(&servers[idx])
		}(i)
	}
	wg.Wait()

	return out, nil
}

// sample never returns an error: a server that cannot be reached still belongs
// on the dashboard, annotated with why. Failing the whole request because one
// container is down would make the page useless exactly when it matters.
func (s *Service) sample(server *models.Server) *ServerMetrics {
	m := &ServerMetrics{
		ServerID:      server.ID,
		Identifier:    server.Identifier,
		SessionName:   server.SessionName,
		MaxPlayers:    server.MaxPlayers,
		PlayersOnline: PlayersUnknown,
		SampledAt:     time.Now().Unix(),
	}

	dm, err := docker_manager.GetDockerManager()
	if err != nil || dm == nil {
		m.Status = "unknown"
		m.Error = "docker unavailable"
		return m
	}

	containerName := utils.GetServerContainerName(server.ID)

	status, err := dm.GetContainerStatus(containerName)
	if err != nil {
		m.Status = "stopped"
		m.Error = "container not found"
		return m
	}
	m.Status = status

	if status != "running" {
		// A stopped container has no stats and no RCON. Reporting zeroes would
		// be indistinguishable from an idle running server.
		return m
	}

	if stats, err := dm.GetContainerStats(containerName); err != nil {
		utils.Warn("container stats failed", zap.Uint("server_id", server.ID), zap.Error(err))
		m.Error = "stats unavailable"
	} else {
		m.CPUPercent = stats.CPUPercent
		m.CPUCores = stats.CPUCores
		m.MemoryUsageMB = stats.MemoryUsageMB
		m.MemoryLimitMB = stats.MemoryLimitMB
		m.MemoryPercent = stats.MemoryPercent
		m.NetworkRxMB = stats.NetworkRxMB
		m.NetworkTxMB = stats.NetworkTxMB
	}

	if n, err := playerCount(server); err != nil {
		utils.Warn("player count failed", zap.Uint("server_id", server.ID), zap.Error(err))
	} else {
		m.PlayersOnline = n
	}

	return m
}

// playerCount asks the game server directly over RCON.
//
// The call is run in a goroutine behind a timeout because the RCON client's own
// deadline covers the socket, not the whole exchange, and a half-open game
// server can otherwise hold the sample open far longer than the dashboard's
// refresh interval.
func playerCount(server *models.Server) (int, error) {
	type result struct {
		out string
		err error
	}
	ch := make(chan result, 1) // buffered: a late reply must not leak the goroutine

	go func() {
		out, err := rcon.ExecuteCommand(config.RCONHost, server.RCONPort, server.AdminPassword, "listplayers")
		ch <- result{out, err}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			return 0, r.err
		}
		return len(player.ParseListPlayersOutput(r.out)), nil
	case <-time.After(rconTimeout):
		return 0, fmt.Errorf("rcon timeout after %s", rconTimeout)
	}
}
