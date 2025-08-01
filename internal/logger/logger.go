package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

// Log levels
const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Config represents the logger configuration
type Config struct {
	Level            LogLevel
	HumanReadable    bool
	JSONFormat       bool
	DefaultComponent string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       string                 `json:"level"`
	Component   string                 `json:"component"`
	Message     string                 `json:"message"`
	Caller      string                 `json:"caller"`
	Data        map[string]interface{} `json:"data,omitempty"`
	StateUpdate bool                   `json:"state_update,omitempty"`
}

var (
	logFile         *os.File
	jsonLogFile     *os.File
	logger          *log.Logger
	jsonLogger      *log.Logger
	once            sync.Once
	mu              sync.Mutex
	config          Config
	currentState    map[string]interface{}
	stateTracker    map[string][]interface{}
	logFilePath     string
	jsonLogFilePath string
)

// Initialize sets up the logger to write to files
func Initialize() error {
	var err error
	once.Do(func() {
		// Set default configuration
		config = Config{
			Level:            InfoLevel,
			HumanReadable:    true,
			JSONFormat:       true,
			DefaultComponent: "system",
		}

		// Initialize state tracking
		currentState = make(map[string]interface{})
		stateTracker = make(map[string][]interface{})

		// Create log files with timestamp in name
		timestamp := time.Now().Format("20060102-150405")
		logFileName := fmt.Sprintf("awsm-%s.log", timestamp)
		jsonLogFileName := fmt.Sprintf("awsm-%s.json", timestamp)

		// Store log file paths
		logFilePath = logFileName
		jsonLogFilePath = jsonLogFileName

		// Create human-readable log file
		if config.HumanReadable {
			logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return
			}
			logger = log.New(logFile, "", log.Ldate|log.Ltime)
		}

		// Create JSON log file
		if config.JSONFormat {
			jsonLogFile, err = os.OpenFile(jsonLogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				return
			}
			jsonLogger = log.New(jsonLogFile, "", 0) // No prefixes for JSON logs
		}

		// Log initialization
		Info("Logger initialized")
	})
	return err
}

// Close closes the log files
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		logFile.Close()
	}
	if jsonLogFile != nil {
		jsonLogFile.Close()
	}
}

// SetLevel sets the minimum log level
func SetLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	config.Level = level
}

// GetCurrentLogPath returns the path to the current log file
func GetCurrentLogPath() string {
	return logFilePath
}

// GetCurrentJSONLogPath returns the path to the current JSON log file
func GetCurrentJSONLogPath() string {
	return jsonLogFilePath
}

