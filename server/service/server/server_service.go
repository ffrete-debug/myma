package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"ark-server-commander/config"
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/service/rcon"
	"ark-server-commander/utils"
	"ark-server-commander/websocket"

	"go.uber.org/zap"
)

const (
	// Pagination bounds, same convention as controllers/audit.
	defaultPageLimit = 20
	maxPageLimit     = 100

	// statusRefreshBudget bounds the Docker daemon work done while building a
	// server list; past the budget the stored status is used instead.
	statusRefreshBudget = 5 * time.Second
)

// ServerService Server Management
type ServerService struct {
	userMutexes sync.Map // map[uint]*sync.Mutex por userID
}

// NewServerService CreateServer Service
func NewServerService() *ServerService {
	return &ServerService{}
}

// broadcastStatus sends a WebSocket status update for this server.
func (s *ServerService) broadcastStatus(serverID uint, status string) {
	if hub := websocket.GetGlobalHub(); hub != nil {
		hub.BroadcastToServer(serverID, map[string]interface{}{
			"status": status,
		})
	}
}

// getUserMutex User，UserServers
func (s *ServerService) getUserMutex(userID uint) *sync.Mutex {
	v, _ := s.userMutexes.LoadOrStore(userID, &sync.Mutex{})
	mu, _ := v.(*sync.Mutex)
	return mu
}

// checkPortConflict
// userID: UserID
// serverID: Server ID（0 Servers，Server ID）
// port, queryPort, rconPort:
// : Error
func (s *ServerService) checkPortConflict(userID uint, serverID uint, port, queryPort, rconPort int) error {
	var existingServers []models.Server
	query := database.DB.Where("user_id = ?", userID)
	if serverID > 0 {
		query = query.Where("id != ?", serverID)
	}
	if err := query.Find(&existingServers).Error; err != nil {
		return fmt.Errorf(" : %w", err)
	}

	for _, existingServer := range existingServers {
		if existingServer.Port == port {
			return fmt.Errorf(" ：  %d  Servers %s  ", port, existingServer.SessionName)
		}
		if existingServer.QueryPort == queryPort {
			return fmt.Errorf(" ：Query Port %d  Servers %s  ", queryPort, existingServer.SessionName)
		}
		if existingServer.RCONPort == rconPort {
			return fmt.Errorf(" ：RCON Port %d  Servers %s  ", rconPort, existingServer.SessionName)
		}
	}
	return nil
}

