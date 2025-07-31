// Package config provides configuration management for the awsm application.
//
// It handles loading, saving, and accessing configuration values such as AWS profiles,
// regions, output formats, and contexts. The configuration is stored in a YAML file
// in the user's home directory.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// AWS specific configuration
	AWS struct {
		Profile string
		Region  string
		Role    string
	}

	// Output configuration
	Output struct {
		Format string // json, yaml, table
	}

	// Application configuration
	App struct {
		Mode string // cli, tui
	}

	// Context configuration
	Contexts map[string]Context

	// Current context name
	CurrentContext string

	// Recent profiles and regions
	Recent struct {
		Profiles []string
		Regions  []string
	}

	// Favorite profiles and regions
	Favorites struct {
		Profiles []string
		Regions  []string
	}
}

// Context represents an AWS context (profile + region + optional role)
type Context struct {
	Profile string
	Region  string
	Role    string
}

var (
	// DefaultConfig holds the default configuration values
	DefaultConfig = Config{
		AWS: struct {
			Profile string
			Region  string
			Role    string
		}{
			Profile: "default",
			Region:  "us-east-1",
			Role:    "",
		},
		Output: struct {
			Format string
		}{
			Format: "table",
		},
		App: struct {
			Mode string
		}{
			Mode: "cli",
		},
		Contexts: map[string]Context{
			"default": {
				Profile: "default",
				Region:  "us-east-1",
				Role:    "",
			},
		},
		CurrentContext: "default",
		Recent: struct {
			Profiles []string
			Regions  []string
		}{
			Profiles: []string{"default"},
			Regions:  []string{"us-east-1"},
		},
		Favorites: struct {
			Profiles []string
			Regions  []string
		}{
			Profiles: []string{},
			Regions:  []string{},
		},
	}

	// ConfigFile is the name of the configuration file
	ConfigFile = ".awsm"

	// ConfigType is the type of the configuration file
	ConfigType = "yaml"

	// GlobalConfig holds the global configuration instance
	GlobalConfig Config
)

// Initialize initializes the configuration system by loading the configuration file
// and setting default values. If the configuration file doesn't exist, it creates
// a new one with default values.
//
// Returns an error if the configuration file cannot be loaded or created.
func Initialize() error {
	// Find home directory
	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("error finding home directory: %w", err)
	}

	// Set default configuration values
	viper.SetDefault("aws.profile", DefaultConfig.AWS.Profile)
	viper.SetDefault("aws.region", DefaultConfig.AWS.Region)
	viper.SetDefault("aws.role", DefaultConfig.AWS.Role)
	viper.SetDefault("output.format", DefaultConfig.Output.Format)
	viper.SetDefault("app.mode", DefaultConfig.App.Mode)
	viper.SetDefault("contexts", DefaultConfig.Contexts)
	viper.SetDefault("currentContext", DefaultConfig.CurrentContext)
	viper.SetDefault("recent.profiles", DefaultConfig.Recent.Profiles)
	viper.SetDefault("recent.regions", DefaultConfig.Recent.Regions)
	viper.SetDefault("favorites.profiles", DefaultConfig.Favorites.Profiles)
	viper.SetDefault("favorites.regions", DefaultConfig.Favorites.Regions)

	// Set configuration file name and type
	viper.SetConfigName(ConfigFile)
	viper.SetConfigType(ConfigType)

	// Add configuration search paths
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("AWSM")

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		// If the configuration file doesn't exist, create it with default values
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configPath := filepath.Join(home, ConfigFile+"."+ConfigType)
			if err := viper.SafeWriteConfigAs(configPath); err != nil {
				return fmt.Errorf("error creating default configuration file: %w", err)
			}
			fmt.Printf("Created default configuration file at %s\n", configPath)
		} else {
			return fmt.Errorf("error reading configuration file: %w", err)
		}
	}

	// Unmarshal configuration into GlobalConfig
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("error unmarshaling configuration: %w", err)
	}

	return nil
}

