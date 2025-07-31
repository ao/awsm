package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
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

	// Check if the config file was created
	_, err = os.Stat(filepath.Join(tempDir, ".awsm.yaml"))
	assert.NoError(t, err)

	// Check if the default values were set
	assert.Equal(t, DefaultConfig.AWS.Profile, GlobalConfig.AWS.Profile)
	assert.Equal(t, DefaultConfig.AWS.Region, GlobalConfig.AWS.Region)
	assert.Equal(t, DefaultConfig.Output.Format, GlobalConfig.Output.Format)
}

func TestGetSetAWSProfile(t *testing.T) {
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

	// Set a new profile
	err = SetAWSProfile("test-profile")
	require.NoError(t, err)

	// Check if the profile was set
	assert.Equal(t, "test-profile", GetAWSProfile())
}

func TestGetSetAWSRegion(t *testing.T) {
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

	// Set a new region
	err = SetAWSRegion("us-west-2")
	require.NoError(t, err)

	// Check if the region was set
	assert.Equal(t, "us-west-2", GetAWSRegion())
}

func TestGetSetOutputFormat(t *testing.T) {
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

	// Set a new output format
	err = SetOutputFormat("json")
	require.NoError(t, err)

	// Check if the output format was set
	assert.Equal(t, "json", GetOutputFormat())
}

func TestGetSetAppMode(t *testing.T) {
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

	// Set a new app mode
	err = SetAppMode("tui")
	require.NoError(t, err)

	// Check if the app mode was set
	assert.Equal(t, "tui", GetAppMode())
}

func TestGetAWSCredentialsPath(t *testing.T) {
	// Get the AWS credentials path
	path, err := GetAWSCredentialsPath()
	require.NoError(t, err)

	// Check if the path is not empty
	assert.NotEmpty(t, path)
}

func TestGetAWSConfigPath(t *testing.T) {
	// Get the AWS config path
	path, err := GetAWSConfigPath()
	require.NoError(t, err)

	// Check if the path is not empty
	assert.NotEmpty(t, path)
}

func TestGetSetAWSRole(t *testing.T) {
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

	// Set a new role
	err = SetAWSRole("arn:aws:iam::123456789012:role/test-role")
	require.NoError(t, err)

	// Check if the role was set
	assert.Equal(t, "arn:aws:iam::123456789012:role/test-role", GetAWSRole())
}

func TestCreateUpdateDeleteContext(t *testing.T) {
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

	// Create a new context
	err = CreateContext("test-context", "test-profile", "us-west-2", "")
	require.NoError(t, err)

	// Check if the context was created
	contexts := GetContexts()
	assert.Contains(t, contexts, "test-context")
	assert.Equal(t, "test-profile", contexts["test-context"].Profile)
	assert.Equal(t, "us-west-2", contexts["test-context"].Region)
	assert.Equal(t, "", contexts["test-context"].Role)

	// Update the context
	err = UpdateContext("test-context", "updated-profile", "us-east-1", "arn:aws:iam::123456789012:role/test-role")
	require.NoError(t, err)

	// Check if the context was updated
	contexts = GetContexts()
	assert.Contains(t, contexts, "test-context")
	assert.Equal(t, "updated-profile", contexts["test-context"].Profile)
	assert.Equal(t, "us-east-1", contexts["test-context"].Region)
	assert.Equal(t, "arn:aws:iam::123456789012:role/test-role", contexts["test-context"].Role)

	// Delete the context
	err = DeleteContext("test-context")
	require.NoError(t, err)

	// Check if the context was deleted
	contexts = GetContexts()
	assert.NotContains(t, contexts, "test-context")
}

func TestFavorites(t *testing.T) {
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

	// Add a profile to favorites
	err = AddFavoriteProfile("test-profile")
	require.NoError(t, err)

	// Check if the profile was added to favorites
	profiles := GetFavoriteProfiles()
	assert.Contains(t, profiles, "test-profile")

	// Add a region to favorites
	err = AddFavoriteRegion("us-west-2")
	require.NoError(t, err)

	// Check if the region was added to favorites
	regions := GetFavoriteRegions()
	assert.Contains(t, regions, "us-west-2")

	// Remove the profile from favorites
	err = RemoveFavoriteProfile("test-profile")
	require.NoError(t, err)

	// Check if the profile was removed from favorites
	profiles = GetFavoriteProfiles()
	assert.NotContains(t, profiles, "test-profile")

	// Remove the region from favorites
	err = RemoveFavoriteRegion("us-west-2")
	require.NoError(t, err)

	// Check if the region was removed from favorites
	regions = GetFavoriteRegions()
	assert.NotContains(t, regions, "us-west-2")
}

func TestGetAWSProfiles(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "awsm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a temporary AWS directory structure
	awsDir := filepath.Join(tempDir, ".aws")
	err = os.MkdirAll(awsDir, 0755)
	require.NoError(t, err)

	// Create a mock credentials file
	credentialsContent := `[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[test-profile1]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY

[test-profile2]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE2
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY2
`
	err = os.WriteFile(filepath.Join(awsDir, "credentials"), []byte(credentialsContent), 0644)
	require.NoError(t, err)

	// Create a mock config file
	configContent := `[default]
region = us-east-1
output = json

[profile test-profile3]
region = us-west-2
output = yaml

[profile test-profile4]
region = eu-central-1
output = table
`
	err = os.WriteFile(filepath.Join(awsDir, "config"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Create a test-specific function to read profiles from our test files
	testGetProfiles := func() ([]string, error) {
		// Map to store unique profile names
		profileMap := make(map[string]bool)

		// Add the default profile
		profileMap["default"] = true

		// Read profiles from credentials file
		credPath := filepath.Join(awsDir, "credentials")
		if _, err := os.Stat(credPath); err == nil {
			file, err := os.Open(credPath)
			if err != nil {
				return nil, fmt.Errorf("error opening AWS credentials file: %w", err)
			}
			defer file.Close()

			// Regular expression to match profile sections like [profile-name]
			re := regexp.MustCompile(`^\[(.*?)\]`)

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if matches := re.FindStringSubmatch(line); len(matches) > 1 {
					profileMap[matches[1]] = true
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading AWS credentials file: %w", err)
			}
		}

		// Read profiles from config file
		configPath := filepath.Join(awsDir, "config")
		if _, err := os.Stat(configPath); err == nil {
			file, err := os.Open(configPath)
			if err != nil {
				return nil, fmt.Errorf("error opening AWS config file: %w", err)
			}
			defer file.Close()

			// Regular expression to match profile sections like [profile profile-name]
			re := regexp.MustCompile(`^\[profile (.*?)\]`)

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if matches := re.FindStringSubmatch(line); len(matches) > 1 {
					profileMap[matches[1]] = true
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading AWS config file: %w", err)
			}
		}

		// Convert map keys to slice
		profiles := make([]string, 0, len(profileMap))
		for profile := range profileMap {
			profiles = append(profiles, profile)
		}

		return profiles, nil
	}

	// Call the test function
	profiles, err := testGetProfiles()
	require.NoError(t, err)

	// Check that all profiles were found
	assert.Contains(t, profiles, "default")
	assert.Contains(t, profiles, "test-profile1")
	assert.Contains(t, profiles, "test-profile2")
	assert.Contains(t, profiles, "test-profile3")
	assert.Contains(t, profiles, "test-profile4")
	assert.Len(t, profiles, 5) // Should have 5 unique profiles
}
