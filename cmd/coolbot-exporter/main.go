// Command coolbot-exporter collects CoolBot Pro data and writes to InfluxDB.
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

		// InfluxDB configuration
		influxURL    = flag.String("influx-url", "http://minis:8086", "InfluxDB server URL")
		influxToken  = flag.String("influx-token", "", "InfluxDB API token")
		influxOrg    = flag.String("influx-org", "home", "InfluxDB organization")
		influxBucket = flag.String("influx-bucket", "coolbot", "InfluxDB bucket")

		// General options
		interval = flag.Duration("interval", 15*time.Second, "Metrics write interval")
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
	if *influxToken == "" {
		*influxToken = os.Getenv("INFLUXDB_TOKEN")
	}
	if *influxURL == "" || *influxURL == "http://minis:8086" {
		if env := os.Getenv("INFLUXDB_URL"); env != "" {
			*influxURL = env
		}
	}

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: CoolBot email and password are required")
		fmt.Fprintln(os.Stderr, "Set via flags or COOLBOT_EMAIL / COOLBOT_PASSWORD environment variables")
		os.Exit(1)
	}

	if *influxToken == "" {
		fmt.Fprintln(os.Stderr, "Error: InfluxDB token is required")
		fmt.Fprintln(os.Stderr, "Set via -influx-token flag or INFLUXDB_TOKEN environment variable")
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

	// Create InfluxDB writer
	influxCfg := metrics.Config{
		URL:    *influxURL,
		Token:  *influxToken,
		Org:    *influxOrg,
		Bucket: *influxBucket,
	}
	writer := metrics.NewWriter(influxCfg, logger)
	defer writer.Close()

	// Test InfluxDB connection
	ctx := context.Background()
	if err := writer.Ping(ctx); err != nil {
		logger.Error("failed to connect to InfluxDB", "error", err, "url", *influxURL)
		os.Exit(1)
	}

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

	// Initial metrics write
	writeMetrics(ctx, cb, writer, logger)

	// Start periodic metrics write
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Start listening for real-time updates
	go func() {
		if err := cb.Listen(ctx); err != nil && ctx.Err() == nil {
			logger.Error("listen error", "error", err)
		}
	}()

	logger.Info("started", "interval", *interval, "influx_url", *influxURL, "bucket", *influxBucket)

	// Main loop
	for {
		select {
		case <-ticker.C:
			writeMetrics(ctx, cb, writer, logger)
		case sig := <-sigCh:
			logger.Info("received signal, shutting down", "signal", sig)
			cancel()
			return
		}
	}
}

func writeMetrics(ctx context.Context, cb *client.Client, writer *metrics.Writer, logger *slog.Logger) {
	devices := cb.GetDevices()
	for _, dev := range devices {
		m := metrics.DeviceMetrics{
			DeviceID:        dev.ID,
			DeviceName:      dev.Name,
			RoomTemp:        dev.RoomTemp,
			FrostTemp:       dev.FrostTemp,
			SetPoint:        dev.SetPoint,
			Humidity:        dev.Humidity,
			CompressorState: dev.CompressorState,
			RSSI:            dev.RSSI,
			Online:          dev.Status == "ONLINE",
			FirmwareVersion: dev.FirmwareVersion,
			Status:          dev.Status,
		}

		if err := writer.WriteDeviceMetrics(ctx, m); err != nil {
			logger.Error("failed to write metrics", "error", err, "device", dev.Name)
		}
	}
}
