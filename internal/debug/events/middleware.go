package events

import (
	"github.com/ao/awsm/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// RecorderMiddleware creates middleware that records events
func RecorderMiddleware(recorder *EventRecorder, componentName string) func(tea.Model, ...tea.Cmd) tea.Model {
	return func(next tea.Model, cmds ...tea.Cmd) tea.Model {
		return recorderModel{
			next:      next,
			recorder:  recorder,
			component: componentName,
			cmds:      cmds,
		}
	}
}

// recorderModel is a model that wraps another model and records events
type recorderModel struct {
	next      tea.Model
	recorder  *EventRecorder
	component string
	cmds      []tea.Cmd
	lastView  string
}

// Init initializes the recorder model
func (m recorderModel) Init() tea.Cmd {
	logger.DebugWithComponent(m.component, "Initializing recorder middleware")

	// Initialize the next model
	cmd := m.next.Init()

	// Record the initialization
	if m.recorder.IsActive() {
		m.recorder.RecordCommand(m.component, "Init", nil)
	}

	return cmd
}

// Update updates the recorder model
func (m recorderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Record the input message
	if m.recorder.IsActive() {
		m.recorder.RecordInput(m.component, msg)
	}

	// Update the next model
	nextModel, cmd := m.next.Update(msg)

	// Create a new model with the updated next model
	newModel := recorderModel{
		next:      nextModel,
		recorder:  m.recorder,
		component: m.component,
		lastView:  m.lastView,
	}

	return newModel, cmd
}

// View renders the view of the next model
func (m recorderModel) View() string {
	// Get the view from the next model
	view := m.next.View()

	// Record the output if it has changed
	if m.recorder.IsActive() && view != m.lastView {
		m.recorder.RecordOutput(m.component, view)
		m.lastView = view
	}

	return view
}

// StateChangeMiddleware creates middleware that records state changes
func StateChangeMiddleware(recorder *EventRecorder, componentName string) func(tea.Model, ...tea.Cmd) tea.Model {
	var lastState interface{}

	return func(next tea.Model, cmds ...tea.Cmd) tea.Model {
		// Record initial state
		if lastState == nil && recorder.IsActive() {
			lastState = next
			recorder.RecordStateChange(componentName, nil, next)
		}

		return stateChangeModel{
			next:      next,
			recorder:  recorder,
			component: componentName,
			cmds:      cmds,
			lastState: lastState,
		}
	}
}

// stateChangeModel is a model that wraps another model and records state changes
type stateChangeModel struct {
	next      tea.Model
	recorder  *EventRecorder
	component string
	cmds      []tea.Cmd
	lastState interface{}
}

// Init initializes the state change model
func (m stateChangeModel) Init() tea.Cmd {
	return m.next.Init()
}

// Update updates the state change model
func (m stateChangeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Update the next model
	nextModel, cmd := m.next.Update(msg)

	// Record state change if the model has changed
	if m.recorder.IsActive() && m.lastState != nextModel {
		m.recorder.RecordStateChange(m.component, m.lastState, nextModel)
		m.lastState = nextModel
	}

	// Create a new model with the updated next model and state
	newModel := stateChangeModel{
		next:      nextModel,
		recorder:  m.recorder,
		component: m.component,
		lastState: m.lastState,
	}

	return newModel, cmd
}

// View renders the view of the next model
func (m stateChangeModel) View() string {
	return m.next.View()
}

// CombineMiddleware combines multiple middleware functions into one
func CombineMiddleware(middlewares ...func(tea.Model, ...tea.Cmd) tea.Model) func(tea.Model, ...tea.Cmd) tea.Model {
	return func(model tea.Model, cmds ...tea.Cmd) tea.Model {
		for i := len(middlewares) - 1; i >= 0; i-- {
			model = middlewares[i](model, cmds...)
		}
		return model
	}
}

// NewRecordingMiddleware creates middleware that records all types of events
func NewRecordingMiddleware(recorder *EventRecorder, componentName string) func(tea.Model, ...tea.Cmd) tea.Model {
	return CombineMiddleware(
		RecorderMiddleware(recorder, componentName),
		StateChangeMiddleware(recorder, componentName),
	)
}