// GetServers UserServers
func (s *ServerService) GetServers(userID uint, page, limit int) ([]models.ServerResponse, int64, error) {
	var servers []models.Server
	var total int64

	// Clamp before computing the offset: limit is part of the offset expression, so
	// clamping afterwards leaves every page pointing at offset 0.
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > maxPageLimit {
		limit = defaultPageLimit
	}
	offset := (page - 1) * limit

	if err := database.DB.Model(&models.Server{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("Count server list : %w", err)
	}

	if err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Offset(offset).Find(&servers).Error; err != nil {
		return nil, 0, fmt.Errorf("Get server list : %w", err)
	}

	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, 0, fmt.Errorf(" Docker Manager : %w", err)
	}

	// The Docker calls below are one blocking round-trip per server and docker_manager
	// owns a context.Background() that cannot be cancelled, so bound the whole refresh
	// with a deadline. TODO: to honour the caller's request context instead, GetServers
	// would need a ctx parameter (controllers/servers.GetServers passing c.Request.Context())
	// and docker_manager would need per-call contexts.
	ctx, cancel := context.WithTimeout(context.Background(), statusRefreshBudget)
	defer cancel()

	// Rows whose Docker status drifted from the DB, grouped by the new status so they
	// can be written back in one statement per status instead of a goroutine per row.
	drift := make(map[string][]uint)

	var serverResponses []models.ServerResponse
	for _, server := range servers {
		realTimeStatus := server.Status

		// GetContainerStatus already reports "not_found", so the extra ContainerExists
		// probe is redundant: one daemon round-trip per server is enough.
		if ctx.Err() == nil {
			containerName := utils.GetServerContainerName(server.ID)
			dockerStatus, err := dockerManager.GetContainerStatus(containerName)
			switch {
			case err != nil:
				utils.Warn("Get container Status ", zap.Uint("server_id", server.ID), zap.Error(err))
			case dockerStatus == "not_found":
				// Container not foundStatusYes，StopStatus
				if server.Status == "running" {
					realTimeStatus = "stopped"
				}
			default:
				realTimeStatus = dockerStatus
			}

			if realTimeStatus != server.Status {
				drift[realTimeStatus] = append(drift[realTimeStatus], server.ID)
			}
		}

		serverResponses = append(serverResponses, models.ServerResponse{
			ID:          server.ID,
			Identifier:  server.Identifier,
			SessionName: server.SessionName,
			ClusterID:   server.ClusterID,
			Port:        server.Port,
			QueryPort:   server.QueryPort,
			RCONPort:    server.RCONPort,
			Map:         server.Map,
			MaxPlayers:  server.MaxPlayers,
			GameModIds:  server.GameModIds,
			Status:      realTimeStatus,
			AutoRestart: server.AutoRestart,
			UserID:      server.UserID,
			CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   server.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	s.persistStatusDrift(userID, drift)

	return serverResponses, total, nil
}

// persistStatusDrift writes back the statuses observed on Docker, one batched update
// per status value, then broadcasts them. Failures are logged rather than returned: a
// stale status row must not fail the list request.
func (s *ServerService) persistStatusDrift(userID uint, drift map[string][]uint) {
	for status, ids := range drift {
		if err := database.DB.Model(&models.Server{}).
			Where("user_id = ? AND id IN ?", userID, ids).
			Update("status", status).Error; err != nil {
			utils.Error("Update Service Status ", zap.String("status", status), zap.Uints("server_ids", ids), zap.Error(err))
			continue
		}

		for _, id := range ids {
			s.broadcastStatus(id, status)
		}
	}
}

// CreateServer Create a new server
func (s *ServerService) CreateServer(userID uint, req models.ServerRequest) (*models.ServerResponse, error) {
	mu := s.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// ServersYesNo
	var existingServer models.Server
	if err := database.DB.Where("identifier = ? AND user_id = ?", req.Identifier, userID).First(&existingServer).Error; err == nil {
		return nil, fmt.Errorf("Server identifier already exists")
	}

	// Settings
	if req.Map == "" {
		req.Map = "TheIsland"
	}
	if req.MaxPlayers == 0 {
		req.MaxPlayers = 70
	}
	if req.AutoRestart == nil {
		defaultVal := true
		req.AutoRestart = &defaultVal
	}

	//
	if err := s.checkPortConflict(userID, 0, req.Port, req.QueryPort, req.RCONPort); err != nil {
		return nil, err
	}

	// On
	tx := database.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf(" Start : %w", tx.Error)
	}

	// CreateServers
	server := models.Server{
		Identifier:    req.Identifier,
		SessionName:   req.SessionName,
		ClusterID:     req.ClusterID,
		Port:          req.Port,
		QueryPort:     req.QueryPort,
		RCONPort:      req.RCONPort,
		AdminPassword: req.AdminPassword,
		Map:           req.Map,
		MaxPlayers:    req.MaxPlayers,
		GameModIds:    req.GameModIds,
		Status:        "stopped",
		AutoRestart:   *req.AutoRestart,
		UserID:        userID,
	}

	if req.ServerArgs != nil {
		argsJson, err := json.Marshal(req.ServerArgs)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("Start Error: %w", err)
		}
		server.ServerArgsJSON = string(argsJson)
	} else {
		server.ServerArgsJSON = "{}"
	}

	if err := tx.Create(&server).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("ServersCreate : %w", err)
	}

	// CreateDocker
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf(" Docker Manager : %w", err)
	}

	_, err = dockerManager.CreateVolume(server.ID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("CreateDocker : %w", err)
	}

	//
	var gameUserSettings string
	var gameIni string

	if req.GameUserSettings != "" {
		if err = utils.ValidateINIContent(req.GameUserSettings); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("GameUserSettings.ini Error: %w", err)
		}
		gameUserSettings = req.GameUserSettings
	} else {
		// Seed from the session name and the configured player cap, not the
		// identifier and a hardcoded 70. SessionName is what the Steam browser
		// shows and it is not passed on the command line (the image expands
		// SERVER_ARGS unquoted, so a space would truncate it), which makes this
		// file the only place it comes from.
		gameUserSettings = utils.GetDefaultGameUserSettings(server.SessionName, server.Map, server.MaxPlayers)
	}

	if req.GameIni != "" {
		if err = utils.ValidateINIContent(req.GameIni); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("game.ini Error: %w", err)
		}
		gameIni = req.GameIni
	} else {
		gameIni = utils.GetDefaultGameIni()
	}

	//
	if err := dockerManager.WriteConfigFile(server.ID, utils.GameUserSettingsFileName, gameUserSettings); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf(" GameUserSettings.ini : %w", err)
	}

	if err := dockerManager.WriteConfigFile(server.ID, utils.GameIniFileName, gameIni); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf(" Game.ini : %w", err)
	}

	//
	if err := tx.Commit().Error; err != nil {
		_ = dockerManager.RemoveVolume(server.ID)
		return nil, fmt.Errorf(" : %w", err)
	}

	//
	response := models.ServerResponse{
		ID:          server.ID,
		Identifier:  server.Identifier,
		SessionName: server.SessionName,
		ClusterID:   server.ClusterID,
		Port:        server.Port,
		QueryPort:   server.QueryPort,
		RCONPort:    server.RCONPort,
		Map:         server.Map,
		GameModIds:  server.GameModIds,
		Status:      server.Status,
		AutoRestart: server.AutoRestart,
		UserID:      server.UserID,
		CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   server.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	//
	if gameUserSettings, err := dockerManager.ReadConfigFile(uint(server.ID), utils.GameUserSettingsFileName); err == nil {
		response.GameUserSettings = gameUserSettings
	}
	if gameIni, err := dockerManager.ReadConfigFile(uint(server.ID), utils.GameIniFileName); err == nil {
		response.GameIni = gameIni
	}

	return &response, nil
}

