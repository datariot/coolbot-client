// Command coolbot-exporter collects CoolBot Pro data and publishes to MQTT.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/datariot/coolbot-client/internal/client"
	"github.com/datariot/coolbot-client/internal/metrics"
)

func main() {
	var (
		// CoolBot credentials
		email    = flag.String("email", "", "CoolBot account email")
		password = flag.String("password", "", "CoolBot account password")
		server   = flag.String("server", client.DefaultServer, "CoolBot server address")

		// MQTT configuration
		mqttBroker   = flag.String("mqtt-broker", "tcp://minis.local:1883", "MQTT broker URL")
		mqttClientID = flag.String("mqtt-client-id", "coolbot-exporter", "MQTT client ID")
		mqttUsername = flag.String("mqtt-username", "", "MQTT username (optional)")
		mqttPassword = flag.String("mqtt-password", "", "MQTT password (optional)")
		mqttPrefix   = flag.String("mqtt-prefix", "farm/sensors/coolbot", "MQTT topic prefix")

		// General options
		interval = flag.Duration("interval", 15*time.Second, "Publish interval")
		logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Environment variable fallbacks
	if *email == "" {
		*email = os.Getenv("COOLBOT_EMAIL")
	}
	if *password == "" {
		*password = os.Getenv("COOLBOT_PASSWORD")
	}
	if *mqttBroker == "tcp://minis.local:1883" {
		if env := os.Getenv("MQTT_BROKER"); env != "" {
			*mqttBroker = env
		}
	}
	if *mqttUsername == "" {
		*mqttUsername = os.Getenv("MQTT_USERNAME")
	}
	if *mqttPassword == "" {
		*mqttPassword = os.Getenv("MQTT_PASSWORD")
	}

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: CoolBot email and password are required")
		fmt.Fprintln(os.Stderr, "Set via flags or COOLBOT_EMAIL / COOLBOT_PASSWORD environment variables")
		os.Exit(1)
	}

	// Configure logging
	level := slog.LevelInfo
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	// Create MQTT publisher
	mqttCfg := metrics.Config{
		Broker:   *mqttBroker,
		ClientID: *mqttClientID,
		Username: *mqttUsername,
		Password: *mqttPassword,
		Prefix:   *mqttPrefix,
	}

	logger.Info("connecting to MQTT", "broker", *mqttBroker)
	publisher, err := metrics.NewPublisher(mqttCfg, logger)
	if err != nil {
		logger.Error("failed to connect to MQTT", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// Create CoolBot client
	cb := client.New(*email, *password,
		client.WithServer(*server),
		client.WithLogger(logger),
	)

	// Set up context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Connect to CoolBot
	logger.Info("connecting to CoolBot", "server", *server, "email", *email)
	if err := cb.Connect(ctx); err != nil {
		logger.Error("failed to connect to CoolBot", "error", err)
		os.Exit(1)
	}
	defer cb.Close()

	// Register callback for real-time updates
	cb.OnUpdate(func(update client.DeviceUpdate) {
		// Publish individual pin updates to MQTT immediately
		if err := publisher.PublishSingle(update.PinName, update.Value); err != nil {
			logger.Error("failed to publish update", "error", err, "pin", update.PinName)
		} else {
			logger.Debug("real-time update", "pin", update.PinName, "value", update.Value)
		}
	})

	// Initial publish
	logDevices(cb, logger)
	publishMetrics(cb, publisher, logger)

	// Start periodic publish
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Listen for real-time updates, and reconnect when the connection dies.
	// The first version of this logged one "listen error" and carried on
	// publishing the zero-initialised device state every interval, for
	// months — a dead socket must be a reconnect, never a quiet fact.
	go func() {
		backoff := 5 * time.Second
		for {
			err := cb.Listen(ctx)
			if ctx.Err() != nil {
				return
			}
			logger.Error("listen error; reconnecting", "error", err, "in", backoff)
			cb.MarkDisconnected()
			cb.Close()
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if err := cb.Connect(ctx); err != nil {
				logger.Error("reconnect failed", "error", err)
				if backoff < 5*time.Minute {
					backoff *= 2
				}
				continue
			}
			backoff = 5 * time.Second
			logDevices(cb, logger)
			publishMetrics(cb, publisher, logger)
		}
	}()

	logger.Info("started", "interval", *interval, "mqtt_broker", *mqttBroker, "prefix", *mqttPrefix)

	// Main loop
	for {
		select {
		case <-ticker.C:
			publishMetrics(cb, publisher, logger)
		case sig := <-sigCh:
			logger.Info("received signal, shutting down", "signal", sig)
			cancel()
			return
		}
	}
}

// logDevices says what the cloud reported at connect: which controllers, and
// whether the cloud can reach them. `online=0` on the wire is this status.
func logDevices(cb *client.Client, logger *slog.Logger) {
	for _, dev := range cb.GetDevices() {
		logger.Info("device", "id", dev.ID, "name", dev.Name, "status", dev.Status,
			"firmware", dev.FirmwareVersion, "room_temp", dev.RoomTemp, "set_point", dev.SetPoint)
	}
}

func publishMetrics(cb *client.Client, publisher *metrics.Publisher, logger *slog.Logger) {
	devices := cb.GetDevices()
	for _, dev := range devices {
		online := dev.Status == "ONLINE"
		if !online {
			// The cloud says the controller is unreachable (or we are not
			// connected to the cloud): say so, and say nothing else. The
			// last-known temperatures are not readings, and a 0.00 °F room
			// is a lie a dashboard would plot.
			if err := publisher.PublishSingle("online", "0"); err != nil {
				logger.Error("failed to publish online", "error", err, "device", dev.Name)
			}
			continue
		}
		m := metrics.DeviceMetrics{
			DeviceID:        dev.ID,
			DeviceName:      dev.Name,
			RoomTemp:        dev.RoomTemp,
			FrostTemp:       dev.FrostTemp,
			SetPoint:        dev.SetPoint,
			Humidity:        dev.Humidity,
			CompressorState: dev.CompressorState,
			RSSI:            dev.RSSI,
			Online:          online,
			FirmwareVersion: dev.FirmwareVersion,
			Status:          dev.Status,
		}

		if err := publisher.PublishDeviceMetrics(m); err != nil {
			logger.Error("failed to publish metrics", "error", err, "device", dev.Name)
		}
	}
}