// getCallerInfo returns the file name and line number of the caller
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3) // Adjusted to get the actual caller
	if !ok {
		file = "unknown"
		line = 0
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// writeLog writes a log entry to both human-readable and JSON logs
func writeLog(level LogLevel, component string, message string, data map[string]interface{}, isStateUpdate bool) {
	mu.Lock()
	defer mu.Unlock()

	// Check if we need to initialize the logger
	if logger == nil && jsonLogger == nil {
		if err := Initialize(); err != nil {
			return
		}
	}

	// Check if this log level should be logged
	if level < config.Level {
		return
	}

	// Get caller information
	caller := getCallerInfo()

	// Get current timestamp
	timestamp := time.Now().Format(time.RFC3339)

	// Create log entry
	entry := LogEntry{
		Timestamp:   timestamp,
		Level:       level.String(),
		Component:   component,
		Message:     message,
		Caller:      caller,
		Data:        data,
		StateUpdate: isStateUpdate,
	}

	// Write to human-readable log
	if config.HumanReadable && logger != nil {
		var dataStr string
		if len(data) > 0 {
			dataBytes, _ := json.Marshal(data)
			dataStr = " " + string(dataBytes)
		}

		stateStr := ""
		if isStateUpdate {
			stateStr = " [STATE UPDATE]"
		}

		logger.Printf("[%s] [%s] [%s]%s %s%s",
			level.String(),
			component,
			caller,
			stateStr,
			message,
			dataStr)
	}

	// Write to JSON log
	if config.JSONFormat && jsonLogger != nil {
		jsonBytes, _ := json.Marshal(entry)
		jsonLogger.Println(string(jsonBytes))
	}
}

// LogState logs the current state of a component
func LogState(component string, state interface{}) {
	// Update the current state
	currentState[component] = state

	// Add to state history
	if _, ok := stateTracker[component]; !ok {
		stateTracker[component] = make([]interface{}, 0)
	}
	stateTracker[component] = append(stateTracker[component], state)

	// Convert state to map for logging
	var stateMap map[string]interface{}
	stateBytes, _ := json.Marshal(state)
	json.Unmarshal(stateBytes, &stateMap)

	// Log the state update
	writeLog(InfoLevel, component, "State updated", stateMap, true)
}

// GetState returns the current state for a component
func GetState(component string) interface{} {
	mu.Lock()
	defer mu.Unlock()
	return currentState[component]
}

// GetStateHistory returns the state history for a component
func GetStateHistory(component string) []interface{} {
	mu.Lock()
	defer mu.Unlock()
	return stateTracker[component]
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(DebugLevel, config.DefaultComponent, message, nil, false)
}

// DebugWithComponent logs a debug message with a specific component
func DebugWithComponent(component string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(DebugLevel, component, message, nil, false)
}

// DebugEvent logs a structured debug event
func DebugEvent(eventName string, data map[string]interface{}) {
	writeLog(DebugLevel, config.DefaultComponent, eventName, data, false)
}

// DebugEventWithComponent logs a structured debug event with a specific component
func DebugEventWithComponent(component string, eventName string, data map[string]interface{}) {
	writeLog(DebugLevel, component, eventName, data, false)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(InfoLevel, config.DefaultComponent, message, nil, false)
}

// InfoWithComponent logs an info message with a specific component
func InfoWithComponent(component string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(InfoLevel, component, message, nil, false)
}

// InfoEvent logs a structured info event
func InfoEvent(eventName string, data map[string]interface{}) {
	writeLog(InfoLevel, config.DefaultComponent, eventName, data, false)
}

// InfoEventWithComponent logs a structured info event with a specific component
func InfoEventWithComponent(component string, eventName string, data map[string]interface{}) {
	writeLog(InfoLevel, component, eventName, data, false)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(WarnLevel, config.DefaultComponent, message, nil, false)
}

// WarnWithComponent logs a warning message with a specific component
func WarnWithComponent(component string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(WarnLevel, component, message, nil, false)
}

// WarnEvent logs a structured warning event
func WarnEvent(eventName string, data map[string]interface{}) {
	writeLog(WarnLevel, config.DefaultComponent, eventName, data, false)
}

// WarnEventWithComponent logs a structured warning event with a specific component
func WarnEventWithComponent(component string, eventName string, data map[string]interface{}) {
	writeLog(WarnLevel, component, eventName, data, false)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(ErrorLevel, config.DefaultComponent, message, nil, false)
}

// ErrorWithComponent logs an error message with a specific component
func ErrorWithComponent(component string, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	writeLog(ErrorLevel, component, message, nil, false)
}

// ErrorEvent logs a structured error event
func ErrorEvent(eventName string, data map[string]interface{}) {
	writeLog(ErrorLevel, config.DefaultComponent, eventName, data, false)
}

// ErrorEventWithComponent logs a structured error event with a specific component
func ErrorEventWithComponent(component string, eventName string, data map[string]interface{}) {
	writeLog(ErrorLevel, component, eventName, data, false)
}

// SetOutput sets the output writer for the human-readable logger
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if logger != nil {
		logger.SetOutput(w)
	}
}

// SetJSONOutput sets the output writer for the JSON logger
func SetJSONOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if jsonLogger != nil {
		jsonLogger.SetOutput(w)
	}
}

// SetDefaultComponent sets the default component name
func SetDefaultComponent(component string) {
	mu.Lock()
	defer mu.Unlock()
	config.DefaultComponent = component
}
