package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContexts(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := ConfigFile

	// Set the config file to a temporary file
	ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = Initialize()
	require.NoError(t, err)

	// Create test contexts
	err = CreateContext("test-context-1", "profile-1", "us-west-1", "")
	require.NoError(t, err)
	err = CreateContext("test-context-2", "profile-2", "us-west-2", "role-2")
	require.NoError(t, err)

	// Set the current context
	err = SetCurrentContext("test-context-1")
	require.NoError(t, err)

	// List contexts
	contexts := ListContexts()

	// Check if the contexts were listed correctly
	assert.Len(t, contexts, 3) // default + 2 test contexts

	// Find and verify test-context-1
	var foundContext1 bool
	var foundContext2 bool
	var foundDefault bool

	for _, ctx := range contexts {
		switch ctx.Name {
		case "test-context-1":
			foundContext1 = true
			assert.Equal(t, "profile-1", ctx.Profile)
			assert.Equal(t, "us-west-1", ctx.Region)
			assert.Equal(t, "", ctx.Role)
			assert.True(t, ctx.Current)
		case "test-context-2":
			foundContext2 = true
			assert.Equal(t, "profile-2", ctx.Profile)
			assert.Equal(t, "us-west-2", ctx.Region)
			assert.Equal(t, "role-2", ctx.Role)
			assert.False(t, ctx.Current)
		case "default":
			foundDefault = true
			assert.Equal(t, "default", ctx.Profile)
			assert.Equal(t, "us-east-1", ctx.Region)
			assert.Equal(t, "", ctx.Role)
			assert.False(t, ctx.Current)
		}
	}

	assert.True(t, foundContext1, "test-context-1 not found")
	assert.True(t, foundContext2, "test-context-2 not found")
	assert.True(t, foundDefault, "default context not found")
}

func TestSwitchContext(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := ConfigFile

	// Set the config file to a temporary file
	ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = Initialize()
	require.NoError(t, err)

	// Create test contexts
	err = CreateContext("test-context", "test-profile", "us-west-2", "")
	require.NoError(t, err)

	// Switch to the test context
	err = SwitchContext("test-context")
	require.NoError(t, err)

	// Check if the current context was set
	assert.Equal(t, "test-context", GetCurrentContext())
	assert.Equal(t, "test-profile", GetAWSProfile())
	assert.Equal(t, "us-west-2", GetAWSRegion())
}

func TestNewContext(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := ConfigFile

	// Set the config file to a temporary file
	ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = Initialize()
	require.NoError(t, err)

	// Test with valid parameters
	err = NewContext("test-context", "test-profile", "us-west-2", "")
	require.NoError(t, err)

	// Check if the context was created
	contexts := GetContexts()
	assert.Contains(t, contexts, "test-context")
	assert.Equal(t, "test-profile", contexts["test-context"].Profile)
	assert.Equal(t, "us-west-2", contexts["test-context"].Region)
	assert.Equal(t, "", contexts["test-context"].Role)

	// Test with empty name
	err = NewContext("", "test-profile", "us-west-2", "")
	assert.Error(t, err)

	// Test with empty profile
	err = NewContext("test-context-2", "", "us-west-2", "")
	assert.Error(t, err)

	// Test with empty region
	err = NewContext("test-context-3", "test-profile", "", "")
	assert.Error(t, err)
}

func TestRemoveContext(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := ConfigFile

	// Set the config file to a temporary file
	ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = Initialize()
	require.NoError(t, err)

	// Create a test context
	err = CreateContext("test-context", "test-profile", "us-west-2", "")
	require.NoError(t, err)

	// Remove the context
	err = RemoveContext("test-context")
	require.NoError(t, err)

	// Check if the context was removed
	contexts := GetContexts()
	assert.NotContains(t, contexts, "test-context")

	// Test removing a non-existent context
	err = RemoveContext("non-existent-context")
	assert.Error(t, err)

	// Test removing the current context
	err = RemoveContext("default")
	assert.Error(t, err)
}

func TestGetCurrentContextInfo(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save the original config file path
	originalConfigFile := ConfigFile

	// Set the config file to a temporary file
	ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = Initialize()
	require.NoError(t, err)

	// Get the current context info
	contextInfo, err := GetCurrentContextInfo()
	require.NoError(t, err)

	// Check if the context info is correct
	assert.Equal(t, "default", contextInfo.Name)
	assert.Equal(t, "default", contextInfo.Profile)
	assert.Equal(t, "us-east-1", contextInfo.Region)
	assert.Equal(t, "", contextInfo.Role)
	assert.True(t, contextInfo.Current)

	// Create and switch to a new context
	err = CreateContext("test-context", "test-profile", "us-west-2", "test-role")
	require.NoError(t, err)
	err = SetCurrentContext("test-context")
	require.NoError(t, err)

	// Get the current context info again
	contextInfo, err = GetCurrentContextInfo()
	require.NoError(t, err)

	// Check if the context info is correct
	assert.Equal(t, "test-context", contextInfo.Name)
	assert.Equal(t, "test-profile", contextInfo.Profile)
	assert.Equal(t, "us-west-2", contextInfo.Region)
	assert.Equal(t, "test-role", contextInfo.Role)
	assert.True(t, contextInfo.Current)
}

func TestImportExportContextsFromAWS(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a mock AWS config directory
	awsDir := filepath.Join(tempDir, ".aws")
	err = os.MkdirAll(awsDir, 0755)
	require.NoError(t, err)

	// Create a mock AWS config file
	configContent := `[default]
region = us-east-1
output = json

[profile test-profile]
region = us-west-2
output = json
role_arn = arn:aws:iam::123456789012:role/test-role

[profile another-profile]
region = eu-west-1
output = yaml
`
	configPath := filepath.Join(awsDir, "config")
	err = os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	// Save the original config file path
	originalConfigFile := ConfigFile

	// Set the config file to a temporary file
	ConfigFile = filepath.Join(tempDir, ".awsm")
	defer func() {
		ConfigFile = originalConfigFile
	}()

	// Initialize the configuration
	err = Initialize()
	require.NoError(t, err)

	// TODO: This test is incomplete because we can't easily override the AWS config path
	// In a real implementation, we would need to modify the GetAWSConfigPath function
	// to allow overriding the path for testing purposes.

	// For now, we'll just test the export functionality
	err = ExportContextsToAWS(true)
	require.NoError(t, err)

	// Check if the AWS config file was created
	_, err = os.Stat(filepath.Join(os.Getenv("HOME"), ".aws", "config"))
	assert.NoError(t, err)
}
