package models

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ao/awsm/internal/aws/lambda"
	"github.com/ao/awsm/internal/logger"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LambdaFunctionMsg is a message containing Lambda function data
type LambdaFunctionMsg struct {
	Functions []lambda.Function
	Error     error
}

// LambdaLogMsg is a message containing Lambda function logs
type LambdaLogMsg struct {
	Logs  []lambda.LogEvent
	Error error
}

// LambdaModel represents the Lambda view
type LambdaModel struct {
	BaseModel
	title            string
	functions        []lambda.Function
	logs             []lambda.LogEvent
	selected         int
	viewingLogs      bool
	currentFunction  string
	loading          bool
	err              error
	adapter          *lambda.Adapter
	loadingStartTime time.Time
	loadingTimeout   time.Duration
}

// NewLambdaModel creates a new Lambda model
func NewLambdaModel() *LambdaModel {
	return &LambdaModel{
		BaseModel:      NewBaseModel(),
		title:          "Lambda Functions",
		functions:      []lambda.Function{},
		logs:           []lambda.LogEvent{},
		selected:       0,
		viewingLogs:    false,
		loading:        false,
		loadingTimeout: 30 * time.Second, // Default timeout of 30 seconds
	}
}

// SetLoadingTimeout sets the timeout duration for loading operations
func (m *LambdaModel) SetLoadingTimeout(timeout time.Duration) {
	m.loadingTimeout = timeout
}

// Init initializes the model
func (m *LambdaModel) Init() tea.Cmd {
	logger.Debug("LambdaModel.Init called")
	m.loading = true
	m.loadingStartTime = time.Now()

	// Create a debug file to verify this function is being called
	f, _ := os.Create("lambda_init_debug.log")
	if f != nil {
		f.WriteString("LambdaModel.Init called\n")
		f.Close()
	}

	// Directly call loadFunctions and handle the result
	result := m.loadFunctions()

	// Log the result
	f2, _ := os.Create("lambda_init_result.log")
	if f2 != nil {
		if msg, ok := result.(LambdaFunctionMsg); ok {
			if msg.Error != nil {
				f2.WriteString(fmt.Sprintf("Error: %v\n", msg.Error))
			} else {
				f2.WriteString(fmt.Sprintf("Functions: %d\n", len(msg.Functions)))
			}
		} else {
			f2.WriteString(fmt.Sprintf("Unknown result type: %T\n", result))
		}
		f2.Close()
	}

	// Return a command that returns the result directly
	return func() tea.Msg {
		logger.Debug("Returning LambdaFunctionMsg from Init command")
		return result
	}
}

// checkTimeout checks if the loading operation has timed out
func (m *LambdaModel) checkTimeout() tea.Msg {
	if !m.loading || m.loadingStartTime.IsZero() {
		return nil
	}

	// Check if we've exceeded the timeout
	if time.Since(m.loadingStartTime) > m.loadingTimeout {
		return TimeoutMsg{
			Message: "Operation timed out",
			Source:  "LambdaModel",
		}
	}

	// Schedule another check in 1 second
	time.Sleep(1 * time.Second)

	// Continue checking for timeout
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return m.checkTimeout()
	})
}

