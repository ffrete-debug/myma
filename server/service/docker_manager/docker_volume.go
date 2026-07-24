package docker_manager

import (
	"ark-server-commander/utils"
	"fmt"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/volume"
	"go.uber.org/zap"
)

// CreateVolume CreateDocker（Plugins）
// serverID: Server ID
// : Error
func (dm *DockerManager) CreateVolume(serverID uint) (string, error) {
	volumeName := utils.GetServerVolumeName(serverID)
	pluginsVolumeName := utils.GetServerPluginsVolumeName(serverID)

	// Create
	if err := dm.createSingleVolume(volumeName); err != nil {
		return "", fmt.Errorf("Create : %v", err)
	}

	// CreatePlugins
	if err := dm.createSingleVolume(pluginsVolumeName); err != nil {
		// PluginsVolume creation failed，Create
		dm.RemoveVolume(serverID)
		return "", fmt.Errorf("CreatePlugins : %v", err)
	}

	utils.Info("Docker Created successfully", zap.String("data_volume", volumeName), zap.String("plugins_volume", pluginsVolumeName))
	return volumeName, nil
}

// createSingleVolume CreateDocker
// volumeName: 
// : Error
func (dm *DockerManager) createSingleVolume(volumeName string) error {
	// YesNo
	exists, err := dm.VolumeExists(volumeName)
	if err != nil {
		return fmt.Errorf(" YesNo : %v", err)
	}
	if exists {
		utils.Debug("DockerVolume already exists， Create", zap.String("volume", volumeName))
		return nil
	}

	// Create
	utils.Infof(" CreateDocker : %s", volumeName)
	volumeCreateBody := volume.CreateOptions{
		Name: volumeName,
	}

	_, err = dm.client.VolumeCreate(dm.ctx, volumeCreateBody)
	if err != nil {
		return fmt.Errorf("CreateDocker : %v", err)
	}

	utils.Info("Docker Created successfully", zap.String("volume", volumeName))
	return nil
}

// RemoveVolume DeleteDocker（Plugins）
// serverID: Server ID
// : Error
func (dm *DockerManager) RemoveVolume(serverID uint) error {
	volumeName := utils.GetServerVolumeName(serverID)
	pluginsVolumeName := utils.GetServerPluginsVolumeName(serverID)

	// Delete
	if err := dm.removeSingleVolume(volumeName); err != nil {
		return err
	}

	// DeletePlugins
	if err := dm.removeSingleVolume(pluginsVolumeName); err != nil {
		// PluginsVolume deletion failed，
		utils.Warn("DeletePlugins ", zap.String("volume", pluginsVolumeName), zap.Error(err))
	}

	return nil
}

// removeSingleVolume DeleteDocker
// volumeName: 
// : Error
func (dm *DockerManager) removeSingleVolume(volumeName string) error {
	// YesNo
	exists, err := dm.VolumeExists(volumeName)
	if err != nil {
		return fmt.Errorf(" YesNo : %v", err)
	}
	if !exists {
		utils.Debug("Docker ， Delete", zap.String("volume", volumeName))
		return nil
	}

	utils.Infof(" DeleteDocker : %s", volumeName)
	err = dm.client.VolumeRemove(dm.ctx, volumeName, false)
	if err != nil {
		return fmt.Errorf("DeleteDocker : %v", err)
	}

	utils.Info("Docker DeleteSuccess", zap.String("volume", volumeName))
	return nil
}

// VolumeExists DockerYesNo
// volumeName: 
// : YesNoError
func (dm *DockerManager) VolumeExists(volumeName string) (bool, error) {
	// 
	_, err := dm.client.VolumeInspect(dm.ctx, volumeName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil // 
		}
		return false, fmt.Errorf(" Docker : %v", err)
	}

	return true, nil
}
