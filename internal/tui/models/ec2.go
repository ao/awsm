package models

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ao/awsm/internal/aws/ec2"
	"github.com/ao/awsm/internal/logger"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EC2InstanceMsg is a message containing EC2 instance data
type EC2InstanceMsg struct {
	Instances []ec2.Instance
	Error     error
}

// EC2Model represents the EC2 view
type EC2Model struct {
	BaseModel
	title            string
	instances        []ec2.Instance
	selected         int
	loading          bool
	err              error
	adapter          *ec2.Adapter
	loadingStartTime time.Time
	loadingTimeout   time.Duration
}

// NewEC2Model creates a new EC2 model
func NewEC2Model() *EC2Model {
	return &EC2Model{
		BaseModel:      NewBaseModel(),
		title:          "EC2 Instances",
		instances:      []ec2.Instance{},
		selected:       0,
		loading:        false,
		loadingTimeout: 30 * time.Second, // Default timeout of 30 seconds
	}
}

// SetLoadingTimeout sets the timeout duration for loading operations
func (m *EC2Model) SetLoadingTimeout(timeout time.Duration) {
	m.loadingTimeout = timeout
}

// Init initializes the model
func (m *EC2Model) Init() tea.Cmd {
	logger.Debug("EC2Model.Init called")
	m.loading = true
	m.loadingStartTime = time.Now()

	// Create a debug file to verify this function is being called
	f, _ := os.Create("ec2_init_debug.log")
	if f != nil {
		f.WriteString("EC2Model.Init called\n")
		f.Close()
	}

	// Directly call loadInstances and handle the result
	result := m.loadInstances()

	// Log the result
	f2, _ := os.Create("ec2_init_result.log")
	if f2 != nil {
		if msg, ok := result.(EC2InstanceMsg); ok {
			if msg.Error != nil {
				f2.WriteString(fmt.Sprintf("Error: %v\n", msg.Error))
			} else {
				f2.WriteString(fmt.Sprintf("Instances: %d\n", len(msg.Instances)))
			}
		} else {
			f2.WriteString(fmt.Sprintf("Unknown result type: %T\n", result))
		}
		f2.Close()
	}

	// Return a command that returns the result directly
	return func() tea.Msg {
		logger.Debug("Returning EC2InstanceMsg from Init command")
		return result
	}
}

// checkTimeout checks if the loading operation has timed out
func (m *EC2Model) checkTimeout() tea.Msg {
	if !m.loading || m.loadingStartTime.IsZero() {
		return nil
	}

	// Check if we've exceeded the timeout
	if time.Since(m.loadingStartTime) > m.loadingTimeout {
		return TimeoutMsg{
			Message: "Operation timed out",
			Source:  "EC2Model",
		}
	}

	// Schedule another check in 1 second
	time.Sleep(1 * time.Second)

	// Continue checking for timeout
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return m.checkTimeout()
	})
}

// loadInstances loads EC2 instances
func (m *EC2Model) loadInstances() tea.Msg {
	logger.Debug("EC2Model.loadInstances called")

	// Set a timeout to ensure we don't get stuck in a loading state
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create EC2 adapter if not already created
	if m.adapter == nil {
		logger.Debug("Creating EC2 adapter")
		adapter, err := ec2.NewAdapter(ctx)
		if err != nil {
			logger.Error("Error creating EC2 adapter: %v", err)

			// Return a more user-friendly error message
			if strings.Contains(err.Error(), "InvalidAccessKeyId") {
				errMsg := "invalid AWS credentials: the access key ID is invalid or expired"
				logger.Error(errMsg)
				return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "ExpiredToken") {
				errMsg := "expired AWS credentials: please refresh your credentials"
				logger.Error(errMsg)
				return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "AccessDenied") {
				errMsg := "access denied: your AWS credentials don't have permission to access EC2"
				logger.Error(errMsg)
				return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
				errMsg := "connection timeout: unable to connect to AWS"
				logger.Error(errMsg)
				return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
			}
			return EC2InstanceMsg{Error: err}
		}
		logger.Debug("EC2 adapter created successfully")
		m.adapter = adapter
	}

	// List EC2 instances with timeout
	logger.Debug("Listing EC2 instances")
	instances, err := m.adapter.ListInstances(ctx, nil, 0)
	if err != nil {
		logger.Error("Error listing EC2 instances: %v", err)

		// Return a more user-friendly error message
		if strings.Contains(err.Error(), "InvalidAccessKeyId") {
			errMsg := "invalid AWS credentials: the access key ID is invalid or expired"
			logger.Error(errMsg)
			return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
		} else if strings.Contains(err.Error(), "ExpiredToken") {
			errMsg := "expired AWS credentials: please refresh your credentials"
			logger.Error(errMsg)
			return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
		} else if strings.Contains(err.Error(), "AccessDenied") {
			errMsg := "access denied: your AWS credentials don't have permission to access EC2"
			logger.Error(errMsg)
			return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
		} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
			errMsg := "connection timeout: unable to connect to AWS"
			logger.Error(errMsg)
			return EC2InstanceMsg{Error: fmt.Errorf(errMsg)}
		}
	} else {
		logger.Info("Found %d EC2 instances", len(instances))
	}

	return EC2InstanceMsg{
		Instances: instances,
		Error:     err,
	}
}

