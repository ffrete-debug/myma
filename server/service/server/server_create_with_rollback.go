package server

import (
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

// CreateServerWithRollback CreateServers（）
// ，Create
func (s *ServerService) CreateServerWithRollback(userID uint, req models.ServerRequest) (*models.ServerResponse, error) {
	// Create
	rollback := docker_manager.NewRollbackManager()
	var err error

	//  defer Error
	defer func() {
		if err != nil && rollback.Count() > 0 {
			utils.Warn("ServersCreate ，On ", zap.Error(err))
			if rollbackErr := rollback.Rollback(); rollbackErr != nil {
				utils.Error(" Error", zap.Error(rollbackErr))
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

	// 3: On
	tx := database.DB.Begin()
	if tx.Error != nil {
		err = fmt.Errorf(" Start : %w", tx.Error)
		return nil, err
	}

	// ：
	rollback.AddAction("database", "transaction", " ", func() error {
		tx.Rollback()
		return nil
	})

	// 4: CreateServers
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
		argsJson, marshalErr := json.Marshal(req.ServerArgs)
		if marshalErr != nil {
			err = fmt.Errorf("Start Error: %w", marshalErr)
			return nil, err
		}
		server.ServerArgsJSON = string(argsJson)
	} else {
		server.ServerArgsJSON = "{}"
	}

	if createErr := tx.Create(&server).Error; createErr != nil {
		err = fmt.Errorf("ServersCreate : %w", createErr)
		return nil, err
	}

	utils.Info("Servers Created successfully", zap.Uint("server_id", server.ID))

	// ：Delete server
	serverID := server.ID
	rollback.AddAction("database", fmt.Sprintf("server_%d", serverID), "Delete server ", func() error {
		return database.DB.Unscoped().Delete(&models.Server{}, serverID).Error
	})

	// ...
	return s.createServerContinue(userID, &server, req, tx, rollback)
}
