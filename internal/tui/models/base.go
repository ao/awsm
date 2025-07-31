package models

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TimeoutMsg is a message sent when a loading operation times out
type TimeoutMsg struct {
	Message string
	Source  string
}

// Model is the interface that all TUI models must implement
type Model interface {
	// Init initializes the model
	Init() tea.Cmd

	// Update updates the model based on messages
	Update(msg tea.Msg) (Model, tea.Cmd)

	// View renders the model
	View() string

	// ShortHelp returns the short help text
	ShortHelp() []key.Binding

	// FullHelp returns the full help text
	FullHelp() [][]key.Binding

	// IsLoading returns whether the model is in a loading state
	IsLoading() bool

	// GetError returns any error that occurred during loading
	GetError() error
}

// BaseModel provides common functionality for all models
type BaseModel struct {
	Width            int
	Height           int
	loading          bool
	err              error
	loadingStartTime time.Time
	loadingTimeout   time.Duration
}

// NewBaseModel creates a new base model
func NewBaseModel() BaseModel {
	return BaseModel{
		loading:        false,
		err:            nil,
		loadingTimeout: 30 * time.Second,
	}
}

// SetSize sets the size of the model
func (m *BaseModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

// IsLoading returns whether the model is in a loading state
func (m *BaseModel) IsLoading() bool {
	return m.loading
}

// GetError returns any error that occurred during loading
func (m *BaseModel) GetError() error {
	return m.err
}

// SetLoading sets the loading state of the model
func (m *BaseModel) SetLoading(loading bool) {
	m.loading = loading
	if loading {
		m.loadingStartTime = time.Now()
	}
}

// SetError sets the error state of the model
func (m *BaseModel) SetError(err error) {
	m.err = err
	m.loading = false
}

// SetLoadingTimeout sets the timeout duration for loading operations
func (m *BaseModel) SetLoadingTimeout(timeout time.Duration) {
	m.loadingTimeout = timeout
}

// CheckTimeout checks if the loading has timed out and returns a command if it has
func (m *BaseModel) CheckTimeout() tea.Cmd {
	if !m.loading || m.loadingStartTime.IsZero() {
		return nil
	}

	if time.Since(m.loadingStartTime) > m.loadingTimeout {
		return func() tea.Msg {
			return TimeoutMsg{
				Message: "Operation timed out",
			}
		}
	}

	return nil
}

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Help      key.Binding
	Quit      key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
	Command   key.Binding
	Refresh   key.Binding
	Dashboard key.Binding
	EC2       key.Binding
	S3        key.Binding
	Lambda    key.Binding
	Context   key.Binding
	Profile   key.Binding
	Region    key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous"),
		),
		Command: key.NewBinding(
			key.WithKeys(":"),
			key.WithHelp(":", "command mode"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Dashboard: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "dashboard"),
		),
		EC2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "EC2"),
		),
		S3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "S3"),
		),
		Lambda: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "Lambda"),
		),
		Context: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "switch context"),
		),
		Profile: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "change profile"),
		),
		Region: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "change region"),
		),
	}
}
