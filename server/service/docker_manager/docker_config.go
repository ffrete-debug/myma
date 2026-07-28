package docker_manager

import (
	"archive/tar"
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"go.uber.org/zap"
)

// ReadConfigFile
// serverID: Server ID
// fileName:
// : Error
func (dm *DockerManager) ReadConfigFile(serverID uint, fileName string) (string, error) {
	volumeName := utils.GetServerVolumeName(serverID)
	alpineImage := "alpine:latest"

	// AlpineYesNo
	if err := dm.ensureAlpine(); err != nil {
		return "", err
	}

	// （ Config/WindowsServer ）
	configPath := fmt.Sprintf("/home/steam/arkserver/ShooterGame/Saved/Config/WindowsServer/%s", fileName)

	// Create
	containerConfig := &container.Config{
		Image: alpineImage,
		Cmd:   []string{"tail", "-f", "/dev/null"}, //
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/home/steam/arkserver/ShooterGame/Saved", volumeName),
		},
	}

	ctx, cancel := dm.opCtx(dockerOpTimeout)
	defer cancel()

	// Create
	resp, err := dm.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("Create : %v", err)
	}

	// ，
	defer dm.removeHelperContainer(resp.ID)

	// Start
	err = dm.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return "", fmt.Errorf("Start : %v", err)
	}

	//
	reader, _, err := dm.client.CopyFromContainer(ctx, resp.ID, configPath)
	if err != nil {
		return "", fmt.Errorf(" : %v", err)
	}
	defer reader.Close()

	//  tar
	tarReader := tar.NewReader(reader)

	// （tar ）
	header, err := tarReader.Next()
	if err != nil {
		if err == io.EOF {
			return "", fmt.Errorf(" : %s", fileName)
		}
		return "", fmt.Errorf("  tar  : %v", err)
	}

	// YesNoYes
	if header.Typeflag != tar.TypeReg {
		return "", fmt.Errorf(" Yes : %s", fileName)
	}

	//
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, tarReader)
	if err != nil {
		return "", fmt.Errorf(" : %v", err)
	}

	content := buffer.String()

	// ，
	content = strings.TrimSpace(content)

	return content, nil
}

// WriteConfigFile
// serverID: Server ID
// fileName:
// content:
// : Error
func (dm *DockerManager) WriteConfigFile(serverID uint, fileName, content string) error {
	volumeName := utils.GetServerVolumeName(serverID)

	// SessionName is not passed on the command line (the image expands
	// SERVER_ARGS unquoted, so a space truncates it and drops later query
	// parameters), which makes this file the authoritative source for the name
	// shown in the Steam browser. Force it to match the server record here, at
	// the single point every config write funnels through, so a rename cannot
	// leave the INI stale.
	if fileName == utils.GameUserSettingsFileName {
		var server models.Server
		if err := database.DB.Where("id = ?", serverID).First(&server).Error; err == nil {
			content = utils.EnsureSessionName(content, server.SessionName)
		} else {
			utils.Warn("could not load server to sync SessionName into config",
				zap.Uint("server_id", serverID), zap.Error(err))
		}
	}
	alpineImage := "alpine:latest"

	// AlpineYesNo
	if err := dm.ensureAlpine(); err != nil {
		return err
	}

	// （ Config/WindowsServer ）
	configDir := "/home/steam/arkserver/ShooterGame/Saved/Config/WindowsServer"

	// Create
	containerConfig := &container.Config{
		Image: alpineImage,
		Cmd:   []string{"tail", "-f", "/dev/null"}, //
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/home/steam/arkserver/ShooterGame/Saved", volumeName),
		},
	}

	ctx, cancel := dm.opCtx(dockerOpTimeout)
	defer cancel()

	// CreateCreate
	resp, err := dm.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("Create : %v", err)
	}

	// ，
	defer dm.removeHelperContainer(resp.ID)

	// Start
	err = dm.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("Start : %v", err)
	}

	// Create
	if err := dm.execInHelperContainer(ctx, resp.ID, fmt.Sprintf("mkdir -p %s", ShellQuote(configDir))); err != nil {
		return fmt.Errorf("Create : %v", err)
	}

	// 、 tar  CopyToContainer 。
	//  Alpine  sh  echo  \n / \t ，
	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	if err := tw.WriteHeader(&tar.Header{
		Name:     fileName,
		Typeflag: tar.TypeReg,
		Mode:     0644,
		Size:     int64(len(content)),
		ModTime:  time.Now(),
	}); err != nil {
		return fmt.Errorf(" tar  : %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		return fmt.Errorf(" tar  : %v", err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf(" tar  : %v", err)
	}

	err = dm.client.CopyToContainer(ctx, resp.ID, configDir, &archive, container.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf(" : %v", err)
	}

	utils.Info(" Success", zap.String("file", fileName))
	return nil
}

// removeHelperContainer Delete
// ， dm.ctx
func (dm *DockerManager) removeHelperContainer(containerID string) {
	ctx, cancel := dm.opCtx(dockerOpTimeout)
	defer cancel()

	if err := dm.client.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		utils.Warn(" Delete ", zap.String("container_id", containerID), zap.Error(err))
	}
}

// execInHelperContainer ，
// : Error
func (dm *DockerManager) execInHelperContainer(ctx context.Context, containerID, command string) error {
	execResp, err := dm.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          []string{"sh", "-c", command},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("Create : %v", err)
	}

	attach, err := dm.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf(" : %v", err)
	}
	// EOF
	_, err = io.Copy(io.Discard, attach.Reader)
	attach.Close()
	if err != nil {
		return fmt.Errorf(" : %v", err)
	}

	inspectResp, err := dm.client.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return fmt.Errorf(" : %v", err)
	}
	if inspectResp.ExitCode != 0 {
		return fmt.Errorf(" ，Logout : %d", inspectResp.ExitCode)
	}

	return nil
}
