// Package components provides tests for the TUI components.
package components

import (
	"testing"

	"github.com/ao/awsm/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestNewStatusBar tests the NewStatusBar constructor function.
// It verifies that a new StatusBar is created with the expected default values.
func TestNewStatusBar(t *testing.T) {
	// Create a new status bar
	statusBar := NewStatusBar()

	// Assert status bar is not nil
	assert.NotNil(t, statusBar)

	// Assert default values
	assert.Equal(t, 0, statusBar.width)
	assert.NotNil(t, statusBar.style)
}

// TestStatusBarSetWidth tests the SetWidth method of the StatusBar.
// It verifies that the width is correctly set on the StatusBar instance.
func TestStatusBarSetWidth(t *testing.T) {
	// Create a new status bar
	statusBar := NewStatusBar()

	// Set width
	statusBar.SetWidth(100)

	// Assert width was set
	assert.Equal(t, 100, statusBar.width)
}

// TestStatusBarRender tests the Render method of the StatusBar.
// It verifies that the rendered output contains the expected information
// from the current configuration, including context, profile, region, and role.
// This test uses the actual configuration values rather than mocking them.
func TestStatusBarRender(t *testing.T) {
	// Save original config values
	originalContext := config.GetCurrentContext()
	originalProfile := config.GetAWSProfile()
	originalRegion := config.GetAWSRegion()
	originalRole := config.GetAWSRole()

	// Create a new status bar
	statusBar := NewStatusBar()

	// Set width
	statusBar.SetWidth(200)

	// Render the status bar
	result := statusBar.Render()

	// Assert result is not empty
	assert.NotEmpty(t, result)

	// Assert result contains the expected values
	assert.Contains(t, result, originalContext)
	assert.Contains(t, result, originalProfile)
	assert.Contains(t, result, originalRegion)
	if originalRole != "" {
		assert.Contains(t, result, originalRole)
	}
	assert.Contains(t, result, "Status:")
	assert.Contains(t, result, "? for help")
}
