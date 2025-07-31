package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	logFile *os.File
	logger  *log.Logger
	once    sync.Once
	mu      sync.Mutex
)

// Initialize sets up the logger to write to a file
func Initialize() error {
	var err error
	once.Do(func() {
		// Create log file with timestamp in name
		timestamp := time.Now().Format("20060102-150405")
		logFileName := fmt.Sprintf("awsm-%s.log", timestamp)

		logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return
		}

		// Create logger
		logger = log.New(logFile, "", log.Ldate|log.Ltime)

		// Log initialization
		Info("Logger initialized")
	})
	return err
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// getCallerInfo returns the file name and line number of the caller
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil {
		if err := Initialize(); err != nil {
			return
		}
	}

	caller := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	logger.Printf("[DEBUG] [%s] %s", caller, message)
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil {
		if err := Initialize(); err != nil {
			return
		}
	}

	caller := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	logger.Printf("[INFO] [%s] %s", caller, message)
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil {
		if err := Initialize(); err != nil {
			return
		}
	}

	caller := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	logger.Printf("[WARN] [%s] %s", caller, message)
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil {
		if err := Initialize(); err != nil {
			return
		}
	}

	caller := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	logger.Printf("[ERROR] [%s] %s", caller, message)
}
