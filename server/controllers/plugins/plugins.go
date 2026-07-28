package plugins

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/docker_manager"
	"ark-server-commander/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func getServerPluginsVolume(serverIDStr string) (string, error) {
	id, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid server ID")
	}
	return utils.GetServerPluginsVolumeName(uint(id)), nil
}

func validateServerOwnership(c *gin.Context, serverIDStr string) error {
	userID := c.GetUint("user_id")
	id, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid server ID")
	}
	var count int64
	database.DB.Model(&models.Server{}).Where("id = ? AND user_id = ?", id, userID).Count(&count)
	if count == 0 {
		return fmt.Errorf("server not found")
	}
	return nil
}

// cleanPath normalizes a user supplied path into an absolute path rooted at "/".
// The leading slash is added before cleaning so that traversal segments are
// resolved away instead of being preserved; anything that still escapes the
// root afterwards is rejected.
func cleanPath(p string) (string, error) {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	clean := filepath.Clean(p)
	if !strings.HasPrefix(clean, "/") {
		return "", fmt.Errorf("invalid path")
	}
	for _, seg := range strings.Split(clean, "/") {
		if seg == ".." {
			return "", fmt.Errorf("invalid path")
		}
	}
	return clean, nil
}

// safeFileName reduces an upload supplied name to a single path component and
// rejects anything that could be used to escape the destination directory.
func safeFileName(name string) (string, error) {
	if strings.ContainsAny(name, "/\\") {
		name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	}
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("invalid file name")
	}
	return name, nil
}

func getDM() (*docker_manager.DockerManager, error) {
	dm, err := docker_manager.GetDockerManager()
	if err != nil {
		return nil, fmt.Errorf("Docker manager not available: %v", err)
	}
	return dm, nil
}

func ListFiles(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.DefaultQuery("path", "/")

	if serverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id is required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	volumePath := "/plugins" + cleaned

	files, err := dm.ListFiles(volumeName, "/plugins", volumePath)
	if err != nil {
		utils.Error("list files failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}
	if files == nil {
		files = []docker_manager.FileInfo{}
	}

	c.JSON(http.StatusOK, gin.H{"files": files, "path": cleaned, "serverId": serverID})
}

func UploadFile(c *gin.Context) {
	serverID := c.Query("server_id")
	destPath := c.DefaultQuery("path", "/")

	if serverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id is required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cleanedDest, err := cleanPath(destPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	var uploaded []string
	var extracted []string
	for _, fh := range files {
		name, err := safeFileName(fh.Filename)
		if err != nil {
			utils.Warn("upload rejected unsafe file name", zap.String("name", fh.Filename))
			continue
		}

		src, err := fh.Open()
		if err != nil {
			continue
		}

		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		hdr := &tar.Header{
			Name: name,
			Size: fh.Size,
			Mode: 0644,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			src.Close()
			continue
		}
		if _, err := io.Copy(tw, src); err != nil {
			utils.Error("copy upload into tar failed", zap.String("file", name), zap.Error(err))
			tw.Close()
			src.Close()
			continue
		}
		tw.Close()
		src.Close()

		dest := "/plugins" + cleanedDest + "/" + name
		if err := dm.WriteFileToVolume(volumeName, "/plugins", dest, &buf); err != nil {
			utils.Error("upload file failed", zap.String("name", name), zap.Error(err))
			continue
		}
		uploaded = append(uploaded, name)

		// Auto-extract .zip files
		if strings.HasSuffix(strings.ToLower(name), ".zip") {
			extractDir := "/plugins" + cleanedDest
			cmd := []string{"unzip", "-o", dest, "-d", extractDir}
			if _, err := dm.RunCommandInVolume(volumeName, "/plugins", cmd); err != nil {
				utils.Error("auto-extract zip failed", zap.String("name", name), zap.Error(err))
			} else {
				extracted = append(extracted, name)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("%d file(s) uploaded", len(uploaded)),
		"uploaded":  uploaded,
		"extracted": extracted,
	})
}

func DeleteFile(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")

	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id and path are required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mount := "/plugins"
	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cleaned == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete the plugins root"})
		return
	}
	fullPath := mount + cleaned
	cmd := []string{"rm", "-rf", fullPath}
	if _, err := dm.RunCommandInVolume(volumeName, mount, cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func RenameFile(c *gin.Context) {
	serverID := c.Query("server_id")
	oldPath := c.Query("old_path")
	newPath := c.Query("new_path")

	if serverID == "" || oldPath == "" || newPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id, old_path, new_path required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mount := "/plugins"
	cleanedOld, err := cleanPath(oldPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cleanedNew, err := cleanPath(newPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if cleanedOld == "/" || cleanedNew == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot rename the plugins root"})
		return
	}
	fullOld := mount + cleanedOld
	fullNew := mount + cleanedNew

	cmd := []string{"sh", "-c", fmt.Sprintf(
		"mkdir -p %s && mv %s %s",
		docker_manager.ShellQuote(filepath.Dir(fullNew)),
		docker_manager.ShellQuote(fullOld),
		docker_manager.ShellQuote(fullNew),
	)}
	if _, err := dm.RunCommandInVolume(volumeName, mount, cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rename failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "renamed"})
}

func CreateDir(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")

	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id and path required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mount := "/plugins"
	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := []string{"mkdir", "-p", mount + cleaned}
	if _, err := dm.RunCommandInVolume(volumeName, mount, cmd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mkdir failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "directory created"})
}

func ReadFile(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")

	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id and path required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mount := "/plugins"
	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reader, err := dm.ReadFileFromVolume(volumeName, mount, mount+cleaned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
		return
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)
	_, err = tarReader.Next()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file not found"})
		return
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, tarReader); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
		return
	}
	c.String(http.StatusOK, buf.String())
}

func WriteFile(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")

	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id and path required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	mount := "/plugins"
	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dest := mount + cleaned

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name: filepath.Base(cleaned),
		Size: int64(len(req.Content)),
		Mode: 0644,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tar error"})
		return
	}
	if _, err := tw.Write([]byte(req.Content)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "tar error"})
		return
	}
	tw.Close()

	if err := dm.WriteFileToVolume(volumeName, mount, dest, &buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "saved"})
}

