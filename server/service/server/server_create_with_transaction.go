package server

import (
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateServerWithTransaction CreateServers（ + Docker ）
// ：，Docker 
func (s *ServerService) CreateServerWithTransaction(userID uint, req models.ServerRequest) (*models.ServerResponse, error) {
	var server models.Server
	var err error

	// Docker 
	dockerRollback := docker_manager.NewRollbackManager()

	//  defer  Docker 
	defer func() {
		if err != nil && dockerRollback.Count() > 0 {
			utils.Warn("Create ，  Docker  ", zap.Error(err))
			if rollbackErr := dockerRollback.Rollback(); rollbackErr != nil {
				utils.Error("Docker  ", zap.Error(rollbackErr))
			}
		}
	}()

	utils.Info("On CreateServers（ ）",
		zap.String("identifier", req.Identifier),
		zap.Uint("user_id", userID))

	// 1: ServersYesNo
	var existingServer models.Server
	if checkErr := database.DB.Where("identifier = ? AND user_id = ?", req.Identifier, userID).First(&existingServer).Error; checkErr == nil {
		err = fmt.Errorf("Server identifier already exists")
		return nil, err
	}

	// 2: Settings
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

	// 3:  - CreateServers
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// CreateServers
		server = models.Server{
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
			argsJson, marshalErr := json.Marshal(req.ServerArgs)
			if marshalErr != nil {
				return fmt.Errorf("Start Error: %w", marshalErr)
			}
			server.ServerArgsJSON = string(argsJson)
		} else {
			server.ServerArgsJSON = "{}"
		}

		// Create
		if createErr := tx.Create(&server).Error; createErr != nil {
			return fmt.Errorf("ServersCreate : %w", createErr)
		}

		utils.Info("Servers Created successfully（ ）", zap.Uint("server_id", server.ID))
		return nil
	})

	// ，（GORM ）
	if err != nil {
		return nil, err
	}

	// Create Docker ...
	return s.createDockerResources(userID, &server, req, dockerRollback)
}
