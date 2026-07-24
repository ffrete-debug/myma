package server

import (
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// createServerContinue CreateServers（：Docker ）
func (s *ServerService) createServerContinue(userID uint, server *models.Server, req models.ServerRequest, tx *gorm.DB, rollback *docker_manager.RollbackManager) (*models.ServerResponse, error) {
	var err error

	// 5:  Docker 
	dockerManager, getErr := docker_manager.GetDockerManager()
	if getErr != nil {
		err = fmt.Errorf(" Docker Manager : %w", getErr)
		return nil, err
	}

	// 6: Create Docker 
	utils.Info("CreateDocker ", zap.Uint("server_id", server.ID))
	volumeName, volErr := dockerManager.CreateVolume(server.ID)
	if volErr != nil {
		err = fmt.Errorf("CreateDocker : %w", volErr)
		return nil, err
	}

	utils.Info("Docker Created successfully", zap.String("volume", volumeName))

	// ：Delete Docker 
	rollback.AddAction("volume", volumeName, "DeleteDocker ", func() error {
		return dockerManager.RemoveVolume(server.ID)
	})

	// 7: 
	var gameUserSettings string
	var gameIni string

	if req.GameUserSettings != "" {
		if validateErr := utils.ValidateINIContent(req.GameUserSettings); validateErr != nil {
			err = fmt.Errorf("GameUserSettings.ini Error: %w", validateErr)
			return nil, err
		}
		gameUserSettings = req.GameUserSettings
	} else {
		gameUserSettings = utils.GetDefaultGameUserSettings(server.Identifier, server.Map, 70)
	}

	if req.GameIni != "" {
		if validateErr := utils.ValidateINIContent(req.GameIni); validateErr != nil {
			err = fmt.Errorf("Game.ini Error: %w", validateErr)
			return nil, err
		}
		gameIni = req.GameIni
	} else {
		gameIni = utils.GetDefaultGameIni()
	}

	// 8:  GameUserSettings.ini
	utils.Info(" GameUserSettings.ini", zap.Uint("server_id", server.ID))
	if writeErr := dockerManager.WriteConfigFile(server.ID, utils.GameUserSettingsFileName, gameUserSettings); writeErr != nil {
		err = fmt.Errorf(" GameUserSettings.ini : %w", writeErr)
		return nil, err
	}

	// ：Delete（Delete）
	// ：

	// 9:  Game.ini
	utils.Info(" Game.ini", zap.Uint("server_id", server.ID))
	if writeErr := dockerManager.WriteConfigFile(server.ID, utils.GameIniFileName, gameIni); writeErr != nil {
		err = fmt.Errorf(" Game.ini : %w", writeErr)
		return nil, err
	}

	// 10: 
	utils.Info(" ", zap.Uint("server_id", server.ID))
	if commitErr := tx.Commit().Error; commitErr != nil {
		err = fmt.Errorf(" : %w", commitErr)
		return nil, err
	}

	// Success，
	//  Docker ，

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
	if gameUserSettingsContent, readErr := dockerManager.ReadConfigFile(server.ID, utils.GameUserSettingsFileName); readErr == nil {
		response.GameUserSettings = gameUserSettingsContent
	}
	if gameIniContent, readErr := dockerManager.ReadConfigFile(server.ID, utils.GameIniFileName); readErr == nil {
		response.GameIni = gameIniContent
	}

	// Success，
	rollback.Clear()

	return &response, nil
}
