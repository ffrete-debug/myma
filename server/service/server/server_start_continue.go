package server

import (
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"
	"fmt"
	"go.uber.org/zap"
	"time"
)

// startServerAsyncContinue Start server（）
func (s *ServerService) startServerAsyncContinue(server models.Server, dockerManager *docker_manager.DockerManager, containerName string, containerExists, needRecreateContainer bool, rollback *docker_manager.RollbackManager) error {
	var err error

	// 4: Create（）
	if !containerExists || needRecreateContainer {
		utils.Info("Create ", zap.String("container", containerName))

		// Create
		containerID, createErr := dockerManager.CreateContainerWithRollback(
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

		if createErr != nil {
			err = fmt.Errorf("Create : %w", createErr)
			return err
		}

		utils.Info(" Created successfully",
			zap.String("container_id", containerID),
			zap.String("container_name", containerName))

		// ：DeleteCreate
		rollback.AddAction("container", containerName, "Delete ", func() error {
			return dockerManager.RemoveContainer(containerName)
		})
	}

	// 5: Start
	utils.Info("Start ", zap.String("container", containerName))
	if startErr := dockerManager.StartContainer(containerName); startErr != nil {
		err = fmt.Errorf("Start : %w", startErr)
		return err
	}

	// ：Stop
	rollback.AddAction("container", containerName, "Stop ", func() error {
		return dockerManager.StopContainer(containerName)
	})

	// 6: Start
	utils.Info(" Start", zap.String("container", containerName))
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		status, statusErr := dockerManager.GetContainerStatus(containerName)
		if statusErr != nil {
			utils.Debug("Failed to get container status， ", zap.Error(statusErr))
			continue
		}

		if status == "running" {
			utils.Info(" StartSuccess",
				zap.String("container", containerName),
				zap.Int("wait_seconds", i+1))

			// Status
			if updateErr := database.DB.Model(&server).Update("status", "running").Error; updateErr != nil {
				utils.Error("Update Service Status running ", zap.Error(updateErr))
			}

			// Success，
			rollback.Clear()
			return nil
		}
	}

	err = fmt.Errorf(" Start ")
	return err
}
