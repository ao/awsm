# AWSM Enhanced Logger

The AWSM Enhanced Logger is a powerful logging package designed for structured logging and state tracking to support AI debugging capabilities. It provides both human-readable logs and structured JSON logs that can be easily parsed by AI tools.

## Features

- **Multiple Log Levels**: Debug, Info, Warn, Error
- **Structured Logging**: JSON format for machine parsing
- **Human-Readable Logs**: Traditional log format for human reading
- **Component-Based Logging**: Organize logs by component/module
- **State Tracking**: Track and log application state changes
- **Context-Rich Logs**: All logs include timestamps, component names, and caller information

## Usage

### Basic Logging

```go
// Initialize the logger
logger.Initialize()
defer logger.Close()

// Set log level (default is InfoLevel)
logger.SetLevel(logger.DebugLevel)

// Basic logging
logger.Debug("This is a debug message")
logger.Info("This is an info message")
logger.Warn("This is a warning message")
logger.Error("This is an error message")
```

### Component-Based Logging

```go
// Set default component for all logs
logger.SetDefaultComponent("main")

// Or specify component for individual logs
logger.DebugWithComponent("auth", "User authentication attempt")
logger.InfoWithComponent("api", "Request processed successfully")
logger.WarnWithComponent("database", "Slow query detected")
logger.ErrorWithComponent("file-service", "Failed to upload file")
```

### Structured Event Logging

```go
// Log structured events
data := map[string]interface{}{
    "user_id": 12345,
    "action": "login",
    "duration_ms": 157.5,
}

logger.DebugEvent("UserAction", data)
logger.InfoEvent("RequestProcessed", data)
logger.WarnEvent("PerformanceIssue", data)
logger.ErrorEvent("OperationFailed", data)

// With component
logger.InfoEventWithComponent("auth", "UserLogin", data)
```

### State Tracking

```go
// Define application state
appState := struct {
    Version     string
    Environment string
    Connections int
}{
    Version:     "1.0.0",
    Environment: "production",
    Connections: 42,
}

// Log initial state
logger.LogState("app", appState)

// Update state
appState.Connections = 50
logger.LogState("app", appState)

// Get current state
currentState := logger.GetState("app")

// Get state history
stateHistory := logger.GetStateHistory("app")
```

### Log File Management

```go
// Get current log file paths
humanReadableLog := logger.GetCurrentLogPath()
jsonLog := logger.GetCurrentJSONLogPath()

// Custom output (for testing)
logger.SetOutput(os.Stdout)
logger.SetJSONOutput(customWriter)
```

## Log Format

### Human-Readable Format

```
[2025/08/01 00:50:22] [INFO] [main] [example.go:42] Message here {"key": "value"}
```

Format: `[timestamp] [level] [component] [caller] message {data}`

### JSON Format

```json
{
  "timestamp": "2025-08-01T00:50:22Z",
  "level": "INFO",
  "component": "main",
  "message": "Message here",
  "caller": "example.go:42",
  "data": {
    "key": "value"
  },
  "state_update": false
}
```

## AI Debugging Support

The structured JSON logs are designed to be easily parsed by AI tools for debugging purposes. Each log entry contains rich context including:

- Timestamp for temporal analysis
- Component name for module-specific filtering
- Caller information for code location
- Structured data for detailed analysis
- State tracking for understanding application state changes

This enables AI tools to:
- Correlate events across time
- Identify patterns in application behavior
- Track state changes that led to issues
- Analyze performance bottlenecks
- Provide more accurate debugging assistance