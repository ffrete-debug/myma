package docker_manager

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// ShellQuote wraps s in single quotes and escapes any embedded single quote,
// so the result can be safely interpolated into a `sh -c` command string.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// ensureAlpine makes the helper image available, pulling it if necessary.
//
// This image is an internal implementation detail - it is used to read and write
// config files and to build backup archives - so requiring the operator to have
// pulled it by hand is not reasonable. It previously returned an error when the
// image was absent, which made server creation fail on a fresh install with a
// message that did not say what to do about it.
func (dm *DockerManager) ensureAlpine() error {
	exists, err := dm.ImageExists("alpine:latest")
	if err != nil {
		return fmt.Errorf("check alpine image failed: %v", err)
	}
	if exists {
		return nil
	}

	utils.Info("helper image alpine:latest is missing; pulling it")

	ctx, cancel := context.WithTimeout(dm.ctx, 5*time.Minute)
	defer cancel()

	reader, err := dm.client.ImagePull(ctx, "alpine:latest", image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pull alpine:latest: %w", err)
	}
	defer func() { _ = reader.Close() }()

	// The pull only completes once the response body has been consumed.
	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("pull alpine:latest: %w", err)
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
	defer func() { _ = dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true}) }()

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
	if _, err := stdcopy.StdCopy(&buf, &buf, out); err != nil {
		return "", fmt.Errorf("read command output: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}

func (dm *DockerManager) RunCommandInVolume(volumeName, volumeMount string, cmd []string) (string, error) {
	bind := fmt.Sprintf("%s:%s", volumeName, volumeMount)
	return dm.runTempContainer(cmd, []string{bind})
}

func (dm *DockerManager) ListFiles(volumeName, volumeMount, dirPath string) ([]FileInfo, error) {
	cmd := []string{"sh", "-c", fmt.Sprintf(
		`find %s -mindepth 1 -maxdepth 1 -exec stat -c '%%F|%%s|%%Y|%%a|%%n' {} \; 2>/dev/null`,
		ShellQuote(dirPath),
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
	defer func() { _ = dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true}) }()

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
	defer func() { _ = dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true}) }()

	if err := dm.client.ContainerStart(dm.ctx, cid, container.StartOptions{}); err != nil {
		return fmt.Errorf("container start: %v", err)
	}

	parentDir := filepath.Dir(destPath)
	mkdirCmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p %s", ShellQuote(parentDir))}
	execCfg := container.ExecOptions{
		Cmd:          mkdirCmd,
		AttachStdout: true,
		AttachStderr: true,
	}
	execResp, execErr := dm.client.ContainerExecCreate(dm.ctx, cid, execCfg)
	if execErr == nil {
		_ = dm.client.ContainerExecStart(dm.ctx, execResp.ID, container.ExecStartOptions{})
	}

	utils.Info("writing file to volume", zap.String("file", destPath))
	if err := dm.client.CopyToContainer(dm.ctx, cid, filepath.Dir(destPath), content, container.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("copy to container: %v", err)
	}

	return nil
}

