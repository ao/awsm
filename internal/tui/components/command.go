package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Command represents a command that can be executed from the command palette
type Command struct {
	Name        string
	Description string
	Action      func() error
}

// CommandPalette represents a command palette component
type CommandPalette struct {
	textInput textinput.Model
	commands  []Command
	filtered  []Command
	active    bool
	width     int
	height    int
	style     lipgloss.Style
}

// NewCommandPalette creates a new command palette
func NewCommandPalette() *CommandPalette {
	ti := textinput.New()
	ti.Placeholder = "Type a command..."
	ti.CharLimit = 100
	ti.Width = 40
	ti.Prompt = ": "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0066cc"))

	return &CommandPalette{
		textInput: ti,
		commands:  []Command{},
		filtered:  []Command{},
		active:    false,
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(1, 2),
	}
}

// SetSize sets the size of the command palette
func (c *CommandPalette) SetSize(width, height int) {
	c.width = width
	c.height = height
	c.textInput.Width = width - 10 // Account for padding and prompt
}

// AddCommand adds a command to the command palette
func (c *CommandPalette) AddCommand(name, description string, action func() error) {
	c.commands = append(c.commands, Command{
		Name:        name,
		Description: description,
		Action:      action,
	})
}

// SetActive sets whether the command palette is active
func (c *CommandPalette) SetActive(active bool) {
	c.active = active
	if active {
		c.textInput.Focus()
		c.filter("")
	} else {
		c.textInput.Blur()
	}
}

// IsActive returns whether the command palette is active
func (c *CommandPalette) IsActive() bool {
	return c.active
}

// Toggle toggles whether the command palette is active
func (c *CommandPalette) Toggle() {
	c.SetActive(!c.active)
}

// Filter filters the commands based on the input
func (c *CommandPalette) Filter(input string) {
	c.filter(input)
}

// filter is the internal implementation of Filter
func (c *CommandPalette) filter(input string) {
	if input == "" {
		c.filtered = c.commands
		return
	}

	input = strings.ToLower(input)
	var filtered []Command

	for _, cmd := range c.commands {
		if strings.Contains(strings.ToLower(cmd.Name), input) ||
			strings.Contains(strings.ToLower(cmd.Description), input) {
			filtered = append(filtered, cmd)
		}
	}

	c.filtered = filtered
}

// GetSelectedCommand returns the selected command
func (c *CommandPalette) GetSelectedCommand() *Command {
	if len(c.filtered) == 0 {
		return nil
	}

	return &c.filtered[0]
}

// ExecuteSelected executes the selected command
func (c *CommandPalette) ExecuteSelected() error {
	cmd := c.GetSelectedCommand()
	if cmd == nil {
		return nil
	}

	return cmd.Action()
}

// HandleInput handles input for the command palette
func (c *CommandPalette) HandleInput(msg tea.Msg) {
	var cmd tea.Cmd
	c.textInput, cmd = c.textInput.Update(msg)
	_ = cmd // Ignore the command for now
	c.filter(c.textInput.Value())
}

// Render renders the command palette
func (c *CommandPalette) Render() string {
	if !c.active {
		return ""
	}

	// Create a title for the command palette
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0066cc")).
		Padding(0, 1).
		Render(" Command Palette ")

	// Render the text input
	input := c.textInput.View()

	// Render the filtered commands
	var commandsView string
	if len(c.filtered) == 0 {
		commandsView = "No commands found"
	} else {
		var commandLines []string
		maxCommands := 10 // Maximum number of commands to show
		for i, cmd := range c.filtered {
			if i >= maxCommands {
				break
			}

			// Highlight the first command
			style := lipgloss.NewStyle()
			if i == 0 {
				style = style.Bold(true).Foreground(lipgloss.Color("#0066cc"))
			}

			line := style.Render(cmd.Name + " - " + cmd.Description)
			commandLines = append(commandLines, line)
		}
		commandsView = strings.Join(commandLines, "\n")
	}

	// Create a styled command palette with a border
	paletteView := c.style.Copy().
		Width(c.width - 4). // Account for border width
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			input,
			"",
			commandsView,
		))

	// Combine title and palette view
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		paletteView,
	)
}
