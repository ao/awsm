// Package config provides configuration management for the awsm application.
package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ContextInfo provides detailed information about a context including whether
// it is the current active context.
type ContextInfo struct {
	Name    string // Name of the context
	Profile string // AWS profile associated with the context
	Region  string // AWS region associated with the context
	Role    string // AWS role ARN associated with the context (optional)
	Current bool   // Whether this is the current active context
}

// ListContexts returns a list of all available contexts with detailed information.
// The current context will have its Current field set to true.
func ListContexts() []ContextInfo {
	contexts := GetContexts()
	currentContext := GetCurrentContext()
	result := make([]ContextInfo, 0, len(contexts))

	for name, ctx := range contexts {
		result = append(result, ContextInfo{
			Name:    name,
			Profile: ctx.Profile,
			Region:  ctx.Region,
			Role:    ctx.Role,
			Current: name == currentContext,
		})
	}

	return result
}

// SwitchContext switches to the specified context, updating the current AWS profile,
// region, and role based on the context's configuration.
//
// Returns an error if the context doesn't exist or if the configuration cannot be saved.
func SwitchContext(name string) error {
	return SetCurrentContext(name)
}

// NewContext creates a new context with the given parameters.
//
// The name, profile, and region parameters are required.
// The role parameter is optional and can be an empty string.
//
// Returns an error if any of the required parameters are empty or if the
// configuration cannot be saved.
func NewContext(name, profile, region, role string) error {
	// Validate name
	if name == "" {
		return fmt.Errorf("context name cannot be empty")
	}

	// Validate profile
	if profile == "" {
		return fmt.Errorf("profile cannot be empty")
	}

	// Validate region
	if region == "" {
		return fmt.Errorf("region cannot be empty")
	}

	// Create the context
	return CreateContext(name, profile, region, role)
}

// RemoveContext removes the specified context.
//
// Returns an error if the context doesn't exist, if it's the current context,
// or if the configuration cannot be saved.
func RemoveContext(name string) error {
	return DeleteContext(name)
}

// ImportContextsFromAWS imports contexts from the AWS config file.
//
// It reads the AWS config file and creates contexts for each profile found.
// The context names are prefixed with "aws:" followed by the profile name.
//
// Returns the number of contexts imported and an error if the AWS config file
// cannot be read or if the contexts cannot be created.
func ImportContextsFromAWS() (int, error) {
	configPath, err := GetAWSConfigPath()
	if err != nil {
		return 0, fmt.Errorf("failed to get AWS config path: %w", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("AWS config file not found at %s", configPath)
	}

	// Open the config file
	file, err := os.Open(configPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open AWS config file: %w", err)
	}
	defer file.Close()

	// Parse the config file
	scanner := bufio.NewScanner(file)
	profileRegex := regexp.MustCompile(`^\[profile\s+(.+)\]$`)
	regionRegex := regexp.MustCompile(`^\s*region\s*=\s*(.+)$`)
	roleRegex := regexp.MustCompile(`^\s*role_arn\s*=\s*(.+)$`)

	var currentProfile string
	var currentRegion string
	var currentRole string
	importCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Check for profile
		if matches := profileRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Save previous profile if complete
			if currentProfile != "" && currentRegion != "" {
				contextName := fmt.Sprintf("aws:%s", currentProfile)
				if err := CreateContext(contextName, currentProfile, currentRegion, currentRole); err == nil {
					importCount++
				}
			}

			// Start new profile
			currentProfile = matches[1]
			currentRegion = ""
			currentRole = ""
			continue
		}

		// Check for region
		if matches := regionRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentRegion = matches[1]
			continue
		}

		// Check for role
		if matches := roleRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentRole = matches[1]
			continue
		}
	}

	// Save the last profile if complete
	if currentProfile != "" && currentRegion != "" {
		contextName := fmt.Sprintf("aws:%s", currentProfile)
		if err := CreateContext(contextName, currentProfile, currentRegion, currentRole); err == nil {
			importCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return importCount, fmt.Errorf("error reading AWS config file: %w", err)
	}

	return importCount, nil
}

// ExportContextsToAWS exports contexts to the AWS config file.
//
// If overwrite is true, it will overwrite the existing AWS config file.
// If overwrite is false, it will return an error if the AWS config file already exists.
//
// Returns an error if the AWS config file cannot be created or written to.
func ExportContextsToAWS(overwrite bool) error {
	configPath, err := GetAWSConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get AWS config path: %w", err)
	}

	// Check if config file exists and we're not overwriting
	if !overwrite {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("AWS config file already exists at %s and overwrite is false", configPath)
		}
	}

	// Create config directory if it doesn't exist
	configDir := strings.TrimSuffix(configPath, "/config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create AWS config directory: %w", err)
	}

	// Create or open the config file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create AWS config file: %w", err)
	}
	defer file.Close()

	// Write header
	file.WriteString("# AWS Config file generated by awsm\n\n")

	// Write default profile
	file.WriteString("[default]\n")
	file.WriteString(fmt.Sprintf("region = %s\n\n", GetAWSRegion()))

	// Write contexts as profiles
	contexts := GetContexts()
	for name, ctx := range contexts {
		// Skip default context as it's already written
		if name == "default" {
			continue
		}

		file.WriteString(fmt.Sprintf("[profile %s]\n", ctx.Profile))
		file.WriteString(fmt.Sprintf("region = %s\n", ctx.Region))
		if ctx.Role != "" {
			file.WriteString(fmt.Sprintf("role_arn = %s\n", ctx.Role))
		}
		file.WriteString("\n")
	}

	return nil
}

// GetCurrentContextInfo returns detailed information about the current context.
//
// Returns an error if the current context doesn't exist.
func GetCurrentContextInfo() (ContextInfo, error) {
	currentName := GetCurrentContext()
	contexts := GetContexts()

	ctx, exists := contexts[currentName]
	if !exists {
		return ContextInfo{}, fmt.Errorf("current context %s not found", currentName)
	}

	return ContextInfo{
		Name:    currentName,
		Profile: ctx.Profile,
		Region:  ctx.Region,
		Role:    ctx.Role,
		Current: true,
	}, nil
}
