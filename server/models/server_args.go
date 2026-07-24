package models

import (
	"fmt"
	"strings"
)

// ServerArgs ARKServersStart
type ServerArgs struct {
	// （?On，）
	QueryParams map[string]string `json:"query_params" gorm:"-"`

	// （-On）
	CommandLineArgs map[string]interface{} `json:"command_line_args" gorm:"-"`

	// （User）
	CustomArgs []string `json:"custom_args" gorm:"-"`
}

// ServerArgsRequest Start
type ServerArgsRequest struct {
	QueryParams     map[string]string      `json:"query_params"`
	CommandLineArgs map[string]interface{} `json:"command_line_args"`
	CustomArgs      []string               `json:"custom_args"`
}

// ServerArgsResponse Start
type ServerArgsResponse struct {
	QueryParams     map[string]string      `json:"query_params"`
	CommandLineArgs map[string]interface{} `json:"command_line_args"`
	CustomArgs      []string               `json:"custom_args"`
	GeneratedArgs   string                 `json:"generated_args"` // Start
}

// NewServerArgs CreateStart
func NewServerArgs() *ServerArgs {
	return &ServerArgs{
		QueryParams:     make(map[string]string),
		CommandLineArgs: make(map[string]interface{}),
		CustomArgs:      []string{},
	}
}

// FromServer ServerCreateServerArgs
func FromServer(server Server) *ServerArgs {
	args := NewServerArgs()

	// Settings（，Server）
	args.QueryParams["listen"] = ""

	// Settings
	args.CommandLineArgs["NoBattlEye"] = true
	args.CommandLineArgs["servergamelog"] = true
	args.CommandLineArgs["structurememopts"] = true
	args.CommandLineArgs["UseStructureStasisGrid"] = true
	args.CommandLineArgs["SecureSendArKPayload"] = true
	args.CommandLineArgs["UseItemDupeCheck"] = true
	args.CommandLineArgs["UseSecureSpawnRules"] = true
	args.CommandLineArgs["nosteamclient"] = true
	args.CommandLineArgs["game"] = true
	args.CommandLineArgs["server"] = true
	args.CommandLineArgs["log"] = true
	args.CommandLineArgs["MinimumTimeBetweenInventoryRetrieval"] = 3600
	args.CommandLineArgs["newsaveformat"] = true
	args.CommandLineArgs["usestore"] = true
	args.CommandLineArgs["BackupTransferPlayerDatas"] = true
	args.CommandLineArgs["converttostore"] = true

	return args
}

// GenerateArgsString Start
// Servers：、Query Port、RCON Port、Password、Map、ID
// Start：
func (sa *ServerArgs) GenerateArgsString(server Server) string {
	var queryParams []string
	var commandLineParams []string

	// Map（Servers）
	result := server.Map

	// （Servers，StartSettings）
	queryParams = append(queryParams, "?listen")
	queryParams = append(queryParams, fmt.Sprintf("?Port=%d", server.Port))
	queryParams = append(queryParams, fmt.Sprintf("?QueryPort=%d", server.QueryPort))
	queryParams = append(queryParams, fmt.Sprintf("?MaxPlayers=%d", server.MaxPlayers))
	queryParams = append(queryParams, "?RCONEnabled=True")
	queryParams = append(queryParams, fmt.Sprintf("?RCONPort=%d", server.RCONPort))
	queryParams = append(queryParams, fmt.Sprintf("?ServerAdminPassword=%s", server.AdminPassword))

	// Servers（SessionName）
	if server.SessionName != "" {
		queryParams = append(queryParams, fmt.Sprintf("?SessionName=%s", server.SessionName))
	}

	if server.GameModIds != "" {
		queryParams = append(queryParams, fmt.Sprintf("?GameModIds=%s", server.GameModIds))
	}

	// 
	for key, value := range sa.QueryParams {
		// ，
		if key == "listen" || key == "Port" || key == "QueryPort" || key == "MaxPlayers" ||
			key == "RCONEnabled" || key == "RCONPort" || key == "ServerAdminPassword" || key == "GameModIds" {
			continue
		}

		// "False"，
		if value == "" || strings.ToLower(value) == "false" {
			continue
		}

		queryParams = append(queryParams, fmt.Sprintf("?%s=%s", key, value))
	}

	// 
	for key, value := range sa.CommandLineArgs {
		switch v := value.(type) {
		case bool:
			if v {
				commandLineParams = append(commandLineParams, fmt.Sprintf("-%s", key))
			}
		case string:
			if v != "" {
				commandLineParams = append(commandLineParams, fmt.Sprintf("-%s=%s", key, v))
			}
			// ，
		case int, int32, int64, float32, float64:
			// YesNo0，Yes0
			if v != 0 {
				commandLineParams = append(commandLineParams, fmt.Sprintf("-%s=%v", key, v))
			}
		default:
			commandLineParams = append(commandLineParams, fmt.Sprintf("-%s=%v", key, v))
		}
	}

	// ID（ClusterID）
	if server.ClusterID != "" {
		commandLineParams = append(commandLineParams, fmt.Sprintf("-clusterid=%s", server.ClusterID))
	}

	// 
	commandLineParams = append(commandLineParams, sa.CustomArgs...)

	// ：Map + （None）+  + 
	if len(queryParams) > 0 {
		result += strings.Join(queryParams, "")
	}

	if len(commandLineParams) > 0 {
		if len(queryParams) > 0 {
			result += " " // 
		}
		result += strings.Join(commandLineParams, " ")
	}

	return result
}
