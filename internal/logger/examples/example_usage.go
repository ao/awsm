package main

import (
	"time"

	"github.com/ao/awsm/internal/logger"
)

// AppState represents the application state
type AppState struct {
	Version     string
	Environment string
	Uptime      int64
	Connections int
	LastUpdated time.Time
}

func main() {
	// Initialize the logger
	err := logger.Initialize()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Close()

	// Set log level to Debug for development
	logger.SetLevel(logger.DebugLevel)

	// Set default component name
	logger.SetDefaultComponent("main")

	// Log startup information
	logger.Info("Application starting up")
	logger.InfoEvent("AppStartup", map[string]interface{}{
		"version": "1.0.0",
		"env":     "development",
	})

	// Initialize application state
	appState := AppState{
		Version:     "1.0.0",
		Environment: "development",
		Uptime:      0,
		Connections: 0,
		LastUpdated: time.Now(),
	}

	// Log initial state
	logger.LogState("app", appState)

	// Simulate user login
	logger.DebugWithComponent("auth", "User login attempt")

	userData := map[string]interface{}{
		"user_id":    12345,
		"username":   "testuser",
		"ip_address": "192.168.1.100",
		"timestamp":  time.Now().Unix(),
	}

	logger.InfoEventWithComponent("auth", "UserLogin", userData)

	// Simulate processing a request
	logger.DebugWithComponent("api", "Processing request")

	// Simulate a slow database query
	logger.WarnEventWithComponent("database", "SlowQuery", map[string]interface{}{
		"query":     "SELECT * FROM large_table",
		"duration":  1532.5,
		"rows":      10000,
		"timestamp": time.Now().Unix(),
	})

	// Update application state
	appState.Connections = 1
	appState.Uptime = 60
	appState.LastUpdated = time.Now()

	// Log updated state
	logger.LogState("app", appState)

	// Simulate an error
	logger.ErrorEventWithComponent("file-service", "UploadFailed", map[string]interface{}{
		"file_name":  "large_document.pdf",
		"file_size":  15728640,
		"error_code": "STORAGE_FULL",
		"timestamp":  time.Now().Unix(),
	})

	// Get current log file paths
	humanReadableLog := logger.GetCurrentLogPath()
	jsonLog := logger.GetCurrentJSONLogPath()

	logger.Info("Logs are being written to %s and %s", humanReadableLog, jsonLog)

	// Application shutdown
	logger.Info("Application shutting down")
}
