// Package protocol implements the Blynk WebSocket protocol used by CoolBot Pro.
package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

// Separator is the middle dot character used in Blynk protocol messages.
const Separator = "·" // U+00B7

// Command types in the Blynk protocol.
const (
	CmdLogin              = "LOGIN"
	CmdResponse           = "RESPONSE"
	CmdLoadProfileGzipped = "LOAD_PROFILE_GZIPPED"
	CmdGetProjectByToken  = "GET_PROJECT_BY_TOKEN"
	CmdHardware           = "HARDWARE"
	CmdPing               = "PING"
	CmdPong               = "PONG"
)

// HardwareStreamID is the special message ID used for streaming hardware updates.
const HardwareStreamID = 7778

// Message represents a parsed Blynk protocol message.
type Message struct {
	Command   string
	MessageID int
	Params    []string
}

// ParseMessage parses a raw Blynk protocol message.
// Format: COMMAND @id: [param1·param2·...]
func ParseMessage(raw string) (*Message, error) {
	// Find command and message ID
	atIdx := strings.Index(raw, " @")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid message format: missing @")
	}

	command := raw[:atIdx]

	// Find message ID
	colonIdx := strings.Index(raw[atIdx:], ":")
	if colonIdx == -1 {
		// Response format: RESPONSE @1 = Ok
		eqIdx := strings.Index(raw[atIdx:], " = ")
		if eqIdx != -1 {
			idStr := raw[atIdx+2 : atIdx+eqIdx]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				return nil, fmt.Errorf("invalid message ID: %w", err)
			}
			value := raw[atIdx+eqIdx+3:]
			return &Message{
				Command:   command,
				MessageID: id,
				Params:    []string{value},
			}, nil
		}
		return nil, fmt.Errorf("invalid message format: missing :")
	}

	idStr := raw[atIdx+2 : atIdx+colonIdx]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid message ID: %w", err)
	}

	// Parse parameters
	paramsStart := atIdx + colonIdx + 1
	paramsStr := strings.TrimSpace(raw[paramsStart:])

	// Remove brackets if present
	if strings.HasPrefix(paramsStr, "[") && strings.HasSuffix(paramsStr, "]") {
		paramsStr = paramsStr[1 : len(paramsStr)-1]
	}

	var params []string
	if paramsStr != "" {
		params = strings.Split(paramsStr, Separator)
	}

	return &Message{
		Command:   command,
		MessageID: id,
		Params:    params,
	}, nil
}

// FormatMessage creates a Blynk protocol message string.
func FormatMessage(command string, messageID int, params ...string) string {
	paramsStr := strings.Join(params, Separator)
	return fmt.Sprintf("%s @%d: [%s]", command, messageID, paramsStr)
}

// LoginMessage creates a LOGIN command message.
func LoginMessage(messageID int, email, passwordHash, clientType, buildNumber, platform string) string {
	return FormatMessage(CmdLogin, messageID, email, passwordHash, clientType, buildNumber, platform)
}

// LoadProfileMessage creates a LOAD_PROFILE_GZIPPED command message.
func LoadProfileMessage(messageID int) string {
	return FormatMessage(CmdLoadProfileGzipped, messageID)
}

// GetProjectByTokenMessage creates a GET_PROJECT_BY_TOKEN command message.
func GetProjectByTokenMessage(messageID int, token string) string {
	return FormatMessage(CmdGetProjectByToken, messageID, token)
}
