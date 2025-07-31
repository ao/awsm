// Package components provides tests for the TUI components.
package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestNewCommandPalette tests the NewCommandPalette constructor function.
// It verifies that a new CommandPalette is created with the expected default values,
// including being inactive and having empty command lists.
func TestNewCommandPalette(t *testing.T) {
	// Create a new command palette
	cp := NewCommandPalette()

	// Assert command palette is not nil
	assert.NotNil(t, cp)

	// Assert default values
	assert.False(t, cp.IsActive())
	assert.NotNil(t, cp.commands)
	assert.Len(t, cp.commands, 0)
	assert.NotNil(t, cp.filtered)
	assert.Len(t, cp.filtered, 0)
}

// TestCommandPaletteAddCommand tests the AddCommand method of the CommandPalette.
// It verifies that commands are correctly added to the command palette with
// the specified name, description, and action function.
func TestCommandPaletteAddCommand(t *testing.T) {
	// Create a new command palette
	cp := NewCommandPalette()

	// Add a command
	cp.AddCommand("test", "Test command", func() error {
		return nil
	})

	// Assert command was added
	assert.Len(t, cp.commands, 1)
	assert.Equal(t, "test", cp.commands[0].Name)
	assert.Equal(t, "Test command", cp.commands[0].Description)
	assert.NotNil(t, cp.commands[0].Action)
}

// TestCommandPaletteSetActive tests the SetActive method of the CommandPalette.
// It verifies that the active state of the command palette can be toggled
// and that the IsActive method correctly reports this state.
func TestCommandPaletteSetActive(t *testing.T) {
	// Create a new command palette
	cp := NewCommandPalette()

	// Set active
	cp.SetActive(true)

	// Assert active state
	assert.True(t, cp.IsActive())

	// Set inactive
	cp.SetActive(false)

	// Assert inactive state
	assert.False(t, cp.IsActive())
}

// TestCommandPaletteHandleInput tests the HandleInput method of the CommandPalette.
// It verifies that the command palette correctly handles keyboard input for:
// - Character input (filtering commands)
// - Backspace (removing characters and updating filters)
// - Escape key (deactivating the command palette)
func TestCommandPaletteHandleInput(t *testing.T) {
	// Create a new command palette
	cp := NewCommandPalette()

	// Add some commands
	cp.AddCommand("command1", "Command 1", func() error {
		return nil
	})
	cp.AddCommand("command2", "Command 2", func() error {
		return nil
	})
	cp.AddCommand("other", "Other command", func() error {
		return nil
	})

	// Set active
	cp.SetActive(true)

	// Test character input
	cp.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	assert.Equal(t, "c", cp.textInput.Value())
	assert.Len(t, cp.filtered, 2) // command1 and command2

	// Test more character input
	cp.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	assert.Equal(t, "co", cp.textInput.Value())
	assert.Len(t, cp.filtered, 1) // command1

	// Test backspace
	cp.HandleInput(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "c", cp.textInput.Value())
	assert.Len(t, cp.filtered, 2) // command1 and command2

	// Test escape
	cp.HandleInput(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, cp.IsActive())
}

// TestCommandPaletteExecuteSelected tests the ExecuteSelected method of the CommandPalette.
// It verifies that the selected command's action function is correctly executed
// when the ExecuteSelected method is called.
func TestCommandPaletteExecuteSelected(t *testing.T) {
	// Create a new command palette
	cp := NewCommandPalette()

	// Add a command
	executed := false
	cp.AddCommand("test", "Test command", func() error {
		executed = true
		return nil
	})

	// Set active and filter
	cp.SetActive(true)
	cp.HandleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})

	// Execute the selected command
	err := cp.ExecuteSelected()

	// Assert command was executed
	assert.NoError(t, err)
	assert.True(t, executed)
}

// TestCommandPaletteSetSize tests the SetSize method of the CommandPalette.
// It verifies that the width and height of the command palette are correctly set.
func TestCommandPaletteSetSize(t *testing.T) {
	// Create a new command palette
	cp := NewCommandPalette()

	// Set size
	cp.SetSize(100, 50)

	// Assert size was set
	assert.Equal(t, 100, cp.width)
	assert.Equal(t, 50, cp.height)
}

// TestCommandPaletteRender tests the Render method of the CommandPalette.
// It verifies that the rendered output contains the expected elements,
// including the command palette title and the list of commands with their descriptions.
func TestCommandPaletteRender(t *testing.T) {
	// Create a new command palette
	cp := NewCommandPalette()

	// Add some commands
	cp.AddCommand("command1", "Command 1", func() error {
		return nil
	})
	cp.AddCommand("command2", "Command 2", func() error {
		return nil
	})

	// Set active and size
	cp.SetActive(true)
	cp.SetSize(100, 50)

	// Render the command palette
	result := cp.Render()

	// Assert result is not empty
	assert.NotEmpty(t, result)

	// Assert result contains the command palette title
	assert.Contains(t, result, "Command Palette")

	// Assert result contains the commands
	assert.Contains(t, result, "command1")
	assert.Contains(t, result, "Command 1")
	assert.Contains(t, result, "command2")
	assert.Contains(t, result, "Command 2")
}
