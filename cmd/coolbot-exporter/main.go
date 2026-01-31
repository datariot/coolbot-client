// Command coolbot-exporter runs a Prometheus exporter for CoolBot Pro devices.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/datariot/coolbot-client/internal/client"
	"github.com/datariot/coolbot-client/internal/metrics"
)

func main() {
	var (
		listenAddr = flag.String("listen", ":9120", "Address to listen on for metrics")
		email      = flag.String("email", "", "CoolBot account email")
		password   = flag.String("password", "", "CoolBot account password")
		server     = flag.String("server", client.DefaultServer, "CoolBot server address")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	// Environment variable fallbacks
	if *email == "" {
		*email = os.Getenv("COOLBOT_EMAIL")
	}
	if *password == "" {
		*password = os.Getenv("COOLBOT_PASSWORD")
	}

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "Error: email and password are required")
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

	// Create client
	cb := client.New(*email, *password,
		client.WithServer(*server),
		client.WithLogger(logger),
	)

	// Register update callback for metrics
	cb.OnUpdate(func(update client.DeviceUpdate) {
		logger.Debug("device update",
			"device_id", update.DeviceID,
			"pin", update.PinName,
			"value", update.Value,
		)
	})

	// Set up context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Connect to CoolBot
	logger.Info("connecting to CoolBot", "server", *server, "email", *email)
	if err := cb.Connect(ctx); err != nil {
		logger.Error("failed to connect", "error", err)
		os.Exit(1)
	}
	defer cb.Close()

	// Initial metrics update
	updateMetrics(cb)

	// Start metrics HTTP server
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>CoolBot Exporter</title></head>
<body>
<h1>CoolBot Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`))
	})

	httpServer := &http.Server{Addr: *listenAddr}
	go func() {
		logger.Info("starting metrics server", "addr", *listenAddr)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("http server error", "error", err)
		}
	}()

	// Start periodic metrics update
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Start listening for real-time updates
	go func() {
		if err := cb.Listen(ctx); err != nil && ctx.Err() == nil {
			logger.Error("listen error", "error", err)
		}
	}()

	// Main loop
	for {
		select {
		case <-ticker.C:
			updateMetrics(cb)
		case sig := <-sigCh:
			logger.Info("received signal, shutting down", "signal", sig)
			cancel()

			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			httpServer.Shutdown(shutdownCtx)
			return
		}
	}
}

func updateMetrics(cb *client.Client) {
	devices := cb.GetDevices()
	for _, dev := range devices {
		deviceID := strconv.Itoa(dev.ID)
		labels := []string{deviceID, dev.Name}

		metrics.RoomTemperature.WithLabelValues(labels...).Set(dev.RoomTemp)
		metrics.FrostTemperature.WithLabelValues(labels...).Set(dev.FrostTemp)
		metrics.SetPointTemperature.WithLabelValues(labels...).Set(dev.SetPoint)
		metrics.Humidity.WithLabelValues(labels...).Set(dev.Humidity)
		metrics.CompressorState.WithLabelValues(labels...).Set(float64(dev.CompressorState))
		metrics.RSSI.WithLabelValues(labels...).Set(float64(dev.RSSI))
		metrics.LastUpdateTimestamp.WithLabelValues(labels...).Set(float64(dev.LastUpdate.Unix()))

		online := 0.0
		if dev.Status == "ONLINE" {
			online = 1.0
		}
		metrics.DeviceOnline.WithLabelValues(labels...).Set(online)

		metrics.DeviceInfo.WithLabelValues(deviceID, dev.Name, dev.FirmwareVersion, dev.Status).Set(1)
	}
}
