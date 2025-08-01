package events

import (
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// EventType represents the type of event being recorded
type EventType string

// Event types
const (
	EventTypeInput       EventType = "input"
	EventTypeOutput      EventType = "output"
	EventTypeStateChange EventType = "state_change"
	EventTypeCommand     EventType = "command"
)

// Event represents a recorded event in the TUI application
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// NewEvent creates a new event with the given type and source
func NewEvent(eventType EventType, source string) *Event {
	return &Event{
		ID:        fmt.Sprintf("%s-%d", eventType, time.Now().UnixNano()),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      make(map[string]interface{}),
	}
}

// WithData adds data to the event
func (e *Event) WithData(key string, value interface{}) *Event {
	e.Data[key] = value
	return e
}

// ToJSON serializes the event to JSON
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes an event from JSON
func FromJSON(data []byte) (*Event, error) {
	var event Event
	err := json.Unmarshal(data, &event)
	return &event, err
}

// InputEvent represents a user input event
type InputEvent struct {
	*Event
	KeyMsg *tea.KeyMsg `json:"-"` // Not serialized directly
}

// NewInputEvent creates a new input event
func NewInputEvent(source string, msg tea.Msg) *InputEvent {
	event := &InputEvent{
		Event: NewEvent(EventTypeInput, source),
	}

	// Handle different types of input messages
	switch m := msg.(type) {
	case tea.KeyMsg:
		event.KeyMsg = &m
		event.WithData("key", m.String())
		event.WithData("type", "key")
		event.WithData("runes", string(m.Runes))
	case tea.MouseMsg:
		event.WithData("type", "mouse")
		event.WithData("x", m.X)
		event.WithData("y", m.Y)
		event.WithData("action", fmt.Sprintf("%d", m.Action))
	case tea.WindowSizeMsg:
		event.WithData("type", "window_size")
		event.WithData("width", m.Width)
		event.WithData("height", m.Height)
	default:
		event.WithData("type", fmt.Sprintf("%T", msg))
	}

	return event
}

// OutputEvent represents an application output event
type OutputEvent struct {
	*Event
	View string `json:"-"` // Not serialized directly
}

// NewOutputEvent creates a new output event
func NewOutputEvent(source string, view string) *OutputEvent {
	event := &OutputEvent{
		Event: NewEvent(EventTypeOutput, source),
		View:  view,
	}

	// Store a summary of the view (first 100 chars)
	summary := view
	if len(view) > 100 {
		summary = view[:100] + "..."
	}
	event.WithData("view_summary", summary)
	event.WithData("view_length", len(view))

	return event
}

// StateChangeEvent represents a state change event
type StateChangeEvent struct {
	*Event
	Before interface{} `json:"-"` // Not serialized directly
	After  interface{} `json:"-"` // Not serialized directly
}

// NewStateChangeEvent creates a new state change event
func NewStateChangeEvent(source string, before, after interface{}) *StateChangeEvent {
	event := &StateChangeEvent{
		Event:  NewEvent(EventTypeStateChange, source),
		Before: before,
		After:  after,
	}

	// Serialize before and after to JSON for storage
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)

	// Store summaries
	event.WithData("before_type", fmt.Sprintf("%T", before))
	event.WithData("after_type", fmt.Sprintf("%T", after))
	event.WithData("before_json", string(beforeJSON))
	event.WithData("after_json", string(afterJSON))

	return event
}

// CommandEvent represents a command execution event
type CommandEvent struct {
	*Event
	Command string      `json:"-"` // Not serialized directly
	Args    []string    `json:"-"` // Not serialized directly
	Result  interface{} `json:"-"` // Not serialized directly
}

// NewCommandEvent creates a new command event
func NewCommandEvent(source string, command string, args []string) *CommandEvent {
	event := &CommandEvent{
		Event:   NewEvent(EventTypeCommand, source),
		Command: command,
		Args:    args,
	}

	event.WithData("command", command)
	event.WithData("args", args)

	return event
}

// WithResult adds the result of the command to the event
func (e *CommandEvent) WithResult(result interface{}) *CommandEvent {
	e.Result = result
	resultJSON, _ := json.Marshal(result)
	e.WithData("result", string(resultJSON))
	return e
}
