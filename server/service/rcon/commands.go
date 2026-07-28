package rcon

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// This file builds RCON command strings for the UI's admin actions.
//
// Commands are assembled here rather than in the browser on purpose. ARK's RCON
// protocol is line-oriented, so a newline inside a player name or a broadcast
// message would end the intended command and start another one — letting a
// caller who can broadcast also run, say, DestroyWildDinos. Building the string
// server-side from validated parts is what stops that.

// steamIDPattern matches a SteamID64. ARK also accepts its internal player id,
// which is likewise decimal, so one numeric rule covers both.
var steamIDPattern = regexp.MustCompile(`^[0-9]{1,20}$`)

// controlChars matches anything that could terminate or split a command:
// CR, LF, NUL and the rest of the C0 range, plus DEL.
var controlChars = regexp.MustCompile(`[\x00-\x1f\x7f]`)

// maxBroadcastLen bounds a broadcast. ARK truncates long messages itself; this
// stops a caller pushing a megabyte through the RCON socket.
const maxBroadcastLen = 256

// Action is a supported admin operation.
type Action string

const (
	ActionBroadcast        Action = "broadcast"
	ActionSaveWorld        Action = "saveworld"
	ActionListPlayers      Action = "listplayers"
	ActionKickPlayer       Action = "kick"
	ActionBanPlayer        Action = "ban"
	ActionUnbanPlayer      Action = "unban"
	ActionDestroyWildDinos Action = "destroywilddinos"
	ActionSetTimeOfDay     Action = "settimeofday"
	ActionDoExit           Action = "doexit"
)

// Destructive reports whether an action changes or ends world state in a way
// the user should confirm first. The UI uses this to gate a confirmation step;
// it is advisory, not an authorisation decision.
func (a Action) Destructive() bool {
	switch a {
	case ActionDestroyWildDinos, ActionDoExit, ActionBanPlayer:
		return true
	default:
		return false
	}
}

// timeOfDayPattern matches "HH:MM" in 24-hour form.
var timeOfDayPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// BuildCommand turns a validated action plus parameters into an RCON command.
//
// Every branch either produces a command built from validated components or an
// error. No caller-supplied string is ever passed through unchecked.
func BuildCommand(action Action, params map[string]string) (string, error) {
	switch action {
	case ActionSaveWorld:
		return "SaveWorld", nil

	case ActionListPlayers:
		return "ListPlayers", nil

	case ActionDestroyWildDinos:
		return "DestroyWildDinos", nil

	case ActionDoExit:
		return "DoExit", nil

	case ActionBroadcast:
		msg, err := cleanMessage(params["message"])
		if err != nil {
			return "", err
		}
		return "Broadcast " + msg, nil

	case ActionKickPlayer:
		id, err := cleanSteamID(params["steam_id"])
		if err != nil {
			return "", err
		}
		return "KickPlayer " + id, nil

	case ActionBanPlayer:
		id, err := cleanSteamID(params["steam_id"])
		if err != nil {
			return "", err
		}
		return "BanPlayer " + id, nil

	case ActionUnbanPlayer:
		id, err := cleanSteamID(params["steam_id"])
		if err != nil {
			return "", err
		}
		return "UnbanPlayer " + id, nil

	case ActionSetTimeOfDay:
		t := strings.TrimSpace(params["time"])
		if !timeOfDayPattern.MatchString(t) {
			return "", fmt.Errorf("time must be HH:MM in 24-hour form")
		}
		return "SetTimeOfDay " + t, nil

	default:
		// An unknown action must never fall through to a free-text command.
		return "", fmt.Errorf("unsupported action %q", action)
	}
}

// cleanSteamID validates a player identifier.
func cleanSteamID(raw string) (string, error) {
	id := strings.TrimSpace(raw)
	if id == "" {
		return "", fmt.Errorf("steam_id is required")
	}
	if !steamIDPattern.MatchString(id) {
		return "", fmt.Errorf("steam_id must be numeric")
	}
	// Reject a leading zero run that would not round-trip as an id.
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return "", fmt.Errorf("steam_id is out of range")
	}
	return id, nil
}

// cleanMessage strips anything that could break out of the command line.
func cleanMessage(raw string) (string, error) {
	msg := strings.TrimSpace(raw)
	if msg == "" {
		return "", fmt.Errorf("message is required")
	}
	if controlChars.MatchString(msg) {
		// Rejected rather than silently stripped: a broadcast that quietly
		// loses part of its text is worse than one that fails loudly.
		return "", fmt.Errorf("message must not contain control characters or newlines")
	}
	if len(msg) > maxBroadcastLen {
		return "", fmt.Errorf("message must be %d characters or fewer", maxBroadcastLen)
	}
	return msg, nil
}