// Save persists the current configuration to the configuration file.
//
// Returns an error if the configuration file cannot be written.
func Save() error {
	return viper.WriteConfig()
}

// GetAWSProfile returns the currently configured AWS profile name.
func GetAWSProfile() string {
	return GlobalConfig.AWS.Profile
}

// SetAWSProfile sets the AWS profile to use for AWS API calls.
//
// The profile must exist in the AWS credentials file.
// Returns an error if the configuration cannot be saved.
func SetAWSProfile(profile string) error {
	GlobalConfig.AWS.Profile = profile
	viper.Set("aws.profile", profile)
	return Save()
}

// GetAWSRegion returns the currently configured AWS region.
func GetAWSRegion() string {
	return GlobalConfig.AWS.Region
}

// SetAWSRegion sets the AWS region to use for AWS API calls.
//
// Returns an error if the configuration cannot be saved.
func SetAWSRegion(region string) error {
	GlobalConfig.AWS.Region = region
	viper.Set("aws.region", region)
	return Save()
}

// GetOutputFormat returns the currently configured output format (json, yaml, table, etc.).
func GetOutputFormat() string {
	return GlobalConfig.Output.Format
}

// SetOutputFormat sets the output format for command results.
//
// Valid formats are json, yaml, table, and text.
// Returns an error if the configuration cannot be saved.
func SetOutputFormat(format string) error {
	GlobalConfig.Output.Format = format
	viper.Set("output.format", format)
	return Save()
}

// GetAppMode returns the currently configured application mode (cli or tui).
func GetAppMode() string {
	return GlobalConfig.App.Mode
}

// SetAppMode sets the application mode (cli or tui).
//
// Returns an error if the configuration cannot be saved.
func SetAppMode(mode string) error {
	GlobalConfig.App.Mode = mode
	viper.Set("app.mode", mode)
	return Save()
}

// GetAWSCredentialsPath returns the path to the AWS credentials file.
//
// Returns an error if the home directory cannot be determined.
func GetAWSCredentialsPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}

	return filepath.Join(home, ".aws", "credentials"), nil
}

// GetAWSConfigPath returns the path to the AWS config file.
//
// Returns an error if the home directory cannot be determined.
func GetAWSConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("error finding home directory: %w", err)
	}

	return filepath.Join(home, ".aws", "config"), nil
}