func UnzipFile(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")

	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id and path required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mount := "/plugins"
	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fullPath := mount + cleaned
	extractDir := filepath.Dir(fullPath)

	cmd := []string{"unzip", "-o", fullPath, "-d", extractDir}
	if _, err := dm.RunCommandInVolume(volumeName, mount, cmd); err != nil {
		utils.Error("unzip failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unzip failed"})
		return
	}

	// Remove the zip file after extraction
	rmCmd := []string{"rm", "-f", fullPath}
	_, _ = dm.RunCommandInVolume(volumeName, mount, rmCmd)

	c.JSON(http.StatusOK, gin.H{"message": "extracted"})
}

func ZipDownload(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")

	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id and path required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mount := "/plugins"
	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	baseName := filepath.Base(cleaned)
	if baseName == "" || baseName == "." || baseName == "/" {
		baseName = "plugins"
	}
	zipName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".zip"
	tempZip := mount + "/.__ark_temp_zip_" + zipName

	// Install zip, create archive, then stream it out via tar
	mountDir := filepath.Dir(mount + cleaned)
	createCmd := []string{"sh", "-c",
		fmt.Sprintf("apk add --no-cache zip unzip >/dev/null 2>&1 && cd %s && zip -r %s . >/dev/null 2>&1",
			docker_manager.ShellQuote(mountDir), docker_manager.ShellQuote(tempZip))}
	if _, err := dm.RunCommandInVolume(volumeName, mount, createCmd); err != nil {
		utils.Error("zip creation failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "zip creation failed"})
		return
	}

	// Read the zip file using existing ReadFileFromVolume (returns tar of the file)
	reader, err := dm.ReadFileFromVolume(volumeName, mount, tempZip)
	if err != nil {
		utils.Error("read temp zip failed", zap.Error(err))
		_, _ = dm.RunCommandInVolume(volumeName, mount, []string{"rm", "-f", tempZip})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read zip failed"})
		return
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)
	if _, err := tarReader.Next(); err != nil {
		_, _ = dm.RunCommandInVolume(volumeName, mount, []string{"rm", "-f", tempZip})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "zip not found in tar"})
		return
	}

	// Clean up temp zip
	_, _ = dm.RunCommandInVolume(volumeName, mount, []string{"rm", "-f", tempZip})

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, zipName))
	c.Header("Content-Type", "application/zip")
	if _, err := io.Copy(c.Writer, tarReader); err != nil {
		utils.Error("stream download to client failed", zap.Error(err))
	}
}

func DownloadFile(c *gin.Context) {
	serverID := c.Query("server_id")
	path := c.Query("path")

	if serverID == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "server_id and path required"})
		return
	}
	if err := validateServerOwnership(c, serverID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	volumeName, err := getServerPluginsVolume(serverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dm, err := getDM()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	mount := "/plugins"
	cleaned, err := cleanPath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reader, err := dm.ReadFileFromVolume(volumeName, mount, mount+cleaned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file failed"})
		return
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)
	_, err = tarReader.Next()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file not found in tar"})
		return
	}

	fileName := filepath.Base(cleaned)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.Header("Content-Type", "application/octet-stream")
	if _, err := io.Copy(c.Writer, tarReader); err != nil {
		utils.Error("stream download to client failed", zap.Error(err))
	}
}