// GetServer Servers
func (s *ServerService) GetServer(userID uint, serverID string) (*models.ServerResponse, error) {
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err = database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return nil, fmt.Errorf("Server not found")
	}

	// Start
	var serverArgs *models.ServerArgs
	if server.ServerArgsJSON != "" && server.ServerArgsJSON != "{}" {
		serverArgs = models.NewServerArgs()
		if err = json.Unmarshal([]byte(server.ServerArgsJSON), serverArgs); err != nil {
			serverArgs = models.FromServer(server)
		}
	} else {
		serverArgs = models.FromServer(server)
	}

	response := models.ServerResponse{
		ID:            server.ID,
		Identifier:    server.Identifier,
		SessionName:   server.SessionName,
		ClusterID:     server.ClusterID,
		Port:          server.Port,
		QueryPort:     server.QueryPort,
		RCONPort:      server.RCONPort,
		Map:           server.Map,
		MaxPlayers:    server.MaxPlayers,
		GameModIds:    server.GameModIds,
		Status:        server.Status,
		AutoRestart:   server.AutoRestart,
		UserID:        server.UserID,
		CreatedAt:     server.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     server.UpdatedAt.Format("2006-01-02 15:04:05"),
		ServerArgs:    serverArgs,
		GeneratedArgs: serverArgs.GenerateArgsString(server),
	}

	//
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, fmt.Errorf(" Docker Manager : %w", err)
	}

	if gameUserSettings, err := dockerManager.ReadConfigFile(uint(id), utils.GameUserSettingsFileName); err == nil {
		response.GameUserSettings = gameUserSettings
	}
	if gameIni, err := dockerManager.ReadConfigFile(uint(id), utils.GameIniFileName); err == nil {
		response.GameIni = gameIni
	}

	return &response, nil
}

// GetServerRCON ServersRCON
// GetServerLogs Get server logs
func (s *ServerService) GetServerLogs(userID uint, serverID string, tail int) (string, error) {
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return "", fmt.Errorf("Server not found")
	}

	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return "", fmt.Errorf(" Docker Manager : %w", err)
	}

	containerName := utils.GetServerContainerName(server.ID)
	logs, err := dockerManager.GetContainerLogs(containerName, tail)
	if err != nil {
		return "", fmt.Errorf(" : %w", err)
	}
	return logs, nil
}

func (s *ServerService) GetServerRCON(userID uint, serverID string) (map[string]interface{}, error) {
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return nil, fmt.Errorf("Server not found")
	}

	return map[string]interface{}{
		"server_id":         server.ID,
		"server_identifier": server.Identifier,
		"rcon_port":         server.RCONPort,
		"admin_password":    server.AdminPassword,
	}, nil
}

// ExecuteRCONCommand executes an RCON command on the server
func (s *ServerService) ExecuteRCONCommand(userID uint, serverID string, command string) (string, error) {
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return "", fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return "", fmt.Errorf("Server not found")
	}

	output, err := rcon.ExecuteCommand(config.RCONHost, server.RCONPort, server.AdminPassword, command)
	if err != nil {
		return "", fmt.Errorf("RCON command failed: %w", err)
	}

	return output, nil
}