// CheckAWSCredentials checks if AWS credentials are available by verifying
// the existence of the AWS credentials file.
//
// Returns true if the credentials file exists, false otherwise.
// Returns an error if there was a problem checking the file.
func CheckAWSCredentials() (bool, error) {
	credPath, err := GetAWSCredentialsPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(credPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetAWSRole returns the currently configured AWS role ARN.
func GetAWSRole() string {
	return GlobalConfig.AWS.Role
}

// SetAWSRole sets the AWS role ARN to assume for AWS API calls.
//
// Returns an error if the configuration cannot be saved.
func SetAWSRole(role string) error {
	GlobalConfig.AWS.Role = role
	viper.Set("aws.role", role)
	return Save()
}

// GetCurrentContext returns the name of the currently active context.
func GetCurrentContext() string {
	return GlobalConfig.CurrentContext
}

// SetCurrentContext sets the current context and updates AWS profile, region, and role
// based on the context's configuration.
//
// Returns an error if the context doesn't exist or if the configuration cannot be saved.
func SetCurrentContext(contextName string) error {
	// Check if context exists
	context, exists := GlobalConfig.Contexts[contextName]
	if !exists {
		return fmt.Errorf("context %s does not exist", contextName)
	}

	// Update current context
	GlobalConfig.CurrentContext = contextName
	viper.Set("currentContext", contextName)

	// Update AWS profile and region
	GlobalConfig.AWS.Profile = context.Profile
	viper.Set("aws.profile", context.Profile)
	GlobalConfig.AWS.Region = context.Region
	viper.Set("aws.region", context.Region)
	GlobalConfig.AWS.Role = context.Role
	viper.Set("aws.role", context.Role)

	// Add to recent profiles and regions
	addToRecent("profiles", context.Profile)
	addToRecent("regions", context.Region)

	return Save()
}

// GetContexts returns all available contexts as a map of context name to Context.
func GetContexts() map[string]Context {
	return GlobalConfig.Contexts
}

// CreateContext creates a new context with the specified name, profile, region, and role.
//
// Returns an error if the configuration cannot be saved.
func CreateContext(name, profile, region, role string) error {
	// Create the context
	GlobalConfig.Contexts[name] = Context{
		Profile: profile,
		Region:  region,
		Role:    role,
	}
	viper.Set("contexts", GlobalConfig.Contexts)

	// Add to recent profiles and regions
	addToRecent("profiles", profile)
	addToRecent("regions", region)

	return Save()
}

// UpdateContext updates an existing context with new profile, region, and role values.
//
// If the context is the current context, the AWS profile, region, and role are also updated.
// Returns an error if the context doesn't exist or if the configuration cannot be saved.
func UpdateContext(name, profile, region, role string) error {
	// Check if context exists
	if _, exists := GlobalConfig.Contexts[name]; !exists {
		return fmt.Errorf("context %s does not exist", name)
	}

	// Update the context
	GlobalConfig.Contexts[name] = Context{
		Profile: profile,
		Region:  region,
		Role:    role,
	}
	viper.Set("contexts", GlobalConfig.Contexts)

	// If this is the current context, update AWS profile and region
	if GlobalConfig.CurrentContext == name {
		GlobalConfig.AWS.Profile = profile
		viper.Set("aws.profile", profile)
		GlobalConfig.AWS.Region = region
		viper.Set("aws.region", region)
		GlobalConfig.AWS.Role = role
		viper.Set("aws.role", role)
	}

	// Add to recent profiles and regions
	addToRecent("profiles", profile)
	addToRecent("regions", region)

	return Save()
}

// DeleteContext deletes a context with the specified name.
//
// Returns an error if the context doesn't exist, if it's the current context,
// or if the configuration cannot be saved.
func DeleteContext(name string) error {
	// Check if context exists
	if _, exists := GlobalConfig.Contexts[name]; !exists {
		return fmt.Errorf("context %s does not exist", name)
	}

	// Cannot delete the current context
	if GlobalConfig.CurrentContext == name {
		return fmt.Errorf("cannot delete the current context")
	}

	// Delete the context
	delete(GlobalConfig.Contexts, name)
	viper.Set("contexts", GlobalConfig.Contexts)

	return Save()
}

// GetRecentProfiles returns the list of recently used AWS profiles.
func GetRecentProfiles() []string {
	return GlobalConfig.Recent.Profiles
}

// GetRecentRegions returns the list of recently used AWS regions.
func GetRecentRegions() []string {
	return GlobalConfig.Recent.Regions
}

// GetFavoriteProfiles returns the list of favorite AWS profiles.
func GetFavoriteProfiles() []string {
	return GlobalConfig.Favorites.Profiles
}

// GetFavoriteRegions returns the list of favorite AWS regions.
func GetFavoriteRegions() []string {
	return GlobalConfig.Favorites.Regions
}

// AddFavoriteProfile adds an AWS profile to the favorites list.
//
// If the profile is already in the favorites list, this is a no-op.
// Returns an error if the configuration cannot be saved.
func AddFavoriteProfile(profile string) error {
	// Check if already in favorites
	for _, p := range GlobalConfig.Favorites.Profiles {
		if p == profile {
			return nil // Already a favorite
		}
	}

	// Add to favorites
	GlobalConfig.Favorites.Profiles = append(GlobalConfig.Favorites.Profiles, profile)
	viper.Set("favorites.profiles", GlobalConfig.Favorites.Profiles)

	return Save()
}

// RemoveFavoriteProfile removes an AWS profile from the favorites list.
//
// Returns an error if the profile is not in the favorites list or if the
// configuration cannot be saved.
func RemoveFavoriteProfile(profile string) error {
	// Find and remove the profile
	for i, p := range GlobalConfig.Favorites.Profiles {
		if p == profile {
			GlobalConfig.Favorites.Profiles = append(
				GlobalConfig.Favorites.Profiles[:i],
				GlobalConfig.Favorites.Profiles[i+1:]...,
			)
			viper.Set("favorites.profiles", GlobalConfig.Favorites.Profiles)
			return Save()
		}
	}

	return fmt.Errorf("profile %s is not in favorites", profile)
}

// AddFavoriteRegion adds an AWS region to the favorites list.
//
// If the region is already in the favorites list, this is a no-op.
// Returns an error if the configuration cannot be saved.
func AddFavoriteRegion(region string) error {
	// Check if already in favorites
	for _, r := range GlobalConfig.Favorites.Regions {
		if r == region {
			return nil // Already a favorite
		}
	}

	// Add to favorites
	GlobalConfig.Favorites.Regions = append(GlobalConfig.Favorites.Regions, region)
	viper.Set("favorites.regions", GlobalConfig.Favorites.Regions)

	return Save()
}

// RemoveFavoriteRegion removes an AWS region from the favorites list.
//
// Returns an error if the region is not in the favorites list or if the
// configuration cannot be saved.
func RemoveFavoriteRegion(region string) error {
	// Find and remove the region
	for i, r := range GlobalConfig.Favorites.Regions {
		if r == region {
			GlobalConfig.Favorites.Regions = append(
				GlobalConfig.Favorites.Regions[:i],
				GlobalConfig.Favorites.Regions[i+1:]...,
			)
			viper.Set("favorites.regions", GlobalConfig.Favorites.Regions)
			return Save()
		}
	}

	return fmt.Errorf("region %s is not in favorites", region)
}

// addToRecent adds an item to the recent list (profiles or regions).
//
// If the item is already in the list, it is moved to the front.
// The list is limited to 10 items.
func addToRecent(listType string, item string) {
	var list *[]string

	// Determine which list to update
	switch listType {
	case "profiles":
		list = &GlobalConfig.Recent.Profiles
	case "regions":
		list = &GlobalConfig.Recent.Regions
	default:
		return
	}

	// Check if already in the list
	for i, existing := range *list {
		if existing == item {
			// Move to front if not already there
			if i > 0 {
				// Remove from current position
				*list = append((*list)[:i], (*list)[i+1:]...)
				// Add to front
				*list = append([]string{item}, *list...)
			}
			return
		}
	}

	// Add to front of list
	*list = append([]string{item}, *list...)

	// Limit list size to 10 items
	if len(*list) > 10 {
		*list = (*list)[:10]
	}

	// Update viper
	switch listType {
	case "profiles":
		viper.Set("recent.profiles", *list)
	case "regions":
		viper.Set("recent.regions", *list)
	}
}

// GetAWSProfiles returns a list of all AWS profiles from the AWS config and credentials files.
//
// This function reads both the AWS config and credentials files and returns a list of all
// profile names found in either file. For the config file, it looks for sections like
// [profile name], and for the credentials file, it looks for sections like [name].
//
// Returns an error if there was a problem reading either file.
func GetAWSProfiles() ([]string, error) {
	// Map to store unique profile names
	profileMap := make(map[string]bool)

	// Add the default profile
	profileMap["default"] = true

	// Read profiles from credentials file
	credPath, err := GetAWSCredentialsPath()
	if err != nil {
		return nil, fmt.Errorf("error getting AWS credentials path: %w", err)
	}

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
	configPath, err := GetAWSConfigPath()
	if err != nil {
		return nil, fmt.Errorf("error getting AWS config path: %w", err)
	}

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
