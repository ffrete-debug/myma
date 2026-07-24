package docker_manager

import (
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
)

// DockerManager Docker Manager
type DockerManager struct {
	client *client.Client
	ctx    context.Context
}

// Off
var (
	instance *DockerManager
	once     sync.Once
)

// GetDockerManager Docker Manager
func GetDockerManager() (*DockerManager, error) {
	var err error
	once.Do(func() {
		// Create
		cli, clientErr := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if clientErr != nil {
			err = fmt.Errorf("CreateDocker : %v", clientErr)
			return
		}

		instance = &DockerManager{
			client: cli,
			ctx:    context.Background(),
		}
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// CloseDockerManager OffDocker Manager（Logout）
func CloseDockerManager() error {
	if instance != nil && instance.client != nil {
		err := instance.client.Close()
		instance = nil
		return err
	}
	return nil
}

// Check DockerStatus DockerStatus
// : DockerYesNoError
func CheckDockerStatus() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("docker Create : %v", err)
	}
	defer cli.Close()

	// DockerYesNo
	_, err = cli.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("docker None : %v", err)
	}

	return nil
}

// ValidateRequiredImages YesNo（）
func (dm *DockerManager) ValidateRequiredImages() ([]string, error) {
	requiredImages := []string{
		"tbro98/ase-server:latest", // ARKServers
		"alpine:latest",            // Alpine（）
	}

	var missingImages []string
	for _, imageName := range requiredImages {
		exists, err := dm.ImageExists(imageName)
		if err != nil {
			utils.Warnf("  %s  : %v", imageName, err)
			return nil, fmt.Errorf("  %s  : %v", imageName, err)
		}

		if !exists {
			missingImages = append(missingImages, imageName)
			utils.Error(" ", zap.String("image", imageName))
		} else {
			utils.Info(" ", zap.String("image", imageName))
		}
	}

	return missingImages, nil
}

// EnsureRequiredImages （，ValidateRequiredImages）
// ，
func (dm *DockerManager) EnsureRequiredImages() error {
	missingImages, err := dm.ValidateRequiredImages()
	if err != nil {
		return err
	}

	if len(missingImages) > 0 {
		return fmt.Errorf(" : %v， ", missingImages)
	}

	utils.Info(" ")
	return nil
}

// CreateContainer CreateARKServers（Start）
// serverID: Server ID
// serverName: Servers
// port: 
// queryPort: Query Port
// rconPort: RCON Port
// adminPassword: Password
// mapName: Map
// gameModIds: ID，
// autoRestart: YesNoRestart
// : IDError
func (dm *DockerManager) CreateContainer(serverID uint, serverName string, port, queryPort, rconPort int, adminPassword, mapName, gameModIds string, autoRestart bool) (string, error) {
	containerName := utils.GetServerContainerName(serverID)
	volumeName := utils.GetServerVolumeName(serverID)
	imageName := "tbro98/ase-server:latest"

	// YesNo
	if exists, err := dm.ContainerExists(containerName); err != nil {
		return "", fmt.Errorf(" YesNo : %v", err)
	} else if exists {
		// ，Delete
		if err := dm.RemoveContainer(containerName); err != nil {
			return "", fmt.Errorf("Delete : %v", err)
		}
	}

	// YesNo
	exists, err := dm.ImageExists(imageName)
	if err != nil {
		return "", fmt.Errorf(" YesNo : %v", err)
	}
	if !exists {
		return "", fmt.Errorf("  %s  ， ", imageName)
	}

	// 1. Server
	var server models.Server
	if err := database.DB.Where("id = ?", serverID).First(&server).Error; err != nil {
		return "", fmt.Errorf(" Servers : %v", err)
	}

	// 2. ServerArgsJSON
	serverArgs := models.NewServerArgs()
	if server.ServerArgsJSON != "" && server.ServerArgsJSON != "{}" {
		_ = json.Unmarshal([]byte(server.ServerArgsJSON), serverArgs)
	} else {
		serverArgs = models.FromServer(server)
	}
	argsString := serverArgs.GenerateArgsString(server)

	// 3. 
	envVars := []string{
		"TZ=Asia/Shanghai",
		fmt.Sprintf("SERVER_ARGS=%s", argsString),
	}
	if server.GameModIds != "" {
		envVars = append(envVars, fmt.Sprintf("GameModIds=%s", server.GameModIds))
	}

	// 
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

	// autoRestartSettingsRestart
	restartPolicyName := container.RestartPolicyMode("unless-stopped")
	if !autoRestart {
		restartPolicyName = container.RestartPolicyMode("no")
	}

	// 
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

	// Create
	utils.Infof(" CreateDocker : %s", containerName)
	resp, err := dm.client.ContainerCreate(dm.ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("CreateDocker : %v", err)
	}

	utils.Info("Docker Created successfully， StopStatus",
		zap.String("container_name", containerName),
		zap.String("container_id", resp.ID),
	)
	return resp.ID, nil
}

