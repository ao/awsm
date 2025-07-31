package tui

import (
	"fmt"
	"os"

	"github.com/ao/awsm/internal/config"
	"github.com/ao/awsm/internal/logger"
	"github.com/ao/awsm/internal/tui/components"
	"github.com/ao/awsm/internal/tui/models"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Version information - imported from main package
var (
	Version    = "0.1.0"
	BuildTime  = "unknown"
	CommitHash = "unknown"
)

// SetVersionInfo sets the version information
func SetVersionInfo(version, buildTime, commitHash string) {
	if version != "" {
		Version = version
	}
	if buildTime != "" {
		BuildTime = buildTime
	}
	if commitHash != "" {
		CommitHash = commitHash
	}

	// Log version information
	logger.Info("TUI Version set to: %s (built: %s, commit: %s)",
		Version, BuildTime, CommitHash)

	// Create a debug file to verify this function is being called
	debugFile, err := os.Create("tui_debug.log")
	if err == nil {
		debugFile.WriteString(fmt.Sprintf("Version: %s\nBuildTime: %s\nCommitHash: %s\n",
			Version, BuildTime, CommitHash))
		debugFile.Close()
	}
}

// App represents the TUI application
type App struct {
	// Current model being displayed
	currentModel models.Model

	// Available models
	dashboardModel models.Model
	ec2Model       models.Model
	s3Model        models.Model
	lambdaModel    models.Model

	// UI components
	statusBar       *components.StatusBar
	helpView        *components.HelpView
	commandPalette  *components.CommandPalette
	contextSwitcher *components.ContextSwitcher
	profileSelector *components.ProfileSelector
	regionSelector  *components.RegionSelector
	logo            *components.Logo
	resultsPanel    *components.ResultsPanel

	// State
	width       int
	height      int
	showHelp    bool
	keyMap      models.KeyMap
	initialized bool
}

// NewApp creates a new TUI application
func NewApp() *App {
	return &App{
		statusBar:      components.NewStatusBar(),
		helpView:       components.NewHelpView(),
		commandPalette: components.NewCommandPalette(),
		logo:           components.NewLogo(),
		resultsPanel:   components.NewResultsPanel(),
		keyMap:         models.DefaultKeyMap(),
		showHelp:       false,
		initialized:    false,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	// Initialize components
	a.statusBar = components.NewStatusBar()
	a.helpView = components.NewHelpView()
	a.commandPalette = components.NewCommandPalette()
	a.logo = components.NewLogo()
	a.resultsPanel = components.NewResultsPanel()

	// Initialize context switcher with a callback to switch contexts
	a.contextSwitcher = components.NewContextSwitcher(func(contextName string) {
		// Switch to the selected context
		if err := config.SetCurrentContext(contextName); err == nil {
			// Refresh the current model to reflect the new context
			a.currentModel.Init()
		}
	})

	// Initialize profile selector with a callback to switch profiles
	a.profileSelector = components.NewProfileSelector(func(profileName string) {
		// Switch to the selected profile
		if err := config.SetAWSProfile(profileName); err == nil {
			// Refresh the current model to reflect the new profile
			a.currentModel.Init()
		}
	})

	// Initialize region selector with a callback to switch regions
	a.regionSelector = components.NewRegionSelector(func(regionName string) {
		// Switch to the selected region
		if err := config.SetAWSRegion(regionName); err == nil {
			// Refresh the current model to reflect the new region
			a.currentModel.Init()
		}
	})

	// Add common commands to the command palette
	a.commandPalette.AddCommand("quit", "Quit the application", func() error {
		return fmt.Errorf("quit")
	})
	a.commandPalette.AddCommand("help", "Toggle help", func() error {
		a.showHelp = !a.showHelp
		return nil
	})
	a.commandPalette.AddCommand("dashboard", "Go to dashboard", func() error {
		a.SwitchToModel(a.dashboardModel)
		return nil
	})
	a.commandPalette.AddCommand("ec2", "Go to EC2 view", func() error {
		a.SwitchToModel(a.ec2Model)
		return nil
	})
	a.commandPalette.AddCommand("s3", "Go to S3 view", func() error {
		a.SwitchToModel(a.s3Model)
		return nil
	})
	a.commandPalette.AddCommand("lambda", "Go to Lambda view", func() error {
		a.SwitchToModel(a.lambdaModel)
		return nil
	})

	// Initialize models
	a.dashboardModel = models.NewDashboardModel()
	a.ec2Model = models.NewEC2Model()
	a.s3Model = models.NewS3Model()
	a.lambdaModel = models.NewLambdaModel()

	// Set the current model to the dashboard
	a.currentModel = a.dashboardModel

	// Mark as initialized
	a.initialized = true

	// Return the current model's init command
	return a.currentModel.Init()
}

// Update updates the application based on messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global key bindings
		switch {
		case a.contextSwitcher.IsVisible():
			// If context switcher is visible, pass the message to it
			handled, cmd := a.contextSwitcher.HandleKeyMsg(msg)
			if handled {
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return a, tea.Batch(cmds...)
			}
		case a.profileSelector.IsVisible():
			// If profile selector is visible, pass the message to it
			handled, cmd := a.profileSelector.HandleKeyMsg(msg)
			if handled {
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return a, tea.Batch(cmds...)
			}
		case a.regionSelector.IsVisible():
			// If region selector is visible, pass the message to it
			handled, cmd := a.regionSelector.HandleKeyMsg(msg)
			if handled {
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return a, tea.Batch(cmds...)
			}
		case a.commandPalette.IsActive():
			// If command palette is active, pass the message to it
			a.commandPalette.HandleInput(msg)
			if msg.String() == "enter" {
				// Execute the selected command
				cmd := a.commandPalette.ExecuteSelected()
				if cmd != nil {
					cmds = append(cmds, func() tea.Msg { return cmd })
				}
				// Close the command palette
				a.commandPalette.SetActive(false)
			} else if msg.String() == "esc" {
				// Close the command palette
				a.commandPalette.SetActive(false)
			}
		case key.Matches(msg, a.keyMap.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keyMap.Help):
			a.showHelp = !a.showHelp
		case key.Matches(msg, a.keyMap.Command):
			a.commandPalette.SetActive(true)
		case key.Matches(msg, a.keyMap.Context):
			// Show context switcher
			a.contextSwitcher.Show()
		case key.Matches(msg, a.keyMap.Profile):
			// Show profile selector
			a.profileSelector.Show()
		case key.Matches(msg, a.keyMap.Region):
			// Show region selector
			a.regionSelector.Show()
		case key.Matches(msg, a.keyMap.Dashboard):
			a.SwitchToModel(a.dashboardModel)
		case key.Matches(msg, a.keyMap.EC2):
			a.SwitchToModel(a.ec2Model)
		case key.Matches(msg, a.keyMap.S3):
			a.SwitchToModel(a.s3Model)
		case key.Matches(msg, a.keyMap.Lambda):
			a.SwitchToModel(a.lambdaModel)
		case key.Matches(msg, a.keyMap.Refresh):
			cmds = append(cmds, a.currentModel.Init())
		default:
			// Pass the message to the current model
			newModel, cmd := a.currentModel.Update(msg)
			if m, ok := newModel.(models.Model); ok {
				a.currentModel = m
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case tea.WindowSizeMsg:
		// Update the size of the application
		a.width = msg.Width
		a.height = msg.Height

		// Update the size of components
		a.statusBar.SetWidth(a.width)
		a.helpView.SetSize(a.width, a.height/3)
		a.commandPalette.SetSize(a.width, a.height/3)
		a.contextSwitcher.SetSize(a.width/2, a.height/2)
		a.profileSelector.SetSize(a.width/2, a.height/2)
		a.regionSelector.SetSize(a.width/2, a.height/2)

		// Set logo size based on terminal width
		logoWidth := a.width / 5
		if logoWidth < 20 {
			logoWidth = 20 // Minimum width
		} else if logoWidth > 30 {
			logoWidth = 30 // Maximum width
		}
		a.logo.SetSize(logoWidth, 4) // 4 lines height

		// Set results panel size
		resultsHeight := a.height - 6 // Leave space for status bar and header
		if resultsHeight < 10 {
			resultsHeight = 10 // Minimum height
		}
		a.resultsPanel.SetSize(a.width, resultsHeight)

		// Update the size of the models
		if m, ok := a.dashboardModel.(*models.DashboardModel); ok {
			m.BaseModel.SetSize(a.width, resultsHeight)
		}
		if m, ok := a.ec2Model.(*models.EC2Model); ok {
			m.BaseModel.SetSize(a.width, resultsHeight)
		}
		if m, ok := a.s3Model.(*models.S3Model); ok {
			m.BaseModel.SetSize(a.width, resultsHeight)
		}
		if m, ok := a.lambdaModel.(*models.LambdaModel); ok {
			m.BaseModel.SetSize(a.width, resultsHeight)
		}
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	if !a.initialized {
		return "Initializing..."
	}

	// Render the logo
	logoView := a.logo.Render()

	// Get the model content and set it to the results panel
	a.resultsPanel.SetTitle(a.getCurrentModelTitle())

	// Check if the current model is loading or has an error
	// All models implement IsLoading() and GetError() methods
	if a.currentModel.IsLoading() {
		a.resultsPanel.SetLoading(true)
	} else if a.currentModel.GetError() != nil {
		a.resultsPanel.SetError(a.currentModel.GetError())
	} else {
		// Only set content if not loading and no error
		a.resultsPanel.SetContent(a.currentModel.View())
	}

	// Render the results panel
	resultsView := a.resultsPanel.Render()

	// Render the status bar with current config
	a.statusBar.SetWidth(a.width)
	statusBarView := a.statusBar.Render()

	// Render the help view if enabled
	var helpView string
	if a.showHelp {
		helpView = a.helpView.RenderBindings(a.currentModel.ShortHelp()...)
	}

	// Render the command palette if active
	var commandPaletteView string
	if a.commandPalette.IsActive() {
		commandPaletteView = a.commandPalette.Render()
	}

	// Render the context switcher if visible
	var contextSwitcherView string
	if a.contextSwitcher.IsVisible() {
		contextSwitcherView = a.contextSwitcher.View()
	}

	// Render the profile selector if visible
	var profileSelectorView string
	if a.profileSelector.IsVisible() {
		profileSelectorView = a.profileSelector.View()
	}

	// Render the region selector if visible
	var regionSelectorView string
	if a.regionSelector.IsVisible() {
		regionSelectorView = a.regionSelector.View()
	}

	// Create a header with the logo positioned at the right and AWSM info at the left
	headerStyle := lipgloss.NewStyle().Width(a.width)

	// Create AWSM info section for the top left (similar to k9s)
	awsmInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF9900")). // AWS Orange color
		Bold(true).
		Padding(1, 2)

	awsmInfo := awsmInfoStyle.Render(fmt.Sprintf("AWSM %s", Version))

	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Center,
		lipgloss.NewStyle().
			Width(a.width-lipgloss.Width(logoView)).
			Align(lipgloss.Left).
			Render(awsmInfo),
		logoView,
	)
	headerRow := headerStyle.Render(headerContent)

	// Combine all views
	var view string
	if a.contextSwitcher.IsVisible() {
		// Show context switcher in the middle
		view = lipgloss.JoinVertical(
			lipgloss.Left,
			headerRow,
			resultsView,
			contextSwitcherView,
			statusBarView,
		)
	} else if a.profileSelector.IsVisible() {
		// Show profile selector in the middle
		view = lipgloss.JoinVertical(
			lipgloss.Left,
			headerRow,
			resultsView,
			profileSelectorView,
			statusBarView,
		)
	} else if a.regionSelector.IsVisible() {
		// Show region selector in the middle
		view = lipgloss.JoinVertical(
			lipgloss.Left,
			headerRow,
			resultsView,
			regionSelectorView,
			statusBarView,
		)
	} else if a.commandPalette.IsActive() {
		// Show command palette in the middle
		view = lipgloss.JoinVertical(
			lipgloss.Left,
			headerRow,
			resultsView,
			commandPaletteView,
			statusBarView,
		)
	} else {
		// Show results view with logo and status bar
		view = lipgloss.JoinVertical(
			lipgloss.Left,
			headerRow,
			resultsView,
			statusBarView,
		)
	}

	// If help is visible, overlay it on top of the view instead of replacing it
	if a.showHelp {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			view,
			helpView,
		)
	}

	return view
}

// getCurrentModelTitle returns the title of the current model
func (a *App) getCurrentModelTitle() string {
	switch a.currentModel {
	case a.dashboardModel:
		return "Dashboard"
	case a.ec2Model:
		return "EC2 Instances"
	case a.s3Model:
		return "S3 Buckets"
	case a.lambdaModel:
		return "Lambda Functions"
	default:
		return "Results"
	}
}

// SwitchToModel switches to the specified model
func (a *App) SwitchToModel(model models.Model) {
	a.currentModel = model
	a.currentModel.Init()
}

// Run runs the TUI application
func Run() error {
	// Initialize configuration
	if err := config.Initialize(); err != nil {
		logger.Error("Failed to initialize configuration: %v", err)
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	// Log version information again to confirm it's still correct
	logger.Info("TUI starting with version: %s (built: %s, commit: %s)",
		Version, BuildTime, CommitHash)

	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
