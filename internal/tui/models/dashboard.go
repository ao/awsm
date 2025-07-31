package models

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// DashboardModel represents the dashboard view
type DashboardModel struct {
	BaseModel
	title string
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() *DashboardModel {
	return &DashboardModel{
		BaseModel: NewBaseModel(),
		title:     "Dashboard",
	}
}

// Init initializes the model
func (m *DashboardModel) Init() tea.Cmd {
	// Return a command to load dashboard data
	return func() tea.Msg {
		return nil // No data to load yet
	}
}

// Update updates the model based on messages
func (m *DashboardModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key messages
		switch {
		case key.Matches(msg, DefaultKeyMap().Up):
			// Handle up key
		case key.Matches(msg, DefaultKeyMap().Down):
			// Handle down key
		}
	}

	return m, nil
}

// View renders the model
func (m *DashboardModel) View() string {
	// Just return the content without styling, as the ResultsPanel will handle that
	return `AWS Resources Overview:

EC2 Instances: Press 2 to view
S3 Buckets: Press 3 to view
Lambda Functions: Press 4 to view

Press ? for help or : for command palette`
}

// ShortHelp returns the short help text
func (m *DashboardModel) ShortHelp() []key.Binding {
	return []key.Binding{
		DefaultKeyMap().Help,
		DefaultKeyMap().Quit,
		DefaultKeyMap().EC2,
		DefaultKeyMap().S3,
		DefaultKeyMap().Lambda,
		DefaultKeyMap().Command,
	}
}

// FullHelp returns the full help text
func (m *DashboardModel) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			DefaultKeyMap().Help,
			DefaultKeyMap().Quit,
			DefaultKeyMap().Command,
		},
		{
			DefaultKeyMap().EC2,
			DefaultKeyMap().S3,
			DefaultKeyMap().Lambda,
		},
	}
}
