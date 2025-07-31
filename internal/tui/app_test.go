// Package tui provides tests for the Terminal User Interface (TUI) components.
package tui

import (
	"testing"

	"github.com/ao/awsm/internal/tui/models"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockModel implements the tea.Model interface and models.BaseModel interface for testing.
// It uses the testify/mock package to mock the model's methods and track their calls.
// This allows testing the App's interactions with its models without using real implementations.
type mockModel struct {
	mock.Mock
	models.BaseModel
}

func (m *mockModel) Init() tea.Cmd {
	args := m.Called()
	return args.Get(0).(tea.Cmd)
}

func (m *mockModel) Update(msg tea.Msg) (models.Model, tea.Cmd) {
	args := m.Called(msg)
	return args.Get(0).(models.Model), args.Get(1).(tea.Cmd)
}

func (m *mockModel) View() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockModel) ShortHelp() []key.Binding {
	args := m.Called()
	return args.Get(0).([]key.Binding)
}

func (m *mockModel) FullHelp() [][]key.Binding {
	args := m.Called()
	return args.Get(0).([][]key.Binding)
}

// mockCmd is a simple mock tea.Cmd function that returns nil.
// It's used as a placeholder when a tea.Cmd is needed in tests but
// the actual command behavior is not important for the test.
func mockCmd() tea.Msg {
	return nil
}

// TestNewApp tests the NewApp constructor function.
// It verifies that a new App is created with all its components properly initialized.
func TestNewApp(t *testing.T) {
	// Create a new app
	app := NewApp()

	// Assert app is not nil
	assert.NotNil(t, app)

	// Assert app components are initialized
	assert.NotNil(t, app.statusBar)
	assert.NotNil(t, app.helpView)
	assert.NotNil(t, app.commandPalette)
	assert.NotNil(t, app.keyMap)
	assert.False(t, app.showHelp)
	assert.False(t, app.initialized)
}

// TestAppInit tests the Init method of the App.
// It verifies that the App initializes correctly, setting the dashboard model
// as the current model and marking itself as initialized.
func TestAppInit(t *testing.T) {
	// Create a new app
	app := NewApp()

	// Create a mock model
	mockDashboardModel := new(mockModel)
	mockEC2Model := new(mockModel)
	mockS3Model := new(mockModel)
	mockLambdaModel := new(mockModel)

	// Set up expectations
	mockDashboardModel.On("Init").Return(mockCmd)

	// Set the mock models
	app.dashboardModel = mockDashboardModel
	app.ec2Model = mockEC2Model
	app.s3Model = mockS3Model
	app.lambdaModel = mockLambdaModel

	// Call Init
	cmd := app.Init()

	// Assert cmd is not nil
	assert.NotNil(t, cmd)

	// Assert app is initialized
	assert.True(t, app.initialized)

	// Assert current model is set to dashboard
	assert.Equal(t, mockDashboardModel, app.currentModel)

	// Verify expectations
	mockDashboardModel.AssertExpectations(t)
}

// TestAppUpdate tests the Update method of the App.
// It verifies that the App correctly delegates updates to the current model
// and returns itself as the model along with any commands from the current model.
func TestAppUpdate(t *testing.T) {
	// Create a new app
	app := NewApp()

	// Create a mock model
	mockModel := new(mockModel)

	// Set up expectations
	mockModel.On("Update", mock.Anything).Return(mockModel, mockCmd)

	// Set the current model
	app.currentModel = mockModel
	app.initialized = true

	// Call Update with a key message
	model, cmd := app.Update(tea.KeyMsg{})

	// Assert model is the app
	assert.Equal(t, app, model)

	// Assert cmd is not nil
	assert.NotNil(t, cmd)

	// Verify expectations
	mockModel.AssertExpectations(t)
}

// TestAppView tests the View method of the App.
// It verifies that the App correctly includes the current model's view
// in its own view output.
func TestAppView(t *testing.T) {
	// Create a new app
	app := NewApp()

	// Create a mock model
	mockModel := new(mockModel)

	// Set up expectations
	mockModel.On("View").Return("Mock model view")

	// Set the current model
	app.currentModel = mockModel
	app.initialized = true

	// Call View
	view := app.View()

	// Assert view contains the model view
	assert.Contains(t, view, "Mock model view")

	// Verify expectations
	mockModel.AssertExpectations(t)
}

// TestSwitchToModel tests the SwitchToModel method of the App.
// It verifies that the App correctly switches to a new model,
// setting it as the current model and initializing it.
func TestSwitchToModel(t *testing.T) {
	// Create a new app
	app := NewApp()

	// Create mock models
	mockCurrentModel := new(mockModel)
	mockNewModel := new(mockModel)

	// Set up expectations
	mockNewModel.On("Init").Return(mockCmd)

	// Set the current model
	app.currentModel = mockCurrentModel

	// Call SwitchToModel
	app.SwitchToModel(mockNewModel)

	// Assert current model is set to the new model
	assert.Equal(t, mockNewModel, app.currentModel)

	// Verify expectations
	mockNewModel.AssertExpectations(t)
}

// TestRun tests the Run method of the App.
// This test is skipped because it would require actually starting the TUI,
// which is not suitable for automated testing. In a more comprehensive test suite,
// we would use dependency injection to mock the tea.Program and test the Run method
// without actually starting the UI.
func TestRun(t *testing.T) {
	// Skip this test in short mode
	if testing.Short() {
		t.Skip("Skipping TUI test in short mode")
	}

	// This test is difficult to run without actually starting the TUI
	// In a real test, we would use dependency injection to mock the dependencies
	// For now, we'll just skip this test
	t.Skip("Skipping TUI Run test - requires dependency injection")
}
