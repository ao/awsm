package models

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ao/awsm/internal/aws/s3"
	"github.com/ao/awsm/internal/logger"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// S3BucketMsg is a message containing S3 bucket data
type S3BucketMsg struct {
	Buckets []s3.Bucket
	Error   error
}

// S3ObjectMsg is a message containing S3 object data
type S3ObjectMsg struct {
	Objects []s3.Object
	Error   error
}

// S3Model represents the S3 view
type S3Model struct {
	BaseModel
	title            string
	buckets          []s3.Bucket
	objects          []s3.Object
	selectedBucket   int
	selectedObject   int
	currentBucket    string
	viewingObjects   bool
	loading          bool
	err              error
	adapter          *s3.Adapter
	loadingStartTime time.Time
	loadingTimeout   time.Duration
}

// NewS3Model creates a new S3 model
func NewS3Model() *S3Model {
	logger.Debug("NewS3Model called")

	return &S3Model{
		BaseModel:      NewBaseModel(),
		title:          "S3 Buckets",
		buckets:        []s3.Bucket{},
		objects:        []s3.Object{},
		selectedBucket: 0,
		selectedObject: 0,
		viewingObjects: false,
		loading:        false,
		loadingTimeout: 30 * time.Second, // Default timeout of 30 seconds
	}
}

// IsLoading returns whether the model is in a loading state
func (m *S3Model) IsLoading() bool {
	return m.loading
}

// GetError returns any error that occurred during loading
func (m *S3Model) GetError() error {
	return m.err
}

// SetLoadingTimeout sets the timeout duration for loading operations
func (m *S3Model) SetLoadingTimeout(timeout time.Duration) {
	m.loadingTimeout = timeout
}

// Init initializes the model
func (m *S3Model) Init() tea.Cmd {
	logger.Debug("S3Model.Init called")

	m.loading = true
	m.loadingStartTime = time.Now()
	m.err = nil

	logger.Debug("S3Model.Init returning commands")

	// Create a debug file to verify this function is being called
	f, _ := os.Create("s3_init_debug_new.log")
	if f != nil {
		f.WriteString(fmt.Sprintf("S3Model.Init called at %s\n", time.Now().String()))
		f.Close()
	}

	// Return a command that will load buckets asynchronously
	return tea.Batch(
		m.asyncLoadBuckets,
		m.startTimeoutCheck,
	)
}

// startTimeoutCheck starts a periodic check for timeout
func (m *S3Model) startTimeoutCheck() tea.Msg {
	// Check every second if we've exceeded the timeout
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !m.loading {
				// If we're no longer loading, stop checking
				return nil
			}

			if time.Since(m.loadingStartTime) > m.loadingTimeout {
				// We've exceeded the timeout
				return TimeoutMsg{
					Message: "Operation timed out after " + m.loadingTimeout.String(),
					Source:  "S3Model",
				}
			}
		}
	}
}

// checkTimeout checks if the loading operation has timed out
func (m *S3Model) checkTimeout() tea.Msg {
	if !m.loading || m.loadingStartTime.IsZero() {
		return nil
	}

	// Check if we've exceeded the timeout
	if time.Since(m.loadingStartTime) > m.loadingTimeout {
		return TimeoutMsg{
			Message: "Operation timed out",
			Source:  "S3Model",
		}
	}

	// Schedule another check in 1 second
	time.Sleep(1 * time.Second)

	// Continue checking for timeout
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return m.checkTimeout()
	})
}

