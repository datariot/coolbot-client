// Package client provides a WebSocket client for CoolBot Pro devices.
package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/datariot/coolbot-client/internal/protocol"
)

const (
	DefaultServer    = "cbws.storeitcold.com"
	DefaultWSPath    = "/websocket"
	HeartbeatTimeout = 10 * time.Second
	ReadTimeout      = 30 * time.Second
)

// Client is a CoolBot Pro WebSocket client.
type Client struct {
	server      string
	credentials *protocol.Credentials
	conn        *websocket.Conn
	msgID       atomic.Uint32
	logger      *slog.Logger

	mu        sync.RWMutex
	profile   *Profile
	devices   map[int]*DeviceState
	callbacks []func(DeviceUpdate)
}

// DeviceState holds the current state of a CoolBot device.
type DeviceState struct {
	ID              int
	Name            string
	Token           string
	Status          string
	RoomTemp        float64
	FrostTemp       float64
	SetPoint        float64
	Humidity        float64
	CompressorState int
	RSSI            int
	FirmwareVersion string
	LastUpdate      time.Time
}

// DeviceUpdate is sent when device state changes.
type DeviceUpdate struct {
	DeviceID  int
	Pin       int
	PinName   string
	Value     string
	Timestamp time.Time
}

// Profile contains the user profile data from the server.
type Profile struct {
	Dashboards   []Dashboard  `json:"dashBoards"`
	Subscription Subscription `json:"subscription"`
	Account      Account      `json:"account"`
}

type Dashboard struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Devices     []Device          `json:"devices"`
	Widgets     []Widget          `json:"widgets"`
	PinsStorage map[string]string `json:"pinsStorage"`
}

type Device struct {
	ID             int          `json:"id"`
	Name           string       `json:"name"`
	BoardType      string       `json:"boardType"`
	Token          string       `json:"token"`
	Status         string       `json:"status"`
	ConnectionType string       `json:"connectionType"`
	ConnectTime    int64        `json:"connectTime"`
	DisconnectTime int64        `json:"disconnectTime"`
	LastLoggedIP   string       `json:"lastLoggedIP"`
	HardwareInfo   HardwareInfo `json:"hardwareInfo"`
}

type HardwareInfo struct {
	Version           string `json:"version"`
	BoardType         string `json:"boardType"`
	Build             string `json:"build"`
	HeartbeatInterval int    `json:"heartbeatInterval"`
}