// UpdateServer Update Service
func (s *ServerService) UpdateServer(userID uint, serverID string, req models.ServerUpdateRequest) (*models.ServerResponse, bool, error) {
	mu := s.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// serverIDuint
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return nil, false, fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err = database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return nil, false, fmt.Errorf("Server not found")
	}

	// YesNo
	if req.Identifier != "" && req.Identifier != server.Identifier {
		var existingServer models.Server
		if err = database.DB.Where("identifier = ? AND user_id = ? AND id != ?", req.Identifier, userID, id).First(&existingServer).Error; err == nil {
			return nil, false, fmt.Errorf("Server identifier already exists")
		}
		server.Identifier = req.Identifier
	}

	//
	if req.SessionName != "" {
		server.SessionName = req.SessionName
	}
	if req.ClusterID != "" {
		server.ClusterID = req.ClusterID
	}
	if req.Port > 0 {
		server.Port = req.Port
	}
	if req.QueryPort > 0 {
		server.QueryPort = req.QueryPort
	}
	if req.RCONPort > 0 {
		server.RCONPort = req.RCONPort
	}
	if req.AdminPassword != "" {
		server.AdminPassword = req.AdminPassword
	}
	if req.Map != "" {
		server.Map = req.Map
	}
	if req.MaxPlayers > 0 {
		server.MaxPlayers = req.MaxPlayers
	}
	if req.GameModIds != "" {
		server.GameModIds = req.GameModIds
	}

	// StartYesNo
	argsChanged := false
	// Config edits also require a restart to take effect, and the UI needs to be
	// told so - previously only ServerArgs changes set the flag.
	configChanged := false
	if req.ServerArgs != nil {
		argsJson, err := json.Marshal(req.ServerArgs)
		if err != nil {
			return nil, false, fmt.Errorf("Start Error: %w", err)
		}
		newArgsJSON := string(argsJson)
		if server.ServerArgsJSON != newArgsJSON {
			argsChanged = true
			server.ServerArgsJSON = newArgsJSON
		}
	}

	//
	if err := s.checkPortConflict(userID, uint(id), server.Port, server.QueryPort, server.RCONPort); err != nil {
		return nil, false, err
	}

	if err := database.DB.Save(&server).Error; err != nil {
		return nil, false, fmt.Errorf("Servers : %w", err)
	}

	//
	if req.GameUserSettings != "" || req.GameIni != "" {
		dockerManager, err := docker_manager.GetDockerManager()
		if err != nil {
			return nil, false, fmt.Errorf(" Docker Manager : %w", err)
		}

		if req.GameUserSettings != "" {
			if err := utils.ValidateINIContent(req.GameUserSettings); err != nil {
				return nil, false, fmt.Errorf("GameUserSettings.ini Error: %w", err)
			}
			if err := dockerManager.WriteConfigFile(uint(id), utils.GameUserSettingsFileName, req.GameUserSettings); err != nil {
				return nil, false, fmt.Errorf(" GameUserSettings.ini : %w", err)
			}
			// Remember the intent. ARK overwrites this file on every clean
			// shutdown, so the copy in the volume is not durable; the restart
			// path re-applies this after the shutdown.
			server.DesiredGameUserSettings = req.GameUserSettings
			configChanged = true
		}

		if req.GameIni != "" {
			if err := utils.ValidateINIContent(req.GameIni); err != nil {
				return nil, false, fmt.Errorf("Game.ini Error: %w", err)
			}
			if err := dockerManager.WriteConfigFile(uint(id), utils.GameIniFileName, req.GameIni); err != nil {
				return nil, false, fmt.Errorf(" Game.ini : %w", err)
			}
			server.DesiredGameIni = req.GameIni
			configChanged = true
		}
	}

	//
	response := models.ServerResponse{
		ID:          server.ID,
		Identifier:  server.Identifier,
		SessionName: server.SessionName,
		ClusterID:   server.ClusterID,
		Port:        server.Port,
		QueryPort:   server.QueryPort,
		RCONPort:    server.RCONPort,
		Map:         server.Map,
		MaxPlayers:  server.MaxPlayers,
		GameModIds:  server.GameModIds,
		Status:      server.Status,
		AutoRestart: server.AutoRestart,
		UserID:      server.UserID,
		CreatedAt:   server.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   server.UpdatedAt.Format("2006-01-02 15:04:05"),
		ServerArgs:  models.FromServer(server),
	}

	//
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, false, fmt.Errorf(" Docker Manager : %w", err)
	}

	if gameUserSettings, err := dockerManager.ReadConfigFile(uint(id), utils.GameUserSettingsFileName); err == nil {
		response.GameUserSettings = gameUserSettings
	}
	if gameIni, err := dockerManager.ReadConfigFile(uint(id), utils.GameIniFileName); err == nil {
		response.GameIni = gameIni
	}

	return &response, argsChanged || configChanged, nil
}

