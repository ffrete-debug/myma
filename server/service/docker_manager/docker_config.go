package docker_manager

import (
	"archive/tar"
	"ark-server-commander/utils"
	"bytes"
	"fmt"
	"io"
	"strings"

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
	exists, err := dm.ImageExists(alpineImage)
	if err != nil {
		return "", fmt.Errorf(" Alpine : %v", err)
	}

	if !exists {
		return "", fmt.Errorf("Alpine ， Start Success ")
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

	// Create
	resp, err := dm.client.ContainerCreate(dm.ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("Create : %v", err)
	}

	// Start
	err = dm.client.ContainerStart(dm.ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return "", fmt.Errorf("Start : %v", err)
	}

	// 
	defer func() {
		dm.client.ContainerRemove(dm.ctx, resp.ID, container.RemoveOptions{
			Force: true,
		})
	}()

	// 
	reader, _, err := dm.client.CopyFromContainer(dm.ctx, resp.ID, configPath)
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
	alpineImage := "alpine:latest"

	// AlpineYesNo
	exists, err := dm.ImageExists(alpineImage)
	if err != nil {
		return fmt.Errorf(" Alpine : %v", err)
	}

	if !exists {
		return fmt.Errorf("Alpine ， Start Success ")
	}

	// （ Config/WindowsServer ）
	configPath := fmt.Sprintf("/home/steam/arkserver/ShooterGame/Saved/Config/WindowsServer/%s", fileName)
	configDir := "/home/steam/arkserver/ShooterGame/Saved/Config/WindowsServer"

	// 
	containerConfig := &container.Config{
		Image: alpineImage,
		Cmd:   []string{"mkdir", "-p", configDir},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/home/steam/arkserver/ShooterGame/Saved", volumeName),
		},
	}

	// CreateCreate
	resp, err := dm.client.ContainerCreate(dm.ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("Create : %v", err)
	}

	// Start
	err = dm.client.ContainerStart(dm.ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("Start : %v", err)
	}

	// 
	waitCh, errCh := dm.client.ContainerWait(dm.ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf(" : %v", err)
		}
	case <-waitCh:
		// 
	}

	// Delete
	dm.client.ContainerRemove(dm.ctx, resp.ID, container.RemoveOptions{
		Force: true,
	})

	// 
	containerConfig = &container.Config{
		Image: alpineImage,
		Cmd:   []string{"sh", "-c", fmt.Sprintf("echo '%s' > %s", strings.ReplaceAll(content, "'", "'\"'\"'"), configPath)},
	}

	// Create
	resp, err = dm.client.ContainerCreate(dm.ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("Create : %v", err)
	}

	// Start
	err = dm.client.ContainerStart(dm.ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("Start : %v", err)
	}

	// 
	waitCh, errCh = dm.client.ContainerWait(dm.ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf(" : %v", err)
		}
	case <-waitCh:
		// 
	}

	// Delete
	dm.client.ContainerRemove(dm.ctx, resp.ID, container.RemoveOptions{
		Force: true,
	})

	utils.Info(" Success", zap.String("file", fileName))
	return nil
}