type Widget struct {
	Type      string  `json:"type"`
	ID        int     `json:"id"`
	Label     string  `json:"label"`
	DeviceID  int     `json:"deviceId"`
	PinType   string  `json:"pinType"`
	Pin       int     `json:"pin"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Value     string  `json:"value"`
	Frequency int     `json:"frequency"`
}

type Subscription struct {
	Plan   string `json:"plan"`
	Status string `json:"status"`
}

type Account struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// New creates a new CoolBot client.
func New(email, password string, opts ...Option) *Client {
	c := &Client{
		server:      DefaultServer,
		credentials: protocol.NewCredentials(email, password),
		logger:      slog.Default(),
		devices:     make(map[int]*DeviceState),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option configures a Client.
type Option func(*Client)

// WithServer sets a custom server address.
func WithServer(server string) Option {
	return func(c *Client) {
		c.server = server
	}
}

// WithLogger sets a custom logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// OnUpdate registers a callback for device updates.
func (c *Client) OnUpdate(fn func(DeviceUpdate)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.callbacks = append(c.callbacks, fn)
}

// Connect establishes a connection and authenticates.
func (c *Client) Connect(ctx context.Context) error {
	url := fmt.Sprintf("wss://%s%s", c.server, DefaultWSPath)
	c.logger.Info("connecting", "url", url)

	opts := &websocket.DialOptions{
		HTTPHeader: map[string][]string{
			"Origin": {"https://cb.storeitcold.com"},
		},
	}

	conn, _, err := websocket.Dial(ctx, url, opts)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	c.conn = conn

	// Send LOGIN
	if err := c.login(ctx); err != nil {
		conn.Close(websocket.StatusInternalError, "login failed")
		return err
	}

	// Load profile
	if err := c.loadProfile(ctx); err != nil {
		conn.Close(websocket.StatusInternalError, "load profile failed")
		return err
	}

	c.logger.Info("connected", "email", c.credentials.Email)
	return nil
}

// Close closes the connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close(websocket.StatusNormalClosure, "")
	}
	return nil
}

func (c *Client) nextMsgID() uint16 {
	return uint16(c.msgID.Add(1))
}

func (c *Client) sendBinary(ctx context.Context, data []byte) error {
	c.logger.Debug("sending", "cmd", protocol.CommandName(data[0]), "hex", hex.EncodeToString(data[:min(len(data), 20)]))
	return c.conn.Write(ctx, websocket.MessageBinary, data)
}

func (c *Client) receiveBinary(ctx context.Context) (*protocol.Message, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, ReadTimeout)
	defer cancel()

	_, data, err := c.conn.Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	msg, err := protocol.ParseMessage(data)
	if err != nil {
		return nil, data, err
	}

	c.logger.Debug("received", "cmd", protocol.CommandName(msg.Command), "msgID", msg.MessageID, "len", msg.Length)
	return msg, data, nil
}

func (c *Client) login(ctx context.Context) error {
	msgID := c.nextMsgID()
	params := c.credentials.LoginParams()
	msg := protocol.FormatMessage(protocol.CmdLogin, msgID, params...)

	c.logger.Debug("login", "params", params[0]) // Log email only

	if err := c.sendBinary(ctx, msg); err != nil {
		return fmt.Errorf("send login: %w", err)
	}

	resp, _, err := c.receiveBinary(ctx)
	if err != nil {
		return fmt.Errorf("receive login response: %w", err)
	}

	if !resp.IsOK() {
		return fmt.Errorf("login failed: status=%d", resp.Status())
	}

	c.logger.Debug("login successful")
	return nil
}

func (c *Client) loadProfile(ctx context.Context) error {
	msgID := c.nextMsgID()
	msg := protocol.FormatMessage(protocol.CmdLoadProfile, msgID)

	if err := c.sendBinary(ctx, msg); err != nil {
		return fmt.Errorf("send load profile: %w", err)
	}

	resp, data, err := c.receiveBinary(ctx)
	if err != nil {
		return fmt.Errorf("receive profile: %w", err)
	}

	// Profile response has the JSON in the body
	var profileData []byte
	if resp.Command == protocol.CmdLoadProfile && len(resp.Body) > 0 {
		profileData = resp.Body
	} else if len(data) > protocol.HeaderSize {
		profileData = data[protocol.HeaderSize:]
	}

	profile, err := c.parseProfile(profileData)
	if err != nil {
		return fmt.Errorf("parse profile: %w", err)
	}

	c.mu.Lock()
	c.profile = profile
	c.initDeviceStates()
	c.mu.Unlock()

	c.logger.Info("profile loaded", "dashboards", len(profile.Dashboards))
	return nil
}

func (c *Client) parseProfile(data []byte) (*Profile, error) {
	// Try to find JSON start - may have protocol prefix or be gzipped
	jsonStart := bytes.IndexByte(data, '{')
	if jsonStart == -1 {
		// Try GZIP decompression
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("gzip reader: %w (data starts with: %x)", err, data[:min(len(data), 10)])
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("gzip read: %w", err)
		}
		data = decompressed
		jsonStart = bytes.IndexByte(data, '{')
	}

	if jsonStart >= 0 {
		data = data[jsonStart:]
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}
	return &profile, nil
}

func (c *Client) initDeviceStates() {
	for _, dash := range c.profile.Dashboards {
		for _, dev := range dash.Devices {
			state := &DeviceState{
				ID:         dev.ID,
				Name:       dev.Name,
				Token:      dev.Token,
				Status:     dev.Status,
				LastUpdate: time.Now(),
			}

			// Initialize from pinsStorage
			for key, val := range dash.PinsStorage {
				var devID, pin int
				if _, err := fmt.Sscanf(key, "%d-v%d", &devID, &pin); err != nil {
					continue
				}
				if devID == dev.ID {
					c.applyPinValue(state, pin, val)
				}
			}

			c.devices[dev.ID] = state
		}
	}
}

func (c *Client) applyPinValue(state *DeviceState, pin int, value string) {
	switch pin {
	case protocol.PinRoomTemp:
		state.RoomTemp, _ = strconv.ParseFloat(value, 64)
	case protocol.PinFrostTemp:
		state.FrostTemp, _ = strconv.ParseFloat(value, 64)
	case protocol.PinSetPoint:
		state.SetPoint, _ = strconv.ParseFloat(value, 64)
	case protocol.PinHumidity:
		state.Humidity, _ = strconv.ParseFloat(value, 64)
	case protocol.PinCompressorState:
		state.CompressorState, _ = strconv.Atoi(value)
	case protocol.PinRSSI, protocol.PinRSSI0:
		state.RSSI, _ = strconv.Atoi(value)
	case protocol.PinFirmwareVersion:
		state.FirmwareVersion = value
	case protocol.PinStatus:
		state.Status = value
	}
	state.LastUpdate = time.Now()
}

// Listen processes incoming messages until context is cancelled.
func (c *Client) Listen(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, _, err := c.receiveBinary(ctx)
		if err != nil {
			return fmt.Errorf("receive: %w", err)
		}

		c.handleMessage(msg)
	}
}

func (c *Client) handleMessage(msg *protocol.Message) {
	switch msg.Command {
	case protocol.CmdHardware:
		c.handleHardware(msg)
	case protocol.CmdPing:
		// Send pong response
		pong := protocol.FormatResponse(msg.MessageID, protocol.StatusOK)
		c.sendBinary(context.Background(), pong)
	}
}

func (c *Client) handleHardware(msg *protocol.Message) {
	// Format: [dashboardId, "vw", pin, value]
	if len(msg.Params) < 4 {
		return
	}

	if msg.Params[1] != "vw" {
		return
	}

	pin, err := strconv.Atoi(msg.Params[2])
	if err != nil {
		return
	}
	value := msg.Params[3]

	c.mu.Lock()
	// Find device by dashboard (device 0 is typical)
	for _, state := range c.devices {
		c.applyPinValue(state, pin, value)
	}
	c.mu.Unlock()

	update := DeviceUpdate{
		DeviceID:  0,
		Pin:       pin,
		PinName:   protocol.PinName(pin),
		Value:     value,
		Timestamp: time.Now(),
	}

	c.mu.RLock()
	for _, cb := range c.callbacks {
		cb(update)
	}
	c.mu.RUnlock()
}

// GetDevices returns a copy of all device states.
func (c *Client) GetDevices() []DeviceState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	devices := make([]DeviceState, 0, len(c.devices))
	for _, d := range c.devices {
		devices = append(devices, *d)
	}
	return devices
}

// GetDevice returns a copy of a device state by ID.
func (c *Client) GetDevice(id int) (DeviceState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if d, ok := c.devices[id]; ok {
		return *d, true
	}
	return DeviceState{}, false
}