// DeleteServer Delete server
func (s *ServerService) DeleteServer(userID uint, serverID string) error {
	mu := s.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// serverIDuint
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return fmt.Errorf("Server not found")
	}

	if server.Status == "running" {
		return fmt.Errorf("None DeleteRunning server， Stop server")
	}

	// Delete server
	if err := database.DB.Delete(&server).Error; err != nil {
		return fmt.Errorf("ServersDelete : %w", err)
	}

	// DeleteDocker
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return fmt.Errorf(" Docker Manager : %w", err)
	}

	containerName := utils.GetServerContainerName(server.ID)
	containerExists, err := dockerManager.ContainerExists(containerName)
	if err != nil {
		utils.Warn(" ", zap.Error(err))
	} else if containerExists {
		if err := dockerManager.RemoveContainer(containerName); err != nil {
			utils.Warn("DeleteDocker ", zap.Error(err))
		}
	}

	// Remove the data and plugins volumes.
	//
	// Previously they were left behind on every delete, so each removed server
	// orphaned its volumes permanently. Worse, CreateSingleVolume adopts an
	// existing volume by name instead of failing, so if a server id was ever
	// reused - which happens when the bind-mounted ./data database is reset
	// while the Docker volumes survive - the new server silently inherited the
	// old server's plugins.
	//
	// Failures are warnings, not errors: the DB row and container are already
	// gone, so returning an error here would report a failed delete for a
	// server that no longer exists.
	if err := dockerManager.RemoveVolume(server.ID); err != nil {
		utils.Warn("could not remove server volumes; they may need manual cleanup",
			zap.Uint("server_id", server.ID), zap.Error(err))
	}

	// Drop rows that reference the server. They are not reachable through any
	// API once the server is gone, so leaving them would only accumulate.
	for _, related := range []interface{}{
		&models.ServerMod{}, &models.BackupSchedule{}, &models.Player{},
	} {
		if err := database.DB.Where("server_id = ?", server.ID).Delete(related).Error; err != nil {
			utils.Warn("could not clean up related rows",
				zap.Uint("server_id", server.ID), zap.Error(err))
		}
	}

	return nil
}

// StartServer Start server
func (s *ServerService) StartServer(userID uint, serverID string) error {
	mu := s.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// serverIDuint
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return fmt.Errorf("Server not found")
	}

	if server.Status == "running" {
		return fmt.Errorf("Servers ")
	}

	if server.Status == "starting" {
		return fmt.Errorf("Servers Start ")
	}

	// Update ServiceStatusStart
	server.Status = "starting"
	if err := database.DB.Save(&server).Error; err != nil {
		return fmt.Errorf("Update Service Status : %w", err)
	}

	// StartDocker
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return fmt.Errorf(" Docker Manager : %w", err)
	}

	containerName := utils.GetServerContainerName(server.ID)

	go func() {
		if err := s.startServerAsync(server, dockerManager, containerName); err != nil {
			utils.Error("Start server ", zap.Error(err))
			database.DB.Model(&server).Update("status", "stopped")
			s.broadcastStatus(server.ID, "stopped")
		}
	}()

	return nil
}