// loadFunctions loads Lambda functions
func (m *LambdaModel) loadFunctions() tea.Msg {
	// Set a timeout to ensure we don't get stuck in a loading state
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create Lambda adapter if not already created
	if m.adapter == nil {
		adapter, err := lambda.NewAdapter(ctx)
		if err != nil {
			// Return a more user-friendly error message
			if strings.Contains(err.Error(), "InvalidAccessKeyId") {
				return LambdaFunctionMsg{Error: fmt.Errorf("invalid AWS credentials: the access key ID is invalid or expired")}
			} else if strings.Contains(err.Error(), "ExpiredToken") {
				return LambdaFunctionMsg{Error: fmt.Errorf("expired AWS credentials: please refresh your credentials")}
			} else if strings.Contains(err.Error(), "AccessDenied") {
				return LambdaFunctionMsg{Error: fmt.Errorf("access denied: your AWS credentials don't have permission to access Lambda")}
			} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
				return LambdaFunctionMsg{Error: fmt.Errorf("connection timeout: unable to connect to AWS")}
			}
			return LambdaFunctionMsg{Error: err}
		}
		m.adapter = adapter
	}

	// List Lambda functions with proper error handling
	functions, err := m.adapter.ListFunctions(ctx, 0)
	if err != nil {
		// Return a more user-friendly error message
		if strings.Contains(err.Error(), "InvalidAccessKeyId") {
			return LambdaFunctionMsg{Error: fmt.Errorf("invalid AWS credentials: the access key ID is invalid or expired")}
		} else if strings.Contains(err.Error(), "ExpiredToken") {
			return LambdaFunctionMsg{Error: fmt.Errorf("expired AWS credentials: please refresh your credentials")}
		} else if strings.Contains(err.Error(), "AccessDenied") {
			return LambdaFunctionMsg{Error: fmt.Errorf("access denied: your AWS credentials don't have permission to access Lambda")}
		} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
			return LambdaFunctionMsg{Error: fmt.Errorf("connection timeout: unable to connect to AWS")}
		}
		return LambdaFunctionMsg{Error: err}
	}

	// Return the functions
	return LambdaFunctionMsg{
		Functions: functions,
		Error:     nil,
	}
}

// loadLogs loads logs for the current Lambda function
func (m *LambdaModel) loadLogs() tea.Cmd {
	return func() tea.Msg {
		logger.Debug("LambdaModel.loadLogs called for function: %s", m.currentFunction)

		// Set a timeout to ensure we don't get stuck in a loading state
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if m.adapter == nil {
			logger.Debug("Creating Lambda adapter")
			adapter, err := lambda.NewAdapter(ctx)
			if err != nil {
				logger.Error("Error creating Lambda adapter: %v", err)

				// Return a more user-friendly error message
				if strings.Contains(err.Error(), "InvalidAccessKeyId") {
					errMsg := "invalid AWS credentials: the access key ID is invalid or expired"
					logger.Error(errMsg)
					return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
				} else if strings.Contains(err.Error(), "ExpiredToken") {
					errMsg := "expired AWS credentials: please refresh your credentials"
					logger.Error(errMsg)
					return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
				} else if strings.Contains(err.Error(), "AccessDenied") {
					errMsg := "access denied: your AWS credentials don't have permission to access Lambda logs"
					logger.Error(errMsg)
					return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
				} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
					errMsg := "connection timeout: unable to connect to AWS"
					logger.Error(errMsg)
					return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
				}
				return LambdaLogMsg{Error: err}
			}
			logger.Debug("Lambda adapter created successfully")
			m.adapter = adapter
		}

		// Get logs for the current function (last 100 events)
		logger.Debug("Getting logs for function: %s", m.currentFunction)
		logs, err := m.adapter.GetFunctionLogs(ctx, m.currentFunction, time.Time{}, 100)
		if err != nil {
			logger.Error("Error getting logs for function %s: %v", m.currentFunction, err)

			// Return a more user-friendly error message
			if strings.Contains(err.Error(), "InvalidAccessKeyId") {
				errMsg := "invalid AWS credentials: the access key ID is invalid or expired"
				logger.Error(errMsg)
				return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "ExpiredToken") {
				errMsg := "expired AWS credentials: please refresh your credentials"
				logger.Error(errMsg)
				return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "AccessDenied") {
				errMsg := "access denied: your AWS credentials don't have permission to access Lambda logs"
				logger.Error(errMsg)
				return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
				errMsg := "connection timeout: unable to connect to AWS"
				logger.Error(errMsg)
				return LambdaLogMsg{Error: fmt.Errorf(errMsg)}
			}
			return LambdaLogMsg{Error: err}
		} else {
			logger.Info("Found %d log events for function %s", len(logs), m.currentFunction)
		}

		return LambdaLogMsg{
			Logs:  logs,
			Error: nil,
		}
	}
}

