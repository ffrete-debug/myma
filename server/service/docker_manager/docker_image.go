package docker_manager

import (
	"ark-server-commander/utils"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"go.uber.org/zap"
)

// ImagePullProgress 
type ImagePullProgress struct {
	Status         string `json:"status"`
	Progress       string `json:"progress"`
	ProgressDetail *struct {
		Current int64 `json:"current"`
		Total   int64 `json:"total"`
	} `json:"progressDetail"`
	ID string `json:"id"`
}

// LayerStatus Status
type LayerStatus struct {
	ID       string `json:"id"`       // ID
	Size     int64  `json:"size"`     // 
	Progress int64  `json:"progress"` // 
	Status   string `json:"status"`   // Status：downloading, extracting, verifying, complete
}

// ImageStatus Image status info
type ImageStatus struct {
	Exists       bool                    `json:"exists"`        // YesNo
	Pulling      bool                    `json:"pulling"`       // YesNo
	Ready        bool                    `json:"ready"`         // YesNo
	Error        string                  `json:"error"`         // Error
	CurrentLayer string                  `json:"current_layer"` // 
	Layers       map[string]*LayerStatus `json:"layers"`        // Status
}

// imagePullState Status
type imagePullState struct {
	pulling      bool
	currentLayer string
	layers       map[string]*LayerStatus
	mu           sync.RWMutex
}

// Status（）
var imagePullStates = &sync.Map{} // map[string]*imagePullState

func getImagePullState(imageName string) *imagePullState {
	if v, ok := imagePullStates.Load(imageName); ok {
		return v.(*imagePullState)
	}
	state := &imagePullState{layers: make(map[string]*LayerStatus)}
	imagePullStates.Store(imageName, state)
	return state
}

func cleanupImagePullState(imageName string) {
	imagePullStates.Delete(imageName)
}

// PullImageWithProgress Docker
// imageName: Image name
// : Error
func (dm *DockerManager) PullImageWithProgress(imageName string) error {
	utils.Info("On Docker ", zap.String("image", imageName))

	state := getImagePullState(imageName)
	state.mu.Lock()
	state.pulling = true
	state.currentLayer = ""
	state.layers = make(map[string]*LayerStatus)
	state.mu.Unlock()

	// Status
	defer func() {
		state := getImagePullState(imageName)
		state.mu.Lock()
		state.pulling = false
		state.currentLayer = ""
		state.layers = make(map[string]*LayerStatus)
		state.mu.Unlock()
		cleanupImagePullState(imageName)
	}()

	// 
	reader, err := dm.client.ImagePull(dm.ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf(" Docker : %v", err)
	}
	defer reader.Close()

	// 
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			// JSON
			progress := string(buffer[:n])
			lines := strings.Split(progress, "\n")

			for _, line := range lines {
				if line == "" {
					continue
				}

				var progressInfo ImagePullProgress
				if err := json.Unmarshal([]byte(line), &progressInfo); err == nil {
					state := getImagePullState(imageName)
					state.mu.Lock()

					// 
					layerID := progressInfo.ID
					if layerID != "" {
						state.currentLayer = layerID

						// 
						if state.layers[layerID] == nil {
							state.layers[layerID] = &LayerStatus{
								ID:       layerID,
								Size:     0,
								Progress: 0,
								Status:   "pending",
							}
						}

						// 
						if progressInfo.ProgressDetail != nil {
							if progressInfo.ProgressDetail.Total > 0 {
								state.layers[layerID].Size = progressInfo.ProgressDetail.Total
							}
							if progressInfo.ProgressDetail.Current > 0 {
								state.layers[layerID].Progress = progressInfo.ProgressDetail.Current
							}
						}

						// ProgressDetail，Progress
						if state.layers[layerID].Size == 0 && progressInfo.Progress != "" {
							if size := parseSizeFromProgress(progressInfo.Progress); size > 0 {
								state.layers[layerID].Size = size
							}
						}

						// Status
						if strings.Contains(progressInfo.Status, "Downloading") {
							state.layers[layerID].Status = "downloading"
						} else if strings.Contains(progressInfo.Status, "Extracting") {
							state.layers[layerID].Status = "extracting"
						} else if strings.Contains(progressInfo.Status, "Verifying") {
							state.layers[layerID].Status = "verifying"
						} else if strings.Contains(progressInfo.Status, "Pull complete") || strings.Contains(progressInfo.Status, "complete") {
							state.layers[layerID].Status = "complete"
							// ，，Settings
							if state.layers[layerID].Size == 0 {
								state.layers[layerID].Size = state.layers[layerID].Progress
								if state.layers[layerID].Size == 0 {
									state.layers[layerID].Size = 1024 * 1024 // 1MB
								}
							}
							state.layers[layerID].Progress = state.layers[layerID].Size
						}
					}

					state.mu.Unlock()

					// 
					if strings.Contains(progressInfo.Status, "Downloading") || strings.Contains(progressInfo.Status, "Extracting") {
						// （）
					}
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf(" : %v", err)
		}
	}

	utils.Info("Docker Success", zap.String("image", imageName))

	return nil
}