// asyncLoadBuckets loads S3 buckets asynchronously
func (m *S3Model) asyncLoadBuckets() tea.Msg {
	logger.Debug("S3Model.asyncLoadBuckets called")

	// Create a debug file to verify this function is being called
	f, _ := os.Create("s3_asyncload_debug.log")
	if f != nil {
		f.WriteString(fmt.Sprintf("S3Model.asyncLoadBuckets called at %s\n", time.Now().String()))
		f.Close()
	}

	// Use a longer timeout since we know the operation can take ~17.5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create S3 adapter if not already created
	if m.adapter == nil {
		logger.Debug("Creating S3 adapter")
		adapter, err := s3.NewAdapter(ctx)
		if err != nil {
			logger.Error("Error creating S3 adapter: %v", err)

			// Return a more user-friendly error message
			if strings.Contains(err.Error(), "InvalidAccessKeyId") {
				errMsg := "invalid AWS credentials: the access key ID is invalid or expired"
				logger.Error(errMsg)
				return S3BucketMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "ExpiredToken") {
				errMsg := "expired AWS credentials: please refresh your credentials"
				logger.Error(errMsg)
				return S3BucketMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "AccessDenied") {
				errMsg := "access denied: your AWS credentials don't have permission to access S3"
				logger.Error(errMsg)
				return S3BucketMsg{Error: fmt.Errorf(errMsg)}
			} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
				errMsg := "connection timeout: unable to connect to AWS"
				logger.Error(errMsg)
				return S3BucketMsg{Error: fmt.Errorf(errMsg)}
			}
			return S3BucketMsg{Error: err}
		}
		logger.Debug("S3 adapter created successfully")
		m.adapter = adapter
	}

	// List S3 buckets with timeout
	logger.Debug("Listing S3 buckets")
	buckets, err := m.adapter.ListBuckets(ctx)
	if err != nil {
		logger.Error("Error listing S3 buckets: %v", err)

		// Return a more user-friendly error message
		if strings.Contains(err.Error(), "InvalidAccessKeyId") {
			errMsg := "invalid AWS credentials: the access key ID is invalid or expired"
			logger.Error(errMsg)
			return S3BucketMsg{Error: fmt.Errorf(errMsg)}
		} else if strings.Contains(err.Error(), "ExpiredToken") {
			errMsg := "expired AWS credentials: please refresh your credentials"
			logger.Error(errMsg)
			return S3BucketMsg{Error: fmt.Errorf(errMsg)}
		} else if strings.Contains(err.Error(), "AccessDenied") {
			errMsg := "access denied: your AWS credentials don't have permission to access S3"
			logger.Error(errMsg)
			return S3BucketMsg{Error: fmt.Errorf(errMsg)}
		} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
			errMsg := "connection timeout: unable to connect to AWS"
			logger.Error(errMsg)
			return S3BucketMsg{Error: fmt.Errorf(errMsg)}
		}
	} else {
		logger.Info("Found %d S3 buckets", len(buckets))
		
		// Create a debug file with the results
		f, _ := os.Create("s3_buckets_result.log")
		if f != nil {
			f.WriteString(fmt.Sprintf("Found %d buckets at %s\n", len(buckets), time.Now().String()))
			for _, bucket := range buckets {
				f.WriteString(fmt.Sprintf("- %s (%s)\n", bucket.Name, bucket.Region))
			}
			f.Close()
		}
	}

	return S3BucketMsg{
		Buckets: buckets,
		Error:   err,
	}
}

// loadObjects loads S3 objects for the current bucket
func (m *S3Model) loadObjects() tea.Cmd {
	return func() tea.Msg {
		logger.Debug("S3Model.loadObjects called for bucket: %s", m.currentBucket)

		// Create a debug file to verify this function is being called
		f, _ := os.Create("s3_loadobjects_debug.log")
		if f != nil {
			f.WriteString(fmt.Sprintf("S3Model.loadObjects called for bucket %s at %s\n",
				m.currentBucket, time.Now().String()))
			f.Close()
		}

		// Use a longer timeout since we know the operation can take time
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if m.adapter == nil {
			logger.Debug("Creating S3 adapter")
			adapter, err := s3.NewAdapter(ctx)
			if err != nil {
				logger.Error("Error creating S3 adapter: %v", err)

				// Return a more user-friendly error message
				if strings.Contains(err.Error(), "InvalidAccessKeyId") {
					errMsg := "invalid AWS credentials: the access key ID is invalid or expired"
					logger.Error(errMsg)
					return S3ObjectMsg{Error: fmt.Errorf(errMsg)}
				} else if strings.Contains(err.Error(), "ExpiredToken") {
					errMsg := "expired AWS credentials: please refresh your credentials"
					logger.Error(errMsg)
					return S3ObjectMsg{Error: fmt.Errorf(errMsg)}
				} else if strings.Contains(err.Error(), "AccessDenied") {
					errMsg := "access denied: your AWS credentials don't have permission to access S3"
					logger.Error(errMsg)
					return S3ObjectMsg{Error: fmt.Errorf(errMsg)}
				} else if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
					errMsg := "connection timeout: unable to connect to AWS"
					logger.Error(errMsg)
					return S3ObjectMsg{Error: fmt.Errorf(errMsg)}
				}
				return S3ObjectMsg{Error: err}
			}
			logger.Debug("S3 adapter created successfully")
			m.adapter = adapter
		}

		// List objects in the current bucket
		logger.Debug("Listing objects in bucket: %s", m.currentBucket)
		objects, err := m.adapter.ListObjects(ctx, m.currentBucket, "", 0)
		if err != nil {
			logger.Error("Error listing objects in bucket %s: %v", m.currentBucket, err)
			
			// Create a debug file with the error
			f, _ := os.Create("s3_objects_error.log")
			if f != nil {
				f.WriteString(fmt.Sprintf("Error listing objects in bucket %s: %v at %s\n",
					m.currentBucket, err, time.Now().String()))
				f.Close()
			}
		} else {
			logger.Info("Found %d objects in bucket %s", len(objects), m.currentBucket)
			
			// Create a debug file with the results
			f, _ := os.Create("s3_objects_result.log")
			if f != nil {
				f.WriteString(fmt.Sprintf("Found %d objects in bucket %s at %s\n",
					len(objects), m.currentBucket, time.Now().String()))
				if len(objects) > 0 {
					f.WriteString("First 5 objects:\n")
					for i, obj := range objects {
						if i >= 5 {
							break
						}
						f.WriteString(fmt.Sprintf("- %s (%d bytes)\n", obj.Key, obj.Size))
					}
				}
				f.Close()
			}
		}

		return S3ObjectMsg{
			Objects: objects,
			Error:   err,
		}
	}
}