// startServerAsync Start server
func (s *ServerService) startServerAsync(server models.Server, dockerManager *docker_manager.DockerManager, containerName string) error {
	// YesNo
	missingImages, err := dockerManager.ValidateRequiredImages()
	if err != nil {
		return fmt.Errorf(" : %w", err)
	}
	if len(missingImages) > 0 {
		return fmt.Errorf("None Start server， : %v。 Start server", missingImages)
	}

	// YesNo
	containerExists, err := dockerManager.ContainerExists(containerName)
	if err != nil {
		return fmt.Errorf(" YesNo : %w", err)
	}

	needRecreateContainer := false

	if containerExists {
		// YesNo
		envVars, err := dockerManager.GetContainerEnvVars(containerName)
		if err != nil {
			needRecreateContainer = true
		} else {
			// ServersStart
			var serverArgs *models.ServerArgs
			if server.ServerArgsJSON != "" && server.ServerArgsJSON != "{}" {
				serverArgs = models.NewServerArgs()
				if err := json.Unmarshal([]byte(server.ServerArgsJSON), serverArgs); err != nil {
					serverArgs = models.FromServer(server)
				}
			} else {
				serverArgs = models.FromServer(server)
			}
			currentArgsString := serverArgs.GenerateArgsString(server)

			//
			if containerArgsString, exists := envVars["SERVER_ARGS"]; exists {
				if containerArgsString != currentArgsString {
					needRecreateContainer = true
				}
			} else {
				needRecreateContainer = true
			}

			//
			if !needRecreateContainer {
				if server.GameModIds != envVars["GameModIds"] {
					needRecreateContainer = true
				}
			}
		}

		if needRecreateContainer {
			if err := dockerManager.RemoveContainer(containerName); err != nil {
				return fmt.Errorf("Delete : %w", err)
			}
		}
	}

	// Create
	if !containerExists || needRecreateContainer {
		_, err = dockerManager.CreateContainer(server.ID, server.Identifier, server.Port, server.QueryPort, server.RCONPort, server.AdminPassword, server.Map, server.GameModIds, server.AutoRestart)
		if err != nil {
			return fmt.Errorf("Create : %w", err)
		}
	}

	// Start
	if err := dockerManager.StartContainer(containerName); err != nil {
		return fmt.Errorf("Start : %w", err)
	}

	// Start
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		status, err := dockerManager.GetContainerStatus(containerName)
		if err != nil {
			continue
		}

		if status == "running" {
			if err := database.DB.Model(&server).Update("status", "running").Error; err != nil {
				utils.Error("Update Service Status running ", zap.Error(err))
			}
			s.broadcastStatus(server.ID, "running")
			return nil
		}
	}

	return fmt.Errorf(" Start ")
}

// StopServer Stop server
func (s *ServerService) StopServer(userID uint, serverID string) error {
	mu := s.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// serverIDuint
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return fmt.Errorf("Server not found")
	}

	if server.Status == "stopped" {
		return fmt.Errorf("Servers Stop")
	}

	if server.Status == "stopping" {
		return fmt.Errorf("Servers Stop ")
	}

	// Update ServiceStatusStop
	server.Status = "stopping"
	if err := database.DB.Save(&server).Error; err != nil {
		return fmt.Errorf("Update Service Status : %w", err)
	}

	// StopDocker
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return fmt.Errorf(" Docker Manager : %w", err)
	}

	containerName := utils.GetServerContainerName(server.ID)

	go func() {
		s.stopServerAsync(server, dockerManager, containerName)
	}()

	return nil
}

// stopServerAsync Stop server
func (s *ServerService) stopServerAsync(server models.Server, dockerManager *docker_manager.DockerManager, containerName string) {
	// YesNo
	containerExists, err := dockerManager.ContainerExists(containerName)
	if err != nil {
		utils.Error(" ", zap.Error(err))
		database.DB.Model(&server).Update("status", "stopped")
		s.broadcastStatus(server.ID, "stopped")
		return
	}

	if !containerExists {
		database.DB.Model(&server).Update("status", "stopped")
		s.broadcastStatus(server.ID, "stopped")
		return
	}

	// Stop
	if err := dockerManager.StopContainer(containerName); err != nil {
		utils.Error("StopDocker ", zap.Error(err))
	}

	// Status
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		status, err := dockerManager.GetContainerStatus(containerName)
		if err != nil {
			continue
		}

		if status == "stopped" || status == "not_found" {
			break
		}
	}

	// Update statusStop
	if err := database.DB.Model(&server).Update("status", "stopped").Error; err != nil {
		utils.Error("Update Service Status stopped ", zap.Error(err))
	}
	s.broadcastStatus(server.ID, "stopped")
}

// ValidateRequiredImages Start serverYesNo
func (s *ServerService) ValidateRequiredImages() (missing []string, err error) {
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, fmt.Errorf(" Docker Manager : %w", err)
	}

	return dockerManager.ValidateRequiredImages()
}

// Check ImageUpdates
func (s *ServerService) CheckImageUpdates() (map[string]bool, error) {
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, fmt.Errorf(" Docker Manager : %w", err)
	}

	requiredImages := []string{
		"tbro98/ase-server:latest",
		"alpine:latest",
	}

	updateStatus := make(map[string]bool)
	for _, imageName := range requiredImages {
		hasUpdate, err := dockerManager.CheckImageUpdate(imageName)
		if err != nil {
			// ，
			updateStatus[imageName] = false
		} else {
			updateStatus[imageName] = hasUpdate
		}
	}

	return updateStatus, nil
}

