// Package protocol implements the Blynk binary protocol used by CoolBot Pro.
package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

// Command bytes for the Blynk protocol.
const (
	CmdResponse           byte = 0x00
	CmdLogin              byte = 0x02
	CmdPing               byte = 0x06
	CmdLoadProfileGzipped byte = 0x0B // 11
	CmdHardware           byte = 0x14 // 20
	CmdGetProjectByToken  byte = 0x1A // 26
)

// HardwareStreamID is the special message ID used for streaming hardware updates.
const HardwareStreamID = 7778

// Message represents a Blynk protocol message.
type Message struct {
	Command   byte
	MessageID uint16
	Body      []byte
}

// Encode serializes the message to binary format.
// Format: [command:1][msgId:2][length:2][body:N]
func (m *Message) Encode() []byte {
	buf := make([]byte, 5+len(m.Body))
	buf[0] = m.Command
	binary.BigEndian.PutUint16(buf[1:3], m.MessageID)
	binary.BigEndian.PutUint16(buf[3:5], uint16(len(m.Body)))
	copy(buf[5:], m.Body)
	return buf
}

// ParseMessage parses a binary Blynk message.
func ParseMessage(data []byte) (*Message, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("message too short: %d bytes", len(data))
	}

	msg := &Message{
		Command:   data[0],
		MessageID: binary.BigEndian.Uint16(data[1:3]),
	}

	length := binary.BigEndian.Uint16(data[3:5])
	if len(data) < int(5+length) {
		return nil, fmt.Errorf("incomplete message: expected %d bytes, got %d", 5+length, len(data))
	}

	msg.Body = data[5 : 5+length]
	return msg, nil
}

// ParseParams splits the message body by null bytes into parameters.
func (m *Message) ParseParams() []string {
	if len(m.Body) == 0 {
		return nil
	}
	// Trim trailing null if present
	body := bytes.TrimRight(m.Body, "\x00")
	return strings.Split(string(body), "\x00")
}

// EncodeParams joins parameters with null bytes.
func EncodeParams(params ...string) []byte {
	return []byte(strings.Join(params, "\x00"))
}

// NewLoginMessage creates a LOGIN message.
func NewLoginMessage(messageID uint16, email, passwordHash, clientType, buildNumber, platform string) *Message {
	return &Message{
		Command:   CmdLogin,
		MessageID: messageID,
		Body:      EncodeParams(email, passwordHash, clientType, buildNumber, platform),
	}
}

// NewLoadProfileMessage creates a LOAD_PROFILE_GZIPPED message.
func NewLoadProfileMessage(messageID uint16) *Message {
	return &Message{
		Command:   CmdLoadProfileGzipped,
		MessageID: messageID,
		Body:      nil,
	}
}

// NewGetProjectByTokenMessage creates a GET_PROJECT_BY_TOKEN message.
func NewGetProjectByTokenMessage(messageID uint16, token string) *Message {
	return &Message{
		Command:   CmdGetProjectByToken,
		MessageID: messageID,
		Body:      []byte(token),
	}
}

// IsOK checks if a response indicates success.
func (m *Message) IsOK() bool {
	// Response with status code 200 in the length field means OK
	return m.Command == CmdResponse && binary.BigEndian.Uint16(m.Body) == 200
}

// ResponseStatus returns the status code from a response message.
func (m *Message) ResponseStatus() uint16 {
	if m.Command == CmdResponse && len(m.Body) >= 2 {
		return binary.BigEndian.Uint16(m.Body)
	}
	// For response messages, the "length" field actually contains the status
	return 0
}
