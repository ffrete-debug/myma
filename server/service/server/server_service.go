package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"

	"go.uber.org/zap"
)

// ServerService Server Management
type ServerService struct {
	userMutexes sync.Map // map[uint]*sync.Mutex por userID
}

// NewServerService CreateServer Service
func NewServerService() *ServerService {
	return &ServerService{}
}

// getUserMutex User，UserServers
func (s *ServerService) getUserMutex(userID uint) *sync.Mutex {
	mu, _ := s.userMutexes.LoadOrStore(userID, &sync.Mutex{})
	return mu.(*sync.Mutex)
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
func (s *ServerService) GetServers(userID uint) ([]models.ServerResponse, error) {
	var servers []models.Server
	if err := database.DB.Where("user_id = ?", userID).Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("Get server list : %w", err)
	}

	dockerManager, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, fmt.Errorf(" Docker Manager : %w", err)
	}

	var serverResponses []models.ServerResponse
	for _, server := range servers {
		// DockerStatus
		containerName := utils.GetServerContainerName(server.ID)
		realTimeStatus := server.Status

		// YesNo
		containerExists, err := dockerManager.ContainerExists(containerName)
		if err == nil && containerExists {
			if dockerStatus, err := dockerManager.GetContainerStatus(containerName); err == nil {
				realTimeStatus = dockerStatus

				// StatusStatus，（）
				if realTimeStatus != server.Status {
					go func(s models.Server, status string) {
						database.DB.Model(&s).Update("status", status)
					}(server, realTimeStatus)
				}
			}
		} else if err == nil && !containerExists && server.Status == "running" {
			// Container not foundStatusYes，StopStatus
			realTimeStatus = "stopped"
			go func(s models.Server) {
				database.DB.Model(&s).Update("status", "stopped")
			}(server)
		}

		serverResponses = append(serverResponses, models.ServerResponse{
			ID:            server.ID,
			Identifier:    server.Identifier,
			SessionName:   server.SessionName,
			ClusterID:     server.ClusterID,
			Port:          server.Port,
			QueryPort:     server.QueryPort,
			RCONPort:      server.RCONPort,
			AdminPassword: server.AdminPassword,
			Map:           server.Map,
			MaxPlayers:    server.MaxPlayers,
			GameModIds:    server.GameModIds,
			Status:        realTimeStatus,
			AutoRestart:   server.AutoRestart,
			UserID:        server.UserID,
			CreatedAt:     server.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     server.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return serverResponses, nil
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
		gameUserSettings = utils.GetDefaultGameUserSettings(server.Identifier, server.Map, 70)
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
		dockerManager.RemoveVolume(server.ID)
		return nil, fmt.Errorf(" : %w", err)
	}

	// 
	response := models.ServerResponse{
		ID:            server.ID,
		Identifier:    server.Identifier,
		SessionName:   server.SessionName,
		ClusterID:     server.ClusterID,
		Port:          server.Port,
		QueryPort:     server.QueryPort,
		RCONPort:      server.RCONPort,
		AdminPassword: server.AdminPassword,
		Map:           server.Map,
		GameModIds:    server.GameModIds,
		Status:        server.Status,
		AutoRestart:   server.AutoRestart,
		UserID:        server.UserID,
		CreatedAt:     server.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     server.UpdatedAt.Format("2006-01-02 15:04:05"),
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
		AdminPassword: server.AdminPassword,
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
		}

		if req.GameIni != "" {
			if err := utils.ValidateINIContent(req.GameIni); err != nil {
				return nil, false, fmt.Errorf("Game.ini Error: %w", err)
			}
			if err := dockerManager.WriteConfigFile(uint(id), utils.GameIniFileName, req.GameIni); err != nil {
				return nil, false, fmt.Errorf(" Game.ini : %w", err)
			}
		}
	}

	// 
	response := models.ServerResponse{
		ID:            server.ID,
		Identifier:    server.Identifier,
		SessionName:   server.SessionName,
		ClusterID:     server.ClusterID,
		Port:          server.Port,
		QueryPort:     server.QueryPort,
		RCONPort:      server.RCONPort,
		AdminPassword: server.AdminPassword,
		Map:           server.Map,
		MaxPlayers:    server.MaxPlayers,
		GameModIds:    server.GameModIds,
		Status:        server.Status,
		AutoRestart:   server.AutoRestart,
		UserID:        server.UserID,
		CreatedAt:     server.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     server.UpdatedAt.Format("2006-01-02 15:04:05"),
		ServerArgs:    models.FromServer(server),
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

	return &response, argsChanged, nil
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
		return
	}

	if !containerExists {
		database.DB.Model(&server).Update("status", "stopped")
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
		return s.GetServers(userID)
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