// ImageExists DockerYesNo
// imageName: Image name
// : YesNoError
func (dm *DockerManager) ImageExists(imageName string) (bool, error) {
	// Image information
	_, err := dm.client.ImageInspect(dm.ctx, imageName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil // 
		}
		return false, fmt.Errorf(" Docker : %v", err)
	}

	return true, nil
}

// GetImageStatus Get image status
// imageName: Image name
// : Image status info
func (dm *DockerManager) GetImageStatus(imageName string) *ImageStatus {
	status := &ImageStatus{
		Exists:       false,
		Pulling:      false,
		Ready:        false,
		Error:        "",
		CurrentLayer: "",
		Layers:       make(map[string]*LayerStatus),
	}

	// YesNo
	exists, err := dm.ImageExists(imageName)
	if err != nil {
		status.Error = fmt.Sprintf(" : %v", err)
		return status
	}

	status.Exists = exists
	if exists {
		status.Ready = true
		return status
	}

	// YesNo
	state := getImagePullState(imageName)
	state.mu.RLock()
	status.Pulling = state.pulling
	if status.Pulling {
		status.CurrentLayer = state.currentLayer

		// Status
		if layers := state.layers; layers != nil {
			for layerID, layerStatus := range layers {
				status.Layers[layerID] = &LayerStatus{
					ID:       layerStatus.ID,
					Size:     layerStatus.Size,
					Progress: layerStatus.Progress,
					Status:   layerStatus.Status,
				}
			}
		}
	}
	state.mu.RUnlock()

	return status
}

// IsImagePulling YesNo
func IsImagePulling(imageName string) bool {
	state := getImagePullState(imageName)
	state.mu.RLock()
	pulling := state.pulling
	state.mu.RUnlock()
	return pulling
}

// WaitForImage （， GetImageStatus）
// imageName: Image name
// timeout: （）
// : YesNoSuccessError
func (dm *DockerManager) WaitForImage(imageName string, timeout int) (bool, error) {
	utils.Info(" ", zap.String("image", imageName))

	// YesNo
	for i := 0; i < timeout; i++ {
		exists, err := dm.ImageExists(imageName)
		if err != nil {
			return false, fmt.Errorf(" : %v", err)
		}

		if exists {
			utils.Info(" ", zap.String("image", imageName))
			return true, nil
		}

		// YesNo
		state := getImagePullState(imageName)
		state.mu.RLock()
		pulling := state.pulling
		state.mu.RUnlock()
		if pulling {
			utils.Debug(" ， ", zap.String("image", imageName))
		}

		// 1
		time.Sleep(1 * time.Second)
	}

	return false, fmt.Errorf("  %s  （%d ）", imageName, timeout)
}

// parseSizeFromProgress 
// progress: ， "1.5MB/2.0MB"
// : （）
func parseSizeFromProgress(progress string) int64 {
	// 
	progress = strings.TrimSpace(progress)

	//  "/" 
	parts := strings.Split(progress, "/")
	if len(parts) != 2 {
		return 0
	}

	// 
	totalStr := strings.TrimSpace(parts[1])
	return parseSizeString(totalStr)
}

// Check ImageUpdate YesNo
// imageName: Image name
// : YesNoError
func (dm *DockerManager) CheckImageUpdate(imageName string) (bool, error) {
	// （ RepoDigests）
	imageInspect, err := dm.client.ImageInspect(dm.ctx, imageName)
	if err != nil {
		// ，（）
		if errdefs.IsNotFound(err) {
			return true, nil
		}
		return false, fmt.Errorf(" Image information : %v", err)
	}

	// Use Docker Distribution API to inspect remote manifest digest
	distInspect, err := dm.client.DistributionInspect(dm.ctx, imageName, "")
	if err != nil {
		// If registry unreachable, assume no update
		utils.Warn("failed to inspect remote manifest", zap.String("image", imageName), zap.Error(err))
		return false, nil
	}

	remoteDigest := distInspect.Descriptor.Digest.String()

	// Local digest comes from RepoDigests (e.g., "tbro98/ase-server@sha256:...")
	localDigest := ""
	if len(imageInspect.RepoDigests) > 0 {
		parts := strings.SplitN(imageInspect.RepoDigests[0], "@", 2)
		if len(parts) == 2 {
			localDigest = parts[1]
		}
	}

	// Fallback: compare image ID (which is also a digest)
	if localDigest == "" {
		localDigest = imageInspect.ID
	}

	hasUpdate := localDigest != remoteDigest
	if hasUpdate {
		utils.Info("image update available", zap.String("image", imageName),
			zap.String("local", localDigest), zap.String("remote", remoteDigest))
	}
	return hasUpdate, nil
}