// Update updates the model based on messages
func (m *EC2Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case EC2InstanceMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
			return m, nil
		}
		m.instances = msg.Instances
		m.err = nil
		return m, nil

	case TimeoutMsg:
		if msg.Source == "EC2Model" && m.loading {
			m.loading = false
			m.err = fmt.Errorf("operation timed out after %v", m.loadingTimeout)
			return m, nil
		}

	case tea.KeyMsg:
		// Handle key messages
		switch {
		case key.Matches(msg, DefaultKeyMap().Up):
			if m.selected > 0 {
				m.selected--
			}
		case key.Matches(msg, DefaultKeyMap().Down):
			if m.selected < len(m.instances)-1 {
				m.selected++
			}
		case key.Matches(msg, DefaultKeyMap().Enter):
			// View details of selected instance
			// (In a real implementation, this would show a detailed view)
		case key.Matches(msg, DefaultKeyMap().Refresh):
			m.loading = true
			return m, m.loadInstances
		}
	}

	return m, nil
}

// View renders the model
func (m *EC2Model) View() string {
	// Create a title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0066cc")).
		Padding(0, 1).
		Render(fmt.Sprintf(" %s ", m.title))

	// Create content
	var content string
	if m.loading {
		elapsed := time.Since(m.loadingStartTime).Round(time.Second)
		if elapsed > 5*time.Second {
			content = fmt.Sprintf("Loading EC2 instances... (%s)", elapsed)
		} else {
			content = "Loading EC2 instances..."
		}
	} else if m.err != nil {
		content = fmt.Sprintf("Error: %s\n\nPress 'r' to retry or 'd' to go to dashboard", m.err.Error())
	} else if len(m.instances) == 0 {
		content = "No EC2 instances found"
	} else {
		// Create a table header
		header := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render("ID\tNAME\tSTATE\tTYPE\tPUBLIC IP")

		// Create table rows
		var rows []string
		for i, instance := range m.instances {
			style := lipgloss.NewStyle()
			if i == m.selected {
				style = style.
					Bold(true).
					Foreground(lipgloss.Color("#FFFFFF")).
					Background(lipgloss.Color("#0066cc"))
			}

			row := style.Render(fmt.Sprintf(
				"%s\t%s\t%s\t%s\t%s",
				instance.ID,
				instance.Name,
				instance.State,
				instance.Type,
				instance.PublicIP,
			))
			rows = append(rows, row)
		}

		// Combine header and rows
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			strings.Join(rows, "\n"),
		)
	}

	// Add help text
	helpText := "\nPress ↑/↓ to navigate, Enter to view details, r to refresh, ? for help"

	// Style the content
	styledContent := lipgloss.NewStyle().
		Padding(1, 2).
		Render(content + helpText)

	// Combine title and content
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		styledContent,
	)
}

// ShortHelp returns the short help text
func (m *EC2Model) ShortHelp() []key.Binding {
	return []key.Binding{
		DefaultKeyMap().Help,
		DefaultKeyMap().Quit,
		DefaultKeyMap().Up,
		DefaultKeyMap().Down,
		DefaultKeyMap().Enter,
		DefaultKeyMap().Refresh,
		DefaultKeyMap().Dashboard,
		DefaultKeyMap().Command,
	}
}

// FullHelp returns the full help text
func (m *EC2Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			DefaultKeyMap().Help,
			DefaultKeyMap().Quit,
			DefaultKeyMap().Command,
		},
		{
			DefaultKeyMap().Up,
			DefaultKeyMap().Down,
			DefaultKeyMap().Enter,
		},
		{
			DefaultKeyMap().Refresh,
			DefaultKeyMap().Dashboard,
		},
	}
}
