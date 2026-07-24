package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	// Configuration file name constants
	GameUserSettingsFileName = "GameUserSettings.ini"
	GameIniFileName          = "Game.ini"

	// Configuration directory - adapted for new ASE server image path
	ConfigDirectory = "Config/WindowsServer"
)

// GetDefaultGameUserSettings returns the default GameUserSettings.ini configuration
func GetDefaultGameUserSettings(serverName, mapName string, maxPlayers int) string {
	return fmt.Sprintf(`[ServerSettings]
SessionName=%s
ServerPassword=
MaxPlayers=%d

[SessionSettings]
SessionName=%s

[MessageOfTheDay]
Message=Welcome to %s ARK Server!

[/Script/Engine.GameSession]
MaxPlayers=%d
`, serverName, maxPlayers, serverName, serverName, maxPlayers)
}

// GetDefaultGameIni returns the default Game.ini configuration
func GetDefaultGameIni() string {
	return `[/script/shootergame.shootergamemode]
bUseSingleplayerSettings=false
bDisableStructurePlacementCollision=false
bAllowFlyerCarryPvE=true
bDisableStructureDecayPvE=false
bAllowUnlimitedRespecs=true
bPassiveDefensesDamageRiderlessDinos=true
MaxNumberOfPlayersInTribe=0

[/Script/ShooterGame.ShooterGameMode]
DifficultyOffset=1.0
OverrideOfficialDifficulty=5.0

# Resource respawn rate
ResourcesRespawnPeriodMultiplier=1.0

# Taming related settings
TamingSpeedMultiplier=1.0
DinoCharacterFoodDrainMultiplier=1.0
DinoCharacterStaminaDrainMultiplier=1.0
DinoCharacterHealthRecoveryMultiplier=1.0
DinoCountMultiplier=1.0

# Experience rate
XPMultiplier=1.0
PlayerCharacterWaterDrainMultiplier=1.0
PlayerCharacterFoodDrainMultiplier=1.0
PlayerCharacterStaminaDrainMultiplier=1.0
PlayerCharacterHealthRecoveryMultiplier=1.0

# Harvest rate
HarvestAmountMultiplier=1.0
HarvestHealthMultiplier=1.0

# Day/night cycle speed
DayCycleSpeedScale=1.0
NightTimeSpeedScale=1.0

# Structure settings
StructureResistanceMultiplier=1.0
StructureDamageMultiplier=1.0
StructureDamageRepairCooldown=180

PvEStructureDecayPeriodMultiplier=1.0

# PvP related settings
bPvEDisableFriendlyFire=False
bEnablePvPGamma=False
bDisableFriendlyFire=False
bAllowFlyerCarryPvE=True
`
}

// ValidateINIContent validates basic INI content format
func ValidateINIContent(content string) error {
	if content == "" {
		return nil // Empty content is valid
	}

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check section format
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return fmt.Errorf("Line %d: section format error, missing closing bracket", i+1)
			}
			continue
		}

		// Check key-value pair format
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
				return fmt.Errorf("Line %d: key-value pair format error", i+1)
			}
			continue
		}
	}

	return nil
}

// GetServerConfigPath returns the server configuration directory path
func GetServerConfigPath(serverID uint) string {
	serverFolder := GetServerFolderPath(serverID)
	return filepath.Join(serverFolder, ConfigDirectory)
}

// GetConfigFilePath returns the full path to a configuration file
func GetConfigFilePath(serverID uint, fileName string) string {
	configPath := GetServerConfigPath(serverID)
	return filepath.Join(configPath, fileName)
}

// GenerateGameModIdsEnv generates the GameModIds environment variable
// Returns empty string if GameModIds is empty
func GenerateGameModIdsEnv(gameModIds string) string {
	if gameModIds == "" {
		return ""
	}
	return gameModIds
}