// PullImage
func (s *ServerService) PullImage(imageName string) error {
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return fmt.Errorf(" Docker Manager : %w", err)
	}

	// Image nameYesNo
	allowedImages := []string{
		"tbro98/ase-server:latest",
		"alpine:latest",
	}

	allowed := false
	for _, allowedImage := range allowedImages {
		if imageName == allowedImage {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf(" : %s", imageName)
	}

	//
	go func() {
		if err := dockerManager.PullImageWithProgress(imageName); err != nil {
			utils.Error(" ", zap.String("image", imageName), zap.Error(err))
		} else {
			utils.Info(" ", zap.String("image", imageName))
		}
	}()

	return nil
}

// UpdateImage Off
func (s *ServerService) UpdateImage(imageName string, userID uint) ([]models.ServerResponse, error) {
	// Image name
	allowedImages := []string{
		"tbro98/ase-server:latest",
		"alpine:latest",
	}

	allowed := false
	for _, allowedImage := range allowedImages {
		if imageName == allowedImage {
			allowed = true
			break
		}
	}

	if !allowed {
		return nil, fmt.Errorf(" : %s", imageName)
	}

	// Servers
	affectedServers, err := s.GetAffectedServers(imageName, userID)
	if err != nil {
		return nil, fmt.Errorf(" Servers : %w", err)
	}

	//
	go func() {
		dockerManager, err := docker_manager.GetDockerManager()
		if err != nil {
			utils.Error(" Docker Manager ", zap.Error(err))
			return
		}

		//
		utils.Info("On ", zap.String("image", imageName))
		if err := dockerManager.PullImageWithProgress(imageName); err != nil {
			utils.Error(" ", zap.String("image", imageName), zap.Error(err))
			return
		}

		utils.Info("Image update complete", zap.String("image", imageName))

		// ，UserImage update complete
		// User
	}()

	return affectedServers, nil
}

// GetAffectedServers returns servers using the given image
func (s *ServerService) GetAffectedServers(imageName string, userID uint) ([]models.ServerResponse, error) {
	// ARKServers
	if imageName == "tbro98/ase-server:latest" {
		// GetServers caps limit at maxPageLimit, so page through instead of asking
		// for one oversized page.
		var affected []models.ServerResponse
		for page := 1; ; page++ {
			servers, total, err := s.GetServers(userID, page, maxPageLimit)
			if err != nil {
				return nil, err
			}
			affected = append(affected, servers...)
			if len(servers) == 0 || int64(len(affected)) >= total {
				break
			}
		}
		return affected, nil
	}

	// ，
	return []models.ServerResponse{}, nil
}

// RecreateContainer
func (s *ServerService) RecreateContainer(userID uint, serverID string) error {
	mu := s.getUserMutex(userID)
	mu.Lock()
	defer mu.Unlock()

	// serverIDuint
	id, err := strconv.ParseUint(serverID, 10, 32)
	if err != nil {
		return fmt.Errorf("None Server ID")
	}

	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&server).Error; err != nil {
		return fmt.Errorf("Server not found")
	}

	// ServersStatus，Stop
	if server.Status == "running" {
		if err := s.StopServer(userID, serverID); err != nil {
			return fmt.Errorf("Stop server : %w", err)
		}

		// ServersStop
		for i := 0; i < 30; i++ {
			time.Sleep(1 * time.Second)
			if err := database.DB.Where("id = ?", id).First(&server).Error; err == nil {
				if server.Status == "stopped" {
					break
				}
			}
		}
	}

	//
	go func() {
		dockerManager, err := docker_manager.GetDockerManager()
		if err != nil {
			utils.Error(" Docker Manager ", zap.Error(err))
			return
		}

		containerName := utils.GetServerContainerName(server.ID)

		// Delete
		if err := dockerManager.RemoveContainer(containerName); err != nil {
			utils.Error("Delete ", zap.Error(err))
		}

		// Create
		_, err = dockerManager.CreateContainer(
			server.ID,
			server.Identifier,
			server.Port,
			server.QueryPort,
			server.RCONPort,
			server.AdminPassword,
			server.Map,
			server.GameModIds,
			server.AutoRestart,
		)
		if err != nil {
			utils.Error(" ", zap.Error(err))
			return
		}

		utils.Info("Servers ", zap.String("identifier", server.Identifier))
	}()

	return nil
}

