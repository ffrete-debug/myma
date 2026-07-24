package docker_manager

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"

	"ark-server-commander/utils"
)

type FileInfo struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
}

func (dm *DockerManager) ensureAlpine() error {
	exists, err := dm.ImageExists("alpine:latest")
	if err != nil {
		return fmt.Errorf("check alpine image failed: %v", err)
	}
	if !exists {
		return fmt.Errorf("alpine:latest image not found")
	}
	return nil
}

func (dm *DockerManager) runTempContainer(cmd []string, binds []string) (string, error) {
	if err := dm.ensureAlpine(); err != nil {
		return "", err
	}
	cc := &container.Config{Image: "alpine:latest", Cmd: cmd}
	hc := &container.HostConfig{Binds: binds}
	resp, err := dm.client.ContainerCreate(dm.ctx, cc, hc, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("container create: %v", err)
	}
	cid := resp.ID
	defer dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true})

	if err := dm.client.ContainerStart(dm.ctx, cid, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("container start: %v", err)
	}

	waitCh, errCh := dm.client.ContainerWait(dm.ctx, cid, container.WaitConditionNotRunning)
	select {
	case e := <-errCh:
		return "", fmt.Errorf("container wait: %v", e)
	case <-waitCh:
	}

	out, err := dm.client.ContainerLogs(dm.ctx, cid, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}
	defer out.Close()
	var buf bytes.Buffer
	stdcopy.StdCopy(&buf, &buf, out)
	return strings.TrimSpace(buf.String()), nil
}

func (dm *DockerManager) RunCommandInVolume(volumeName, volumeMount string, cmd []string) (string, error) {
	bind := fmt.Sprintf("%s:%s", volumeName, volumeMount)
	return dm.runTempContainer(cmd, []string{bind})
}

func (dm *DockerManager) ListFiles(volumeName, volumeMount, dirPath string) ([]FileInfo, error) {
	cmd := []string{"sh", "-c", fmt.Sprintf(
		`find '%s' -mindepth 1 -maxdepth 1 -exec stat -c '%%F|%%s|%%Y|%%a|%%n' {} \; 2>/dev/null`,
		dirPath,
	)}
	bind := fmt.Sprintf("%s:%s", volumeName, volumeMount)

	out, err := dm.runTempContainer(cmd, []string{bind})
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		fileName := filepath.Base(parts[4])
		if fileName == "." || fileName == ".." || fileName == "/" {
			continue
		}

		fi := FileInfo{Name: fileName, Mode: parts[3]}
		if parts[0] == "directory" {
			fi.IsDir = true
		}
		if s, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			fi.Size = s
		}
		if s, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
			fi.ModTime = strconv.FormatInt(s, 10)
		}

		if fi.Name == "" || fi.Name == "." {
			continue
		}

		files = append(files, fi)
	}
	return files, nil
}

func (dm *DockerManager) ReadFileFromVolume(volumeName, volumeMount, filePath string) (io.ReadCloser, error) {
	if err := dm.ensureAlpine(); err != nil {
		return nil, err
	}
	cc := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}
	hc := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", volumeName, volumeMount)},
	}
	resp, err := dm.client.ContainerCreate(dm.ctx, cc, hc, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("container create: %v", err)
	}
	cid := resp.ID
	defer dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true})

	if err := dm.client.ContainerStart(dm.ctx, cid, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("container start: %v", err)
	}

	reader, _, err := dm.client.CopyFromContainer(dm.ctx, cid, filePath)
	if err != nil {
		return nil, fmt.Errorf("copy from container: %v", err)
	}
	return reader, nil
}

func (dm *DockerManager) WriteFileToVolume(volumeName, volumeMount, destPath string, content io.Reader) error {
	if err := dm.ensureAlpine(); err != nil {
		return err
	}

	cc := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"tail", "-f", "/dev/null"},
	}
	hc := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:%s", volumeName, volumeMount)},
	}
	resp, err := dm.client.ContainerCreate(dm.ctx, cc, hc, nil, nil, "")
	if err != nil {
		return fmt.Errorf("container create: %v", err)
	}
	cid := resp.ID
	defer dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true})

	if err := dm.client.ContainerStart(dm.ctx, cid, container.StartOptions{}); err != nil {
		return fmt.Errorf("container start: %v", err)
	}

	parentDir := filepath.Dir(destPath)
	mkdirCmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p '%s'", strings.ReplaceAll(parentDir, "'", "'\"'\"'"))}
	execCfg := container.ExecOptions{
		Cmd:          mkdirCmd,
		AttachStdout: true,
		AttachStderr: true,
	}
	execResp, execErr := dm.client.ContainerExecCreate(dm.ctx, cid, execCfg)
	if execErr == nil {
		dm.client.ContainerExecStart(dm.ctx, execResp.ID, container.ExecStartOptions{})
	}

	utils.Info("writing file to volume", zap.String("file", destPath))
	if err := dm.client.CopyToContainer(dm.ctx, cid, filepath.Dir(destPath), content, container.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("copy to container: %v", err)
	}

	return nil
}