// BackupVolume archives a Docker volume and streams the result to dst.
//
// It deliberately does NOT bind-mount a host directory for the output. A bind
// mount source is resolved by the Docker daemon on the HOST, while this process
// computes its paths inside its own container - so the archive landed on the
// host and the app could never find it again. Downloads 404'd and cloud uploads
// failed, while the backup was still recorded as successful.
//
// Instead the helper container writes the archive to its own filesystem and we
// copy it out over the Docker API, which involves no host paths at all.
func (dm *DockerManager) BackupVolume(volumeName, filename string, dst io.Writer) error {
	if err := dm.ensureAlpine(); err != nil {
		return err
	}

	const archivePath = "/tmp/backup.tar.gz"

	cmd := []string{"sh", "-c", fmt.Sprintf("tar czf %s -C /data .", ShellQuote(archivePath))}

	cc := &container.Config{Image: "alpine:latest", Cmd: cmd}
	hc := &container.HostConfig{Binds: []string{fmt.Sprintf("%s:/data", volumeName)}}

	resp, err := dm.client.ContainerCreate(dm.ctx, cc, hc, nil, nil, "")
	if err != nil {
		return fmt.Errorf("container create: %v", err)
	}
	cid := resp.ID
	defer func() { _ = dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true}) }()

	if err := dm.client.ContainerStart(dm.ctx, cid, container.StartOptions{}); err != nil {
		return fmt.Errorf("container start: %v", err)
	}

	waitCh, errCh := dm.client.ContainerWait(dm.ctx, cid, container.WaitConditionNotRunning)
	select {
	case e := <-errCh:
		return fmt.Errorf("container wait: %v", e)
	case <-waitCh:
	}

	// CopyFromContainer returns a tar stream wrapping the file, so unwrap the
	// single entry rather than writing the outer tar to disk.
	reader, _, err := dm.client.CopyFromContainer(dm.ctx, cid, archivePath)
	if err != nil {
		return fmt.Errorf("copy archive out of helper container: %w", err)
	}
	defer func() { _ = reader.Close() }()

	tr := tar.NewReader(reader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("helper container produced no archive")
		}
		if err != nil {
			return fmt.Errorf("read archive stream: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if _, err := io.Copy(dst, tr); err != nil {
			return fmt.Errorf("write backup archive: %w", err)
		}
		return nil
	}
}

// RestoreVolume extracts a tar.gz backup into a Docker volume.
// volumeName: the Docker volume name
// backupDir: the host path to the backup directory (mounted as /backup)
// RestoreVolume extracts a backup archive into a Docker volume.
//
// The archive is streamed IN over the Docker API rather than bind-mounting a
// host directory, for the same reason BackupVolume streams out: a bind source is
// resolved on the host, not inside this container, so the path would not exist.
func (dm *DockerManager) RestoreVolume(volumeName, filename string, src io.Reader) error {
	if err := dm.ensureAlpine(); err != nil {
		return err
	}

	const stagedPath = "/tmp/restore.tar.gz"

	// Sleep, so the container stays alive long enough to receive the upload
	// before the extraction command runs.
	cc := &container.Config{
		Image: "alpine:latest",
		Cmd:   []string{"sh", "-c", "sleep 3600"},
	}
	hc := &container.HostConfig{Binds: []string{fmt.Sprintf("%s:/data", volumeName)}}

	resp, err := dm.client.ContainerCreate(dm.ctx, cc, hc, nil, nil, "")
	if err != nil {
		return fmt.Errorf("container create: %v", err)
	}
	cid := resp.ID
	defer func() { _ = dm.client.ContainerRemove(dm.ctx, cid, container.RemoveOptions{Force: true}) }()

	if err := dm.client.ContainerStart(dm.ctx, cid, container.StartOptions{}); err != nil {
		return fmt.Errorf("container start: %v", err)
	}

	// CopyToContainer takes a tar stream, so wrap the archive as a single entry.
	buf, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("read backup archive: %w", err)
	}

	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	if err := tw.WriteHeader(&tar.Header{
		Name: "restore.tar.gz",
		Mode: 0o600,
		Size: int64(len(buf)),
	}); err != nil {
		return fmt.Errorf("build upload stream: %w", err)
	}
	if _, err := tw.Write(buf); err != nil {
		return fmt.Errorf("build upload stream: %w", err)
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("build upload stream: %w", err)
	}

	if err := dm.client.CopyToContainer(dm.ctx, cid, "/tmp", &tarBuf, container.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("upload archive to helper container: %w", err)
	}

	if err := dm.execInHelperContainer(dm.ctx, cid,
		fmt.Sprintf("tar xzf %s -C /data", ShellQuote(stagedPath))); err != nil {
		return fmt.Errorf("extract archive: %w", err)
	}

	return nil
}