// GetImageStatus Get image status
func (s *ServerService) GetImageStatus() (map[string]interface{}, error) {
	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, fmt.Errorf(" Docker Manager : %w", err)
	}

	requiredImages := []string{
		"tbro98/ase-server:latest",
		"alpine:latest",
	}

	imageStatuses := make(map[string]*docker_manager.ImageStatus)
	allReady := true
	anyPulling := false
	pullingCount := 0

	for _, imageName := range requiredImages {
		status := dockerManager.GetImageStatus(imageName)
		imageStatuses[imageName] = status

		if !status.Ready {
			allReady = false
		}

		if status.Pulling {
			anyPulling = true
			pullingCount++
		}
	}

	// Status
	var overallStatus string
	if allReady {
		overallStatus = " "
	} else if anyPulling {
		overallStatus = " "
	} else {
		overallStatus = " ， "
	}

	return map[string]interface{}{
		"images":            imageStatuses,
		"any_pulling":       anyPulling,
		"any_not_ready":     !allReady,
		"can_create_server": true,
		"can_start_server":  allReady,
		"overall_status":    overallStatus,
		"pulling_count":     pullingCount,
		"total_images":      len(requiredImages),
	}, nil
}

// RestartServer stops a server, waits for it to actually be down, then starts it.
//
// The controller used to call StopServer followed immediately by StartServer.
// Both dispatch their work to a goroutine and return straight away, so the start
// raced the stop: the start would find a still-running container (or be undone
// by the stop finishing afterwards) and the server ended up stopped. That is why
// the restart button appeared to only stop the server.
func (s *ServerService) RestartServer(userID uint, serverID string) error {
	// A restart on a server that is already stopped is a plain start, not an
	// error. StopServer rejects a stopped or stopping server, and aborting here
	// meant the restart never started it again - and never re-applied the config
	// either, so a config edit made while the server was down was silently lost.
	var current models.Server
	if err := database.DB.Where("id = ?", serverID).First(&current).Error; err != nil {
		return fmt.Errorf("load server: %w", err)
	}
	if current.Status != "stopped" {
		if err := s.StopServer(userID, serverID); err != nil {
			utils.Warn("restart: stop reported an error; continuing",
				zap.String("server_id", serverID), zap.Error(err))
		}
	}

	id, err := utils.ParseUint(serverID)
	if err != nil {
		return fmt.Errorf("invalid server id")
	}

	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return fmt.Errorf("docker manager: %w", err)
	}
	containerName := utils.GetServerContainerName(id)

	// Wait for the container to actually be down before starting again. ARK also
	// rewrites its config on a clean shutdown, so starting early can race that
	// too.
	const (
		pollInterval = time.Second
		stopTimeout  = 90 * time.Second
	)
	deadline := time.Now().Add(stopTimeout)
	for time.Now().Before(deadline) {
		exists, existsErr := dockerManager.ContainerExists(containerName)
		if existsErr != nil {
			utils.Warn("restart: could not check container state", zap.Error(existsErr))
			break
		}
		if !exists {
			break
		}
		status, statusErr := dockerManager.GetContainerStatus(containerName)
		if statusErr != nil || status != "running" {
			break
		}
		time.Sleep(pollInterval)
	}

	// Re-apply the stored config now, while the server is DOWN.
	//
	// ARK rewrites GameUserSettings.ini on a clean shutdown with its own current
	// settings, so a config edit made while the server was running is destroyed
	// by the very restart meant to apply it. Writing after the shutdown - and
	// before the next start - is the only point where the user's values survive
	// to be read. This is what makes "save config, then restart" actually work.
	if err := s.applyStoredConfig(id); err != nil {
		utils.Warn("restart: could not re-apply stored config",
			zap.Uint("server_id", id), zap.Error(err))
	}

	if err := s.StartServer(userID, serverID); err != nil {
		return fmt.Errorf("start for restart: %w", err)
	}
	return nil
}

// applyStoredConfig writes the server's saved INI files into its volume.
func (s *ServerService) applyStoredConfig(serverID uint) error {
	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return fmt.Errorf("load server: %w", err)
	}

	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return fmt.Errorf("docker manager: %w", err)
	}

	if server.DesiredGameUserSettings != "" {
		if err := dockerManager.WriteConfigFile(serverID, utils.GameUserSettingsFileName, server.DesiredGameUserSettings); err != nil {
			return fmt.Errorf("write GameUserSettings.ini: %w", err)
		}
	}
	if server.DesiredGameIni != "" {
		if err := dockerManager.WriteConfigFile(serverID, utils.GameIniFileName, server.DesiredGameIni); err != nil {
			return fmt.Errorf("write Game.ini: %w", err)
		}
	}
	return nil
}
