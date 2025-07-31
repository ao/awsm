package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ao/awsm/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigInitialize tests the initialization of the configuration system.
// It verifies that:
// 1. The configuration file is created when Initialize() is called
// 2. Default configuration values are set correctly
//
// This test uses a temporary directory to avoid modifying the actual configuration.
func TestConfigInitialize(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := config.ConfigFile

	// Set the config file to a temporary file
	config.ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		config.ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = config.Initialize()
	require.NoError(t, err)

	// Check if the config file was created
	_, err = os.Stat(filepath.Join(tempDir, ".awsm.yaml"))
	assert.NoError(t, err)

	// Check if the default values were set
	assert.Equal(t, config.DefaultConfig.AWS.Profile, config.GetAWSProfile())
	assert.Equal(t, config.DefaultConfig.AWS.Region, config.GetAWSRegion())
	assert.Equal(t, config.DefaultConfig.Output.Format, config.GetOutputFormat())
}

// TestConfigSave tests saving and loading the configuration.
// It verifies that:
// 1. Configuration changes are persisted to the config file
// 2. Configuration values can be loaded from the file on re-initialization
//
// This test modifies AWS profile, region, and output format settings
// and ensures they are correctly saved and reloaded.
func TestConfigSave(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := config.ConfigFile

	// Set the config file to a temporary file
	config.ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		config.ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = config.Initialize()
	require.NoError(t, err)

	// Modify the configuration
	err = config.SetAWSProfile("test-profile")
	require.NoError(t, err)
	err = config.SetAWSRegion("us-west-2")
	require.NoError(t, err)
	err = config.SetOutputFormat("json")
	require.NoError(t, err)

	// Check if the values were set
	assert.Equal(t, "test-profile", config.GetAWSProfile())
	assert.Equal(t, "us-west-2", config.GetAWSRegion())
	assert.Equal(t, "json", config.GetOutputFormat())

	// Re-initialize the configuration to load from the file
	err = config.Initialize()
	require.NoError(t, err)

	// Check if the values were loaded from the file
	assert.Equal(t, "test-profile", config.GetAWSProfile())
	assert.Equal(t, "us-west-2", config.GetAWSRegion())
	assert.Equal(t, "json", config.GetOutputFormat())
}

// TestContextSwitching tests the context switching functionality.
// It verifies that:
// 1. New contexts can be created with specific profiles and regions
// 2. Switching between contexts updates the current AWS profile and region
// 3. Multiple contexts can be managed simultaneously
//
// This test creates multiple contexts and switches between them,
// verifying that the correct settings are applied each time.
func TestContextSwitching(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := config.ConfigFile

	// Set the config file to a temporary file
	config.ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		config.ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = config.Initialize()
	require.NoError(t, err)

	// Create a new context
	err = config.CreateContext("test-context", "test-profile", "us-west-2", "")
	require.NoError(t, err)

	// Switch to the new context
	err = config.SetCurrentContext("test-context")
	require.NoError(t, err)

	// Check if the current context was set
	assert.Equal(t, "test-context", config.GetCurrentContext())
	assert.Equal(t, "test-profile", config.GetAWSProfile())
	assert.Equal(t, "us-west-2", config.GetAWSRegion())

	// Create another context
	err = config.CreateContext("another-context", "another-profile", "eu-west-1", "")
	require.NoError(t, err)

	// Switch to the other context
	err = config.SetCurrentContext("another-context")
	require.NoError(t, err)

	// Check if the current context was set
	assert.Equal(t, "another-context", config.GetCurrentContext())
	assert.Equal(t, "another-profile", config.GetAWSProfile())
	assert.Equal(t, "eu-west-1", config.GetAWSRegion())

	// Switch back to the first context
	err = config.SetCurrentContext("test-context")
	require.NoError(t, err)

	// Check if the current context was set
	assert.Equal(t, "test-context", config.GetCurrentContext())
	assert.Equal(t, "test-profile", config.GetAWSProfile())
	assert.Equal(t, "us-west-2", config.GetAWSRegion())
}

// TestContextManagement tests the context management functionality.
// It verifies that:
// 1. Contexts can be created with specific settings
// 2. Contexts can be updated with new settings
// 3. Contexts can be deleted
//
// This test performs a full CRUD (Create, Read, Update, Delete) cycle
// on a test context and verifies each operation works correctly.
func TestContextManagement(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := config.ConfigFile

	// Set the config file to a temporary file
	config.ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		config.ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = config.Initialize()
	require.NoError(t, err)

	// Create a new context
	err = config.CreateContext("test-context", "test-profile", "us-west-2", "")
	require.NoError(t, err)

	// Check if the context was created
	contexts := config.GetContexts()
	assert.Contains(t, contexts, "test-context")
	assert.Equal(t, "test-profile", contexts["test-context"].Profile)
	assert.Equal(t, "us-west-2", contexts["test-context"].Region)
	assert.Equal(t, "", contexts["test-context"].Role)

	// Update the context
	err = config.UpdateContext("test-context", "updated-profile", "us-east-1", "test-role")
	require.NoError(t, err)

	// Check if the context was updated
	contexts = config.GetContexts()
	assert.Contains(t, contexts, "test-context")
	assert.Equal(t, "updated-profile", contexts["test-context"].Profile)
	assert.Equal(t, "us-east-1", contexts["test-context"].Region)
	assert.Equal(t, "test-role", contexts["test-context"].Role)

	// Delete the context
	err = config.DeleteContext("test-context")
	require.NoError(t, err)

	// Check if the context was deleted
	contexts = config.GetContexts()
	assert.NotContains(t, contexts, "test-context")
}

// TestContextInfo tests retrieving context information.
// It verifies that:
// 1. Current context information can be retrieved accurately
// 2. The list of all contexts can be retrieved
// 3. Context information includes all expected fields (name, profile, region, role, current status)
//
// This test creates a context, sets it as current, and then verifies
// that both the current context info and the context list contain the correct information.
func TestContextInfo(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := config.ConfigFile

	// Set the config file to a temporary file
	config.ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		config.ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = config.Initialize()
	require.NoError(t, err)

	// Create a new context
	err = config.CreateContext("test-context", "test-profile", "us-west-2", "test-role")
	require.NoError(t, err)

	// Switch to the new context
	err = config.SetCurrentContext("test-context")
	require.NoError(t, err)

	// Get the current context info
	contextInfo, err := config.GetCurrentContextInfo()
	require.NoError(t, err)

	// Check if the context info is correct
	assert.Equal(t, "test-context", contextInfo.Name)
	assert.Equal(t, "test-profile", contextInfo.Profile)
	assert.Equal(t, "us-west-2", contextInfo.Region)
	assert.Equal(t, "test-role", contextInfo.Role)
	assert.True(t, contextInfo.Current)

	// List all contexts
	contextList := config.ListContexts()

	// Check if the context list is correct
	assert.Len(t, contextList, 2) // default + test-context

	// Find the test context in the list
	var found bool
	for _, ctx := range contextList {
		if ctx.Name == "test-context" {
			found = true
			assert.Equal(t, "test-profile", ctx.Profile)
			assert.Equal(t, "us-west-2", ctx.Region)
			assert.Equal(t, "test-role", ctx.Role)
			assert.True(t, ctx.Current)
		}
	}
	assert.True(t, found, "test-context not found in context list")
}