// Update updates the model based on messages
func (m *LambdaModel) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case LambdaFunctionMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
			return m, nil
		}
		m.functions = msg.Functions
		m.err = nil
		return m, nil

	case LambdaLogMsg:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
			return m, nil
		}
		m.logs = msg.Logs
		m.err = nil
		return m, nil

	case TimeoutMsg:
		if msg.Source == "LambdaModel" && m.loading {
			m.loading = false
			m.err = fmt.Errorf("operation timed out after %v", m.loadingTimeout)
			return m, nil
		}

	case tea.KeyMsg:
		// Handle key messages
		switch {
		case key.Matches(msg, DefaultKeyMap().Up):
			if m.viewingLogs {
				// No selection in logs view
			} else {
				if m.selected > 0 {
					m.selected--
				}
			}
		case key.Matches(msg, DefaultKeyMap().Down):
			if m.viewingLogs {
				// No selection in logs view
			} else {
				if m.selected < len(m.functions)-1 {
					m.selected++
				}
			}
		case key.Matches(msg, DefaultKeyMap().Enter):
			if !m.viewingLogs && len(m.functions) > 0 {
				// View logs for the selected function
				m.viewingLogs = true
				m.currentFunction = m.functions[m.selected].Name
				m.title = fmt.Sprintf("Lambda Logs: %s", m.currentFunction)
				m.loading = true
				return m, m.loadLogs()
			}
		case key.Matches(msg, DefaultKeyMap().Escape):
			if m.viewingLogs {
				// Go back to function list
				m.viewingLogs = false
				m.title = "Lambda Functions"
			}
		case key.Matches(msg, DefaultKeyMap().Refresh):
			m.loading = true
			if m.viewingLogs {
				return m, m.loadLogs()
			} else {
				return m, func() tea.Msg {
					return m.loadFunctions()
				}
			}
		}
	}

	return m, nil
}

// View renders the model
func (m *LambdaModel) View() string {
	// Create a title with consistent styling across all views
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
			content = fmt.Sprintf("Loading Lambda data... (%s)", elapsed)
		} else {
			content = "Loading Lambda data..."
		}
	} else if m.err != nil {
		content = fmt.Sprintf("Error: %s\n\nPress 'r' to retry or 'd' to go to dashboard", m.err.Error())
	} else if m.viewingLogs {
		if len(m.logs) == 0 {
			content = "No logs found for this function"
		} else {
			// Create log entries
			var logEntries []string
			for _, log := range m.logs {
				// Format timestamp
				timestamp := time.Unix(0, log.Timestamp*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")

				// Format log entry
				entry := fmt.Sprintf("[%s] %s", timestamp, log.Message)
				logEntries = append(logEntries, entry)
			}

			// Combine log entries
			content = strings.Join(logEntries, "\n")
		}
	} else {
		if len(m.functions) == 0 {
			content = "No Lambda functions found"
		} else {
			// Create a table header
			header := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render("NAME\tRUNTIME\tMEMORY\tTIMEOUT\tLAST MODIFIED")

			// Create table rows
			var rows []string
			for i, function := range m.functions {
				style := lipgloss.NewStyle()
				if i == m.selected {
					style = style.
						Bold(true).
						Foreground(lipgloss.Color("#FFFFFF")).
						Background(lipgloss.Color("#0066cc"))
				}

				row := style.Render(fmt.Sprintf(
					"%s\t%s\t%d MB\t%d sec\t%s",
					function.Name,
					function.Runtime,
					function.Memory,
					function.Timeout,
					function.LastModified,
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
	}

	// Add help text
	var helpText string
	if m.viewingLogs {
		helpText = "\nPress Esc to go back, r to refresh, ? for help"
	} else {
		helpText = "\nPress ↑/↓ to navigate, Enter to view logs, r to refresh, ? for help"
	}

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
func (m *LambdaModel) ShortHelp() []key.Binding {
	if m.viewingLogs {
		return []key.Binding{
			DefaultKeyMap().Help,
			DefaultKeyMap().Quit,
			DefaultKeyMap().Escape,
			DefaultKeyMap().Refresh,
			DefaultKeyMap().Dashboard,
			DefaultKeyMap().Command,
		}
	}
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
func (m *LambdaModel) FullHelp() [][]key.Binding {
	if m.viewingLogs {
		return [][]key.Binding{
			{
				DefaultKeyMap().Help,
				DefaultKeyMap().Quit,
				DefaultKeyMap().Command,
			},
			{
				DefaultKeyMap().Escape,
				DefaultKeyMap().Refresh,
				DefaultKeyMap().Dashboard,
			},
		}
	}
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
