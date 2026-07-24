package docker_manager

import (
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/utils"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
)

// CreateContainerWithRollback Create（）
// Create，Create
func (dm *DockerManager) CreateContainerWithRollback(serverID uint, serverName string, port, queryPort, rconPort int, adminPassword, mapName, gameModIds string, autoRestart bool) (containerID string, err error) {
	// Create
	rollback := NewRollbackManager()

	//  defer Error
	defer func() {
		if err != nil && rollback.Count() > 0 {
			utils.Warn("Container creation failed，On ", zap.Error(err))
			if rollbackErr := rollback.Rollback(); rollbackErr != nil {
				utils.Error(" Error", zap.Error(rollbackErr))
			}
		}
	}()

	containerName := utils.GetServerContainerName(serverID)
	volumeName := utils.GetServerVolumeName(serverID)
	imageName := "tbro98/ase-server:latest"

	utils.Info("On Create （ ）",
		zap.String("container", containerName),
		zap.Uint("server_id", serverID))

	// 1: Delete
	if exists, checkErr := dm.ContainerExists(containerName); checkErr != nil {
		err = fmt.Errorf(" YesNo : %w", checkErr)
		return "", err
	} else if exists {
		utils.Info(" ， Delete", zap.String("container", containerName))
		if removeErr := dm.RemoveContainer(containerName); removeErr != nil {
			err = fmt.Errorf("Delete : %w", removeErr)
			return "", err
		}
	}

	// 2: YesNo
	imageExists, checkErr := dm.ImageExists(imageName)
	if checkErr != nil {
		err = fmt.Errorf(" YesNo : %w", checkErr)
		return "", err
	}
	if !imageExists {
		err = fmt.Errorf("  %s  ， ", imageName)
		return "", err
	}

	// 3: Servers
	var server models.Server
	if dbErr := database.DB.Where("id = ?", serverID).First(&server).Error; dbErr != nil {
		err = fmt.Errorf(" Servers : %w", dbErr)
		return "", err
	}

	// 4: Start
	serverArgs := models.NewServerArgs()
	if server.ServerArgsJSON != "" && server.ServerArgsJSON != "{}" {
		_ = json.Unmarshal([]byte(server.ServerArgsJSON), serverArgs)
	} else {
		serverArgs = models.FromServer(server)
	}
	argsString := serverArgs.GenerateArgsString(server)

	// 5: 
	envVars := []string{
		"TZ=Asia/Shanghai",
		fmt.Sprintf("SERVER_ARGS=%s", argsString),
	}
	if server.GameModIds != "" {
		envVars = append(envVars, fmt.Sprintf("GameModIds=%s", server.GameModIds))
	}

	// 6: 
	containerConfig := &container.Config{
		Image: imageName,
		Env:   envVars,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/udp", port)):      struct{}{},
			nat.Port(fmt.Sprintf("%d/tcp", port)):      struct{}{},
			nat.Port(fmt.Sprintf("%d/udp", port+1)):    struct{}{},
			nat.Port(fmt.Sprintf("%d/tcp", port+1)):    struct{}{},
			nat.Port(fmt.Sprintf("%d/udp", queryPort)): struct{}{},
			nat.Port(fmt.Sprintf("%d/tcp", queryPort)): struct{}{},
			nat.Port(fmt.Sprintf("%d/udp", rconPort)):  struct{}{},
			nat.Port(fmt.Sprintf("%d/tcp", rconPort)):  struct{}{},
		},
	}

	// 7: SettingsRestart
	restartPolicyName := container.RestartPolicyMode("unless-stopped")
	if !autoRestart {
		restartPolicyName = container.RestartPolicyMode("no")
	}

	// 8: 
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: restartPolicyName,
		},
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/udp", port)): {
				{HostPort: fmt.Sprintf("%d", port)},
			},
			nat.Port(fmt.Sprintf("%d/tcp", port)): {
				{HostPort: fmt.Sprintf("%d", port)},
			},
			nat.Port(fmt.Sprintf("%d/udp", port+1)): {
				{HostPort: fmt.Sprintf("%d", port+1)},
			},
			nat.Port(fmt.Sprintf("%d/tcp", port+1)): {
				{HostPort: fmt.Sprintf("%d", port+1)},
			},
			nat.Port(fmt.Sprintf("%d/udp", queryPort)): {
				{HostPort: fmt.Sprintf("%d", queryPort)},
			},
			nat.Port(fmt.Sprintf("%d/tcp", queryPort)): {
				{HostPort: fmt.Sprintf("%d", queryPort)},
			},
			nat.Port(fmt.Sprintf("%d/udp", rconPort)): {
				{HostPort: fmt.Sprintf("%d", rconPort)},
			},
			nat.Port(fmt.Sprintf("%d/tcp", rconPort)): {
				{HostPort: fmt.Sprintf("%d", rconPort)},
			},
		},
		Binds: []string{
			fmt.Sprintf("%s:/home/steam/arkserver/ShooterGame/Saved", volumeName),
			fmt.Sprintf("%s:/home/steam/arkserver/ShooterGame/Binaries/Win64/ArkApi/Plugins", utils.GetServerPluginsVolumeName(serverID)),
		},
	}

	// 9: Create
	utils.Info(" CreateDocker ", zap.String("container", containerName))
	resp, createErr := dm.client.ContainerCreate(dm.ctx, containerConfig, hostConfig, nil, nil, containerName)
	if createErr != nil {
		err = fmt.Errorf("CreateDocker : %w", createErr)
		return "", err
	}

	containerID = resp.ID

	// ：DeleteCreate
	rollback.AddAction("container", containerName, "Delete ", func() error {
		return dm.RemoveContainer(containerName)
	})

	utils.Info("Docker Created successfully",
		zap.String("container_name", containerName),
		zap.String("container_id", containerID))

	// Success，
	rollback.Clear()

	return containerID, nil
}
