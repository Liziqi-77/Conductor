//go:build zmq
// +build zmq

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"conductor.local/kvevent"
)

// -----------------------------------------------------------------------------
// 2. Main Entry
// -----------------------------------------------------------------------------

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Enable Debug logs to see ZMQ details
	}))
	slog.SetDefault(logger)

	slog.Info("Starting KV Event Manager Demo...")

	// 1. Configuration (Static Service List)
	// You should modify these values to match your vLLM environment
	services := []kvevent.ServiceConfig{
		{
			Name:      "vllm-local",
			IP:        "127.0.0.1", // Assuming vLLM is running locally
			Port:      5557,        // Default vLLM ZMQ port
			Type:      kvevent.ServiceTypeVLLM,
			ModelName: "qwen2.5-7b",
			LoraID:    -1,
		},
		// Add more services here...
	}

	// 2. Initialize Dependencies
	provider := kvevent.SyncIndexProvider{}

	// 3. Create Manager
	manager := kvevent.NewStaticManager(services, provider)

	// 4. Start Manager
	if err := manager.Start(); err != nil {
		slog.Error("Failed to start manager", "error", err)
		os.Exit(1)
	}

	// 5. Wait for Signal (Graceful Shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Manager is running. Press Ctrl+C to stop.")
	<-sigChan

	// 6. Shutdown
	slog.Info("Shutting down...")
	manager.Stop()
	slog.Info("Bye!")
}
