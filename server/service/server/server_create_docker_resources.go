package server

import (
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
	"fmt"
	"go.uber.org/zap"
)

// createDockerResources Create Docker （）
func (s *ServerService) createDockerResources(userID uint, server *models.Server, req models.ServerRequest, dockerRollback *docker_manager.RollbackManager) (*models.ServerResponse, error) {
	var err error

	//  Docker 
	dockerManager, getErr := docker_manager.GetDockerManager()
	if getErr != nil {
		// Docker ，Delete
		database.DB.Unscoped().Delete(server)
		return nil, fmt.Errorf(" Docker Manager : %w", getErr)
	}

	// 1: Create Docker 
	utils.Info("CreateDocker ", zap.Uint("server_id", server.ID))
	volumeName, volErr := dockerManager.CreateVolume(server.ID)
	if volErr != nil {
		// Volume creation failed，Delete
		database.DB.Unscoped().Delete(server)
		return nil, fmt.Errorf("CreateDocker : %w", volErr)
	}

	utils.Info("Docker Created successfully", zap.String("volume", volumeName))

	// ：Delete Docker 
	dockerRollback.AddAction("volume", volumeName, "DeleteDocker ", func() error {
		return dockerManager.RemoveVolume(server.ID)
	})

	// 2: 
	var gameUserSettings string
	var gameIni string

	if req.GameUserSettings != "" {
		if validateErr := utils.ValidateINIContent(req.GameUserSettings); validateErr != nil {
			err = fmt.Errorf("GameUserSettings.ini Error: %w", validateErr)
			//  Docker  + Delete
			database.DB.Unscoped().Delete(server)
			return nil, err
		}
		gameUserSettings = req.GameUserSettings
	} else {
		gameUserSettings = utils.GetDefaultGameUserSettings(server.Identifier, server.Map, 70)
	}

	if req.GameIni != "" {
		if validateErr := utils.ValidateINIContent(req.GameIni); validateErr != nil {
			err = fmt.Errorf("Game.ini Error: %w", validateErr)
			database.DB.Unscoped().Delete(server)
			return nil, err
		}
		gameIni = req.GameIni
	} else {
		gameIni = utils.GetDefaultGameIni()
	}

	// 3: 
	utils.Info(" ", zap.Uint("server_id", server.ID))
	if writeErr := dockerManager.WriteConfigFile(server.ID, utils.GameUserSettingsFileName, gameUserSettings); writeErr != nil {
		err = fmt.Errorf(" GameUserSettings.ini : %w", writeErr)
		database.DB.Unscoped().Delete(server)
		return nil, err
	}

	if writeErr := dockerManager.WriteConfigFile(server.ID, utils.GameIniFileName, gameIni); writeErr != nil {
		err = fmt.Errorf(" Game.ini : %w", writeErr)
		database.DB.Unscoped().Delete(server)
		return nil, err
	}

	utils.Info("Server created successfully",
		zap.Uint("server_id", server.ID),
		zap.String("identifier", server.Identifier))

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
	if content, readErr := dockerManager.ReadConfigFile(server.ID, utils.GameUserSettingsFileName); readErr == nil {
		response.GameUserSettings = content
	}
	if content, readErr := dockerManager.ReadConfigFile(server.ID, utils.GameIniFileName); readErr == nil {
		response.GameIni = content
	}

	// Success， Docker 
	dockerRollback.Clear()

	return &response, nil
}
