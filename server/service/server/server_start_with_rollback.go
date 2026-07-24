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

// startServerAsyncWithRollback Start server（）
func (s *ServerService) startServerAsyncWithRollback(server models.Server, dockerManager *docker_manager.DockerManager, containerName string) error {
	// Create
	rollback := docker_manager.NewRollbackManager()
	var err error

	//  defer Error
	defer func() {
		if err != nil {
			utils.Warn("Server failed to start，On ", zap.Error(err))
			// Update ServiceStatusStop
			database.DB.Model(&server).Update("status", "stopped")

			if rollback.Count() > 0 {
				if rollbackErr := rollback.Rollback(); rollbackErr != nil {
					utils.Error(" Error", zap.Error(rollbackErr))
				}
			}
		}
	}()

	utils.Info("On Start server（ ）",
		zap.String("container", containerName),
		zap.Uint("server_id", server.ID))

	// 1: YesNo
	missingImages, validateErr := dockerManager.ValidateRequiredImages()
	if validateErr != nil {
		err = fmt.Errorf(" : %w", validateErr)
		return err
	}
	if len(missingImages) > 0 {
		err = fmt.Errorf("None Start server， : %v。 Start server", missingImages)
		return err
	}

	// 2: YesNo
	containerExists, checkErr := dockerManager.ContainerExists(containerName)
	if checkErr != nil {
		err = fmt.Errorf(" YesNo : %w", checkErr)
		return err
	}

	needRecreateContainer := false

	// 3: ，YesNo
	if containerExists {
		envVars, envErr := dockerManager.GetContainerEnvVars(containerName)
		if envErr != nil {
			needRecreateContainer = true
			utils.Info("None ， ", zap.Error(envErr))
		} else {
			// ServersStart
			var serverArgs *models.ServerArgs
			if server.ServerArgsJSON != "" && server.ServerArgsJSON != "{}" {
				serverArgs = models.NewServerArgs()
				if unmarshalErr := json.Unmarshal([]byte(server.ServerArgsJSON), serverArgs); unmarshalErr != nil {
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
					utils.Info("Servers ， ")
				}
			} else {
				needRecreateContainer = true
			}

			// 
			if !needRecreateContainer {
				if server.GameModIds != envVars["GameModIds"] {
					needRecreateContainer = true
					utils.Info("Mod ， ")
				}
			}
		}

		// ，Delete
		if needRecreateContainer {
			utils.Info("Delete ", zap.String("container", containerName))
			if removeErr := dockerManager.RemoveContainer(containerName); removeErr != nil {
				err = fmt.Errorf("Delete : %w", removeErr)
				return err
			}
		}
	}

	// ...
	return s.startServerAsyncContinue(server, dockerManager, containerName, containerExists, needRecreateContainer, rollback)
}
