package logger

import (
	"os"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	// Initialize the logger
	err := Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Close()
	defer func() {
		// Clean up log files after test
		os.Remove(GetCurrentLogPath())
		os.Remove(GetCurrentJSONLogPath())
	}()

	// Set log level to Debug
	SetLevel(DebugLevel)

	// Set default component
	SetDefaultComponent("test-component")

	// Test basic logging
	Debug("This is a debug message")
	Info("This is an info message")
	Warn("This is a warning message")
	Error("This is an error message")

	// Test component-specific logging
	DebugWithComponent("custom-component", "Debug message with custom component")
	InfoWithComponent("custom-component", "Info message with custom component")
	WarnWithComponent("custom-component", "Warning message with custom component")
	ErrorWithComponent("custom-component", "Error message with custom component")

	// Test structured logging
	data := map[string]interface{}{
		"user_id":    12345,
		"request_id": "abc-123-xyz",
		"duration":   157.5,
	}
	DebugEvent("UserLogin", data)
	InfoEvent("RequestProcessed", data)
	WarnEvent("SlowResponse", data)
	ErrorEvent("DatabaseError", data)

	// Test component-specific structured logging
	DebugEventWithComponent("auth-service", "UserLogin", data)
	InfoEventWithComponent("api-gateway", "RequestProcessed", data)
	WarnEventWithComponent("database", "SlowQuery", data)
	ErrorEventWithComponent("file-service", "UploadFailed", data)

	// Test state tracking
	appState := struct {
		Version     string
		Environment string
		Uptime      int64
		Connections int
	}{
		Version:     "1.0.0",
		Environment: "test",
		Uptime:      3600,
		Connections: 42,
	}

	// Log application state
	LogState("app", appState)

	// Update state and log again
	appState.Connections = 50
	LogState("app", appState)

	// Get current state
	currentState := GetState("app")
	if currentState == nil {
		t.Error("Expected state to be tracked, got nil")
	}

	// Get state history
	stateHistory := GetStateHistory("app")
	if len(stateHistory) != 2 {
		t.Errorf("Expected 2 state entries, got %d", len(stateHistory))
	}

	// Verify log files were created
	if _, err := os.Stat(GetCurrentLogPath()); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", GetCurrentLogPath())
	}

	if _, err := os.Stat(GetCurrentJSONLogPath()); os.IsNotExist(err) {
		t.Errorf("JSON log file was not created at %s", GetCurrentJSONLogPath())
	}

	// Give some time for logs to be written
	time.Sleep(100 * time.Millisecond)

	t.Log("Log files created at:", GetCurrentLogPath(), "and", GetCurrentJSONLogPath())
}

func ExampleDebug() {
	Initialize()
	defer Close()

	// Set default component
	SetDefaultComponent("example")

	// Basic debug logging
	Debug("This is a debug message")

	// Debug with structured data
	DebugEvent("UserAction", map[string]interface{}{
		"user_id": 123,
		"action":  "login",
	})
}

func ExampleLogState() {
	Initialize()
	defer Close()

	// Define application state
	appState := struct {
		Status      string
		Connections int
	}{
		Status:      "running",
		Connections: 5,
	}

	// Log initial state
	LogState("app", appState)

	// Update state
	appState.Connections = 10
	LogState("app", appState)

	// Get current state
	currentState := GetState("app")
	_ = currentState // Use the state

	// Get state history
	stateHistory := GetStateHistory("app")
	_ = stateHistory // Use the history
}
