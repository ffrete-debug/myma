package utils

import "fmt"

// GetServerContainerName returns the server container name
// serverID: server ID
// Returns: container name
func GetServerContainerName(serverID uint) string {
	return fmt.Sprintf("ase-server-%d", serverID)
}

// GetServerVolumeName returns the server volume name
// serverID: server ID
// Returns: volume name
func GetServerVolumeName(serverID uint) string {
	return fmt.Sprintf("ase-server-%d", serverID)
}

// GetServerPluginsVolumeName returns the server plugins volume name
// serverID: server ID
// Returns: plugins volume name
func GetServerPluginsVolumeName(serverID uint) string {
	return fmt.Sprintf("ase-server-plugins-%d", serverID)
}