// Update updates the model based on messages
func (m *S3Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	logger.Debug("S3Model.Update called with message type: %T", msg)

	switch msg := msg.(type) {
	case S3BucketMsg:
		logger.Debug("Received S3BucketMsg")
		
		// Create a debug file to verify message handling
		f, _ := os.Create("s3_update_bucket_msg.log")
		if f != nil {
			if msg.Error != nil {
				f.WriteString(fmt.Sprintf("Error: %v at %s\n", msg.Error, time.Now().String()))
			} else {
				f.WriteString(fmt.Sprintf("Received %d buckets at %s\n", len(msg.Buckets), time.Now().String()))
			}
			f.Close()
		}
		
		m.loading = false
		if msg.Error != nil {
			logger.Error("S3BucketMsg error: %v", msg.Error)
			m.err = msg.Error
			return m, nil
		}
		
		logger.Debug("S3BucketMsg contains %d buckets", len(msg.Buckets))
		m.buckets = msg.Buckets
		m.err = nil
		return m, nil

	case S3ObjectMsg:
		logger.Debug("Received S3ObjectMsg")
		
		// Create a debug file to verify message handling
		f, _ := os.Create("s3_update_object_msg.log")
		if f != nil {
			if msg.Error != nil {
				f.WriteString(fmt.Sprintf("Error: %v at %s\n", msg.Error, time.Now().String()))
			} else {
				f.WriteString(fmt.Sprintf("Received %d objects at %s\n", len(msg.Objects), time.Now().String()))
			}
			f.Close()
		}
		
		m.loading = false
		if msg.Error != nil {
			logger.Error("S3ObjectMsg error: %v", msg.Error)
			m.err = msg.Error
			return m, nil
		}
		
		logger.Debug("S3ObjectMsg contains %d objects", len(msg.Objects))
		m.objects = msg.Objects
		m.err = nil
		return m, nil

	case TimeoutMsg:
		logger.Debug("Received TimeoutMsg: %s", msg.Message)
		if msg.Source == "S3Model" && m.loading {
			logger.Warn("S3Model operation timed out after %v", m.loadingTimeout)
			m.loading = false
			m.err = fmt.Errorf("operation timed out after %v", m.loadingTimeout)
			
			// Create a debug file for the timeout
			f, _ := os.Create("s3_timeout.log")
			if f != nil {
				f.WriteString(fmt.Sprintf("Operation timed out after %v at %s\n",
					m.loadingTimeout, time.Now().String()))
				f.Close()
			}
			
			return m, nil
		}

	case tea.KeyMsg:
		logger.Debug("Received KeyMsg: %s", msg.String())
		// Handle key messages
		switch {
		case key.Matches(msg, DefaultKeyMap().Up):
			if m.viewingObjects {
				if m.selectedObject > 0 {
					m.selectedObject--
				}
			} else {
				if m.selectedBucket > 0 {
					m.selectedBucket--
				}
			}
		case key.Matches(msg, DefaultKeyMap().Down):
			if m.viewingObjects {
				if m.selectedObject < len(m.objects)-1 {
					m.selectedObject++
				}
			} else {
				if m.selectedBucket < len(m.buckets)-1 {
					m.selectedBucket++
				}
			}
		case key.Matches(msg, DefaultKeyMap().Enter):
			if !m.viewingObjects && len(m.buckets) > 0 {
				// View objects in the selected bucket
				m.viewingObjects = true
				m.currentBucket = m.buckets[m.selectedBucket].Name
				m.title = fmt.Sprintf("S3 Objects: %s", m.currentBucket)
				m.loading = true
				m.loadingStartTime = time.Now()
				m.err = nil
				
				// Create a debug file for entering bucket view
				f, _ := os.Create("s3_enter_bucket.log")
				if f != nil {
					f.WriteString(fmt.Sprintf("Entering bucket %s at %s\n",
						m.currentBucket, time.Now().String()))
					f.Close()
				}
				
				return m, tea.Batch(
					m.loadObjects(),
					m.startTimeoutCheck,
				)
			}
		case key.Matches(msg, DefaultKeyMap().Escape):
			if m.viewingObjects {
				// Go back to bucket list
				m.viewingObjects = false
				m.title = "S3 Buckets"
				m.selectedObject = 0
			}
		case key.Matches(msg, DefaultKeyMap().Refresh):
			m.loading = true
			m.loadingStartTime = time.Now()
			m.err = nil
			
			// Create a debug file for refresh action
			f, _ := os.Create("s3_refresh.log")
			if f != nil {
				f.WriteString(fmt.Sprintf("Refreshing at %s\n", time.Now().String()))
				f.Close()
			}
			
			if m.viewingObjects {
				return m, tea.Batch(
					m.loadObjects(),
					m.startTimeoutCheck,
				)
			} else {
				return m, tea.Batch(
					m.asyncLoadBuckets,
					m.startTimeoutCheck,
				)
			}
		}
	}

	return m, nil
}

