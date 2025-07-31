// Package testutils provides utilities for testing the awsm application.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ao/awsm/internal/config"
	"github.com/stretchr/testify/require"
)

// SetupTestConfig creates a temporary configuration file for testing.
// It returns a cleanup function that should be deferred.
func SetupTestConfig(t *testing.T) func() {
	t.Helper()

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)

	// Save the original config file path
	originalConfigFile := config.ConfigFile

	// Set the config file to a temporary file
	config.ConfigFile = filepath.Join(tempDir, ".awsm")

	// Return a cleanup function
	return func() {
		// Restore the original config file path
		config.ConfigFile = originalConfigFile

		// Remove the temporary directory
		os.RemoveAll(tempDir)
	}
}

// CreateTempFile creates a temporary file with the given content.
// It returns the file path and a cleanup function that should be deferred.
func CreateTempFile(t *testing.T, content string) (string, func()) {
	t.Helper()

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer tempFile.Close()

	// Write the content to the file
	_, err = tempFile.WriteString(content)
	require.NoError(t, err)

	// Return the file path and a cleanup function
	return tempFile.Name(), func() {
		os.Remove(tempFile.Name())
	}
}

// CreateTempDir creates a temporary directory.
// It returns the directory path and a cleanup function that should be deferred.
func CreateTempDir(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)

	// Return the directory path and a cleanup function
	return tempDir, func() {
		os.RemoveAll(tempDir)
	}
}

// MockAWSCredentials creates mock AWS credentials for testing.
// It returns a cleanup function that should be deferred.
func MockAWSCredentials(t *testing.T) func() {
	t.Helper()

	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-aws-*")
	require.NoError(t, err)

	// Create a mock AWS credentials file
	credentialsContent := `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[test-profile]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
`
	credentialsPath := filepath.Join(tempDir, "credentials")
	err = os.WriteFile(credentialsPath, []byte(credentialsContent), 0600)
	require.NoError(t, err)

	// Create a mock AWS config file
	configContent := `[default]
region = us-east-1
output = json

[profile test-profile]
region = us-west-2
output = json
`
	configPath := filepath.Join(tempDir, "config")
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Store original paths for restoration in cleanup function
	// We'll implement a proper override mechanism in the actual tests

	// Return a cleanup function
	return func() {
		// Remove the temporary directory
		os.RemoveAll(tempDir)
	}
}
