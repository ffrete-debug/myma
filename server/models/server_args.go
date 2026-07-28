package models

import (
	"fmt"
	"sort"
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
// GenerateArgsString builds the launch arguments with no mods attached.
func (sa *ServerArgs) GenerateArgsString(server Server) string {
	return sa.GenerateArgsStringWithMods(server, nil)
}

// GenerateArgsStringWithMods builds the launch arguments including the server's
// enabled Steam Workshop mods, in load order.
//
// modIDs is passed in rather than looked up here because models must not depend
// on the service layer. Callers that create or inspect a container resolve the
// list first; callers that only need the base arguments pass nil.
//
// The IDs are validated at the service boundary (decimal only) — they are
// interpolated into a launch string, so nothing else may reach this.
func (sa *ServerArgs) GenerateArgsStringWithMods(server Server, modIDs []string) string {
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

	// SessionName is deliberately NOT emitted here. The game image runs
	//     SERVER_CMD="$PROTON run ShooterGameServer.exe ${SERVER_ARGS}"; $SERVER_CMD
	// with SERVER_ARGS expanded UNQUOTED, so a session name containing a space
	// word-splits the launch string: the name is truncated at the first space
	// and every query parameter ordered after it is silently dropped.
	// SessionName is written to GameUserSettings.ini [SessionSettings] instead,
	// which has no quoting problem. See utils.GetDefaultGameUserSettings.

	if server.GameModIds != "" {
		queryParams = append(queryParams, fmt.Sprintf("?GameModIds=%s", server.GameModIds))
	}

	// Sorted: Go randomises map iteration, which made SERVER_ARGS differ on every
	// call. That defeated the container drift check (it compared unequal strings
	// and rebuilt the container on every start) and made which parameters
	// survived a word-split non-deterministic.
	for _, key := range sortedKeys(sa.QueryParams) {
		value := sa.QueryParams[key]
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

	for _, key := range sortedAnyKeys(sa.CommandLineArgs) {
		value := sa.CommandLineArgs[key]
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

	// NOTE: no -mods= flag. That is an ASA/CurseForge option and ASE ignores it
	// (docs/wiki.md:358, "inASE = No"). This image installs Steam Workshop mods
	// itself from the GameModIds environment variable, so the mod list travels
	// via server.GameModIds and the ?GameModIds= query parameter above.
	// modIDs is retained in the signature so callers can keep passing the
	// resolved list; it is used only for the query parameter fallback below.
	if len(modIDs) > 0 && server.GameModIds == "" {
		queryParams = append(queryParams, fmt.Sprintf("?GameModIds=%s", strings.Join(modIDs, ",")))
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

// sortedKeys returns map keys in a stable order.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// sortedAnyKeys is sortedKeys for the heterogeneous command-line map.
func sortedAnyKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