// GetImageInfo 
// imageName: Image name
// : Image informationError
func (dm *DockerManager) GetImageInfo(imageName string) (*ImageInfo, error) {
	imageInspect, err := dm.client.ImageInspect(dm.ctx, imageName)
	if err != nil {
		return nil, err
	}

	return &ImageInfo{
		ID:      imageInspect.ID,
		Tags:    imageInspect.RepoTags,
		Size:    imageInspect.Size,
		Created: imageInspect.Created,
	}, nil
}

// ImageInfo Image information
type ImageInfo struct {
	ID      string   `json:"id"`      // ID
	Tags    []string `json:"tags"`    // 
	Size    int64    `json:"size"`    // 
	Created string   `json:"created"` // Created at
}

// RemoveOldImage Delete
// imageName: Image name
// keepLatest: YesNo
// : Error
func (dm *DockerManager) RemoveOldImage(imageName string, keepLatest bool) error {
	if !keepLatest {
		// Delete
		_, err := dm.client.ImageRemove(dm.ctx, imageName, image.RemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			return fmt.Errorf("Delete : %v", err)
		}
		utils.Info(" Delete", zap.String("image", imageName))
	} else {
		utils.Debug(" ， Delete")
	}

	return nil
}

// GetContainersByImage 
// imageName: Image name
// : Error
func (dm *DockerManager) GetContainersByImage(imageName string) ([]ContainerInfo, error) {
	containers, err := dm.client.ContainerList(dm.ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf(" : %v", err)
	}

	var result []ContainerInfo
	for _, c := range containers {
		if c.Image == imageName {
			result = append(result, ContainerInfo{
				ID:     c.ID,
				Name:   c.Names[0],
				Image:  c.Image,
				Status: c.Status,
				State:  c.State,
			})
		}
	}

	return result, nil
}

// ContainerInfo 
type ContainerInfo struct {
	ID     string `json:"id"`     // ID
	Name   string `json:"name"`   // 
	Image  string `json:"image"`  // Image name
	Status string `json:"status"` // Status
	State  string `json:"state"`  // Status
}

// GetImageHistory 
// imageName: Image name
// : Error
func (dm *DockerManager) GetImageHistory(imageName string) ([]ImageHistoryEntry, error) {
	history, err := dm.client.ImageHistory(dm.ctx, imageName)
	if err != nil {
		return nil, fmt.Errorf(" : %v", err)
	}

	var result []ImageHistoryEntry
	for _, h := range history {
		result = append(result, ImageHistoryEntry{
			ID:        h.ID,
			Created:   h.Created,
			CreatedBy: h.CreatedBy,
			Size:      h.Size,
			Comment:   h.Comment,
		})
	}

	return result, nil
}

// ImageHistoryEntry 
type ImageHistoryEntry struct {
	ID        string `json:"id"`         // ID
	Created   int64  `json:"created"`    // Created at
	CreatedBy string `json:"created_by"` // Create
	Size      int64  `json:"size"`       // 
	Comment   string `json:"comment"`    // 
}

// parseSizeString 
// sizeStr: ， "2.0MB", "1.5GB"
// : 
func parseSizeString(sizeStr string) int64 {
	sizeStr = strings.ToLower(strings.TrimSpace(sizeStr))

	// 
	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "kb") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "kb")
	} else if strings.HasSuffix(sizeStr, "mb") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "mb")
	} else if strings.HasSuffix(sizeStr, "gb") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "gb")
	} else if strings.HasSuffix(sizeStr, "b") {
		multiplier = 1
		sizeStr = strings.TrimSuffix(sizeStr, "b")
	}

	// 
	if size, err := strconv.ParseFloat(sizeStr, 64); err == nil {
		return int64(size * float64(multiplier))
	}
	return 0
}