// StartContainer Start
// containerName: 
// : Error
func (dm *DockerManager) StartContainer(containerName string) error {
	utils.Infof(" Start : %s", containerName)
	err := dm.client.ContainerStart(dm.ctx, containerName, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("StartDocker : %v", err)
	}

	utils.Info(" StartSuccess", zap.String("container_name", containerName))
	return nil
}

// StopContainer Stop
// containerName: 
// : Error
func (dm *DockerManager) StopContainer(containerName string) error {
	utils.Infof(" Stop : %s", containerName)

	// Settings30
	timeout := 30
	err := dm.client.ContainerStop(dm.ctx, containerName, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("StopDocker : %v", err)
	}

	utils.Info(" StopSuccess", zap.String("container_name", containerName))
	return nil
}

// RemoveContainer Delete
// containerName: 
// : Error
func (dm *DockerManager) RemoveContainer(containerName string) error {
	// Stop
	dm.StopContainer(containerName)

	// Delete
	utils.Infof(" Delete : %s", containerName)
	err := dm.client.ContainerRemove(dm.ctx, containerName, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("DeleteDocker : %v", err)
	}

	utils.Info(" DeleteSuccess", zap.String("container_name", containerName))
	return nil
}

// ContainerExists YesNo
// containerName: 
// : YesNoError
func (dm *DockerManager) ContainerExists(containerName string) (bool, error) {
	// inspectYesNo
	_, err := dm.client.ContainerInspect(dm.ctx, containerName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil // Container not found
		}
		return false, fmt.Errorf(" Docker : %v", err)
	}

	return true, nil
}

// GetContainerStatus Status
// containerName: 
// : StatusError
func (dm *DockerManager) GetContainerStatus(containerName string) (string, error) {
	containerInfo, err := dm.client.ContainerInspect(dm.ctx, containerName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return "not_found", nil
		}
		return "", fmt.Errorf(" Docker Status : %v", err)
	}

	// DockerStatusStatus
	state := containerInfo.State
	if state.Running {
		return "running", nil
	} else if state.Status == "exited" {
		return "stopped", nil
	} else if state.Status == "created" {
		return "stopped", nil
	} else if state.Status == "restarting" {
		return "starting", nil
	} else {
		return "unknown", nil
	}
}

// ExecuteCommand 
// containerName: 
// command: 
// : Error
func (dm *DockerManager) ExecuteCommand(containerName string, command string) (string, error) {
	// Create
	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", command},
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create
	execResp, err := dm.client.ContainerExecCreate(dm.ctx, containerName, execConfig)
	if err != nil {
		return "", fmt.Errorf("Create : %v", err)
	}

	// 
	resp, err := dm.client.ContainerExecAttach(dm.ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return "", fmt.Errorf(" : %v", err)
	}
	defer resp.Close()

	// 
	output, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", fmt.Errorf(" : %v", err)
	}

	// 
	inspectResp, err := dm.client.ContainerExecInspect(dm.ctx, execResp.ID)
	if err != nil {
		return "", fmt.Errorf(" : %v", err)
	}

	if inspectResp.ExitCode != 0 {
		return string(output), fmt.Errorf(" ，Logout : %d", inspectResp.ExitCode)
	}

	return string(output), nil
}

// GetContainerEnvVars 
// containerName: 
// : Error
func (dm *DockerManager) GetContainerEnvVars(containerName string) (map[string]string, error) {
	containerInfo, err := dm.client.ContainerInspect(dm.ctx, containerName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, fmt.Errorf("Container not found: %s", containerName)
		}
		return nil, fmt.Errorf(" Docker : %v", err)
	}

	// 
	envVars := make(map[string]string)
	for _, env := range containerInfo.Config.Env {
		// : KEY=VALUE
		for i, char := range env {
			if char == '=' {
				key := env[:i]
				value := env[i+1:]
				envVars[key] = value
				break
			}
		}
	}

	return envVars, nil
}

// GetContainerLogs Container Logs
// containerName: 
// tail: N，0
// : Error
func (dm *DockerManager) GetContainerLogs(containerName string, tail int) (string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}
	if tail > 0 {
		options.Tail = strconv.Itoa(tail)
	}
	reader, err := dm.client.ContainerLogs(dm.ctx, containerName, options)
	if err != nil {
		return "", fmt.Errorf(" Container Logs : %v", err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = stdcopy.StdCopy(&buf, &buf, reader)
	if err != nil {
		return "", fmt.Errorf(" Container Logs : %v", err)
	}
	return strings.TrimSpace(buf.String()), nil
}