// View renders the model
func (m *S3Model) View() string {
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
			content = fmt.Sprintf("Loading S3 data... (%s)", elapsed)
		} else {
			content = "Loading S3 data..."
		}
	} else if m.err != nil {
		content = fmt.Sprintf("Error: %s\n\nPress 'r' to retry or 'd' to go to dashboard", m.err.Error())
	} else if m.viewingObjects {
		if len(m.objects) == 0 {
			content = "No objects found in this bucket"
		} else {
			// Create a table header
			header := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render("KEY\tSIZE\tLAST MODIFIED")

			// Create table rows
			var rows []string
			for i, object := range m.objects {
				// Format size
				size := fmt.Sprintf("%d B", object.Size)
				if object.Size > 1024*1024*1024 {
					size = fmt.Sprintf("%.2f GB", float64(object.Size)/(1024*1024*1024))
				} else if object.Size > 1024*1024 {
					size = fmt.Sprintf("%.2f MB", float64(object.Size)/(1024*1024))
				} else if object.Size > 1024 {
					size = fmt.Sprintf("%.2f KB", float64(object.Size)/1024)
				}

				style := lipgloss.NewStyle()
				if i == m.selectedObject {
					style = style.
						Bold(true).
						Foreground(lipgloss.Color("#FFFFFF")).
						Background(lipgloss.Color("#0066cc"))
				}

				// Format row
				row := style.Render(fmt.Sprintf(
					"%s\t%s\t%s",
					object.Key,
					size,
					object.LastModified.Format("2006-01-02 15:04:05"),
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
	} else {
		if len(m.buckets) == 0 {
			content = "No S3 buckets found"
		} else {
			// Create a table header
			header := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Render("NAME\tREGION\tCREATION DATE")

			// Create table rows
			var rows []string
			for i, bucket := range m.buckets {
				style := lipgloss.NewStyle()
				if i == m.selectedBucket {
					style = style.
						Bold(true).
						Foreground(lipgloss.Color("#FFFFFF")).
						Background(lipgloss.Color("#0066cc"))
				}

				// Format row
				row := style.Render(fmt.Sprintf(
					"%s\t%s\t%s",
					bucket.Name,
					bucket.Region,
					bucket.CreationDate.Format("2006-01-02"),
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

	// Add help text with consistent styling across all views
	var helpText string
	if m.viewingObjects {
		helpText = "\nPress ↑/↓ to navigate, Esc to go back, r to refresh, ? for help"
	} else {
		helpText = "\nPress ↑/↓ to navigate, Enter to view objects, r to refresh, ? for help"
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
func (m *S3Model) ShortHelp() []key.Binding {
	if m.viewingObjects {
		return []key.Binding{
			DefaultKeyMap().Help,
			DefaultKeyMap().Quit,
			DefaultKeyMap().Up,
			DefaultKeyMap().Down,
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
func (m *S3Model) FullHelp() [][]key.Binding {
	if m.viewingObjects {
		return [][]key.Binding{
			{
				DefaultKeyMap().Help,
				DefaultKeyMap().Quit,
				DefaultKeyMap().Command,
			},
			{
				DefaultKeyMap().Up,
				DefaultKeyMap().Down,
				DefaultKeyMap().Escape,
			},
			{
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
