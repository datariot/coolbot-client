// Package protocol implements the Blynk binary WebSocket protocol used by CoolBot Pro.
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// Command types in the Blynk binary protocol.
const (
	CmdResponse          byte = 0
	CmdLogin             byte = 2
	CmdSaveProfile       byte = 3
	CmdLoadProfile       byte = 4
	CmdGetToken          byte = 5
	CmdPing              byte = 6
	CmdActivate          byte = 7
	CmdDeactivate        byte = 8
	CmdRefresh           byte = 9
	CmdTweet             byte = 12
	CmdEmail             byte = 13
	CmdNotify            byte = 14
	CmdBridge            byte = 15
	CmdHardwareSync      byte = 16
	CmdInternal          byte = 17
	CmdProperty          byte = 19
	CmdHardware          byte = 20
	CmdHardwareLogin     byte = 29
	CmdRedirect          byte = 41
	CmdDebugPrint        byte = 55
	CmdEventLog          byte = 64
)

// Response status codes.
const (
	StatusOK                 uint16 = 200
	StatusQuotaLimit         uint16 = 1
	StatusIllegalCommand     uint16 = 2
	StatusNotRegistered      uint16 = 3
	StatusAlreadyRegistered  uint16 = 4
	StatusNotAuthenticated   uint16 = 5
	StatusNotAllowed         uint16 = 6
	StatusDeviceNotInNetwork uint16 = 7
	StatusNoActiveDevice     uint16 = 8
	StatusInvalidToken       uint16 = 9
	StatusIllegalCommandBody uint16 = 11
	StatusGetGraphData       uint16 = 12
	StatusNoData             uint16 = 17
	StatusDeviceWentOffline  uint16 = 18
	StatusServerError        uint16 = 19
	StatusNotSupported       uint16 = 20
)

// HeaderSize is the size of the Blynk message header.
const HeaderSize = 5

// Message represents a parsed Blynk protocol message.
type Message struct {
	Command   byte
	MessageID uint16
	Length    uint16
	Body      []byte
	Params    []string // Parsed null-separated parameters
}

// ParseMessage parses a binary Blynk protocol message.
func ParseMessage(data []byte) (*Message, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("message too short: %d bytes", len(data))
	}

	msg := &Message{
		Command:   data[0],
		MessageID: binary.BigEndian.Uint16(data[1:3]),
		Length:    binary.BigEndian.Uint16(data[3:5]),
	}

	if len(data) > HeaderSize {
		msg.Body = data[HeaderSize:]
		// Parse null-separated parameters
		msg.Params = parseParams(msg.Body)
	}

	return msg, nil
}

// parseParams splits body by null bytes into string parameters.
func parseParams(body []byte) []string {
	if len(body) == 0 {
		return nil
	}
	// Split by null bytes, but handle last param which may not have trailing null
	parts := bytes.Split(body, []byte{0})
	params := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) > 0 {
			params = append(params, string(p))
		}
	}
	return params
}

// FormatMessage creates a binary Blynk protocol message.
func FormatMessage(cmd byte, msgID uint16, params ...string) []byte {
	// Join params with null separator
	body := []byte(strings.Join(params, "\x00"))

	// Build message: header + body
	msg := make([]byte, HeaderSize+len(body))
	msg[0] = cmd
	binary.BigEndian.PutUint16(msg[1:3], msgID)
	binary.BigEndian.PutUint16(msg[3:5], uint16(len(body)))
	copy(msg[HeaderSize:], body)

	return msg
}

// FormatResponse creates a response message (used for pong, etc).
func FormatResponse(msgID uint16, status uint16) []byte {
	msg := make([]byte, HeaderSize)
	msg[0] = CmdResponse
	binary.BigEndian.PutUint16(msg[1:3], msgID)
	binary.BigEndian.PutUint16(msg[3:5], status)
	return msg
}

// IsResponse returns true if this is a response message.
func (m *Message) IsResponse() bool {
	return m.Command == CmdResponse
}

// Status returns the status code for response messages.
func (m *Message) Status() uint16 {
	return m.Length // For responses, the length field contains status
}

// IsOK returns true if this is a successful response.
func (m *Message) IsOK() bool {
	return m.IsResponse() && m.Status() == StatusOK
}

// CommandName returns a human-readable name for the command.
func CommandName(cmd byte) string {
	names := map[byte]string{
		CmdResponse:      "RESPONSE",
		CmdLogin:         "LOGIN",
		CmdSaveProfile:   "SAVE_PROFILE",
		CmdLoadProfile:   "LOAD_PROFILE",
		CmdGetToken:      "GET_TOKEN",
		CmdPing:          "PING",
		CmdActivate:      "ACTIVATE",
		CmdDeactivate:    "DEACTIVATE",
		CmdRefresh:       "REFRESH",
		CmdHardwareSync:  "HW_SYNC",
		CmdInternal:      "INTERNAL",
		CmdProperty:      "PROPERTY",
		CmdHardware:      "HARDWARE",
		CmdHardwareLogin: "HW_LOGIN",
		CmdRedirect:      "REDIRECT",
	}
	if name, ok := names[cmd]; ok {
		return name
	}
	return fmt.Sprintf("CMD_%d", cmd)
}
