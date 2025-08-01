package events

import (
	"encoding/json"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	// Create a new event
	event := NewEvent(EventTypeInput, "test")

	// Check that the event has the correct type and source
	assert.Equal(t, EventTypeInput, event.Type)
	assert.Equal(t, "test", event.Source)

	// Check that the event has a timestamp
	assert.False(t, event.Timestamp.IsZero())

	// Check that the event has an ID
	assert.NotEmpty(t, event.ID)

	// Add data to the event
	event.WithData("key", "value")

	// Check that the data was added
	assert.Equal(t, "value", event.Data["key"])

	// Serialize the event to JSON
	jsonData, err := event.ToJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize the event from JSON
	deserializedEvent, err := FromJSON(jsonData)
	assert.NoError(t, err)

	// Check that the deserialized event matches the original
	assert.Equal(t, event.ID, deserializedEvent.ID)
	assert.Equal(t, event.Type, deserializedEvent.Type)
	assert.Equal(t, event.Source, deserializedEvent.Source)
	assert.Equal(t, "value", deserializedEvent.Data["key"])
}

func TestInputEvent(t *testing.T) {
	// Create a key message
	keyMsg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("a"),
	}

	// Create a new input event
	event := NewInputEvent("test", keyMsg)

	// Check that the event has the correct type and source
	assert.Equal(t, EventTypeInput, event.Event.Type)
	assert.Equal(t, "test", event.Event.Source)

	// Check that the event has the correct data
	assert.Equal(t, "key", event.Data["type"])
	assert.Equal(t, "a", event.Data["key"])
	assert.Equal(t, "a", event.Data["runes"])

	// Create a mouse message
	mouseMsg := tea.MouseMsg{
		X:      10,
		Y:      20,
		Action: tea.MouseActionPress,
	}

	// Create a new input event
	event = NewInputEvent("test", mouseMsg)

	// Check that the event has the correct data
	assert.Equal(t, "mouse", event.Data["type"])
	assert.Equal(t, float64(10), event.Data["x"])
	assert.Equal(t, float64(20), event.Data["y"])
}

func TestOutputEvent(t *testing.T) {
	// Create a new output event
	event := NewOutputEvent("test", "Hello, world!")

	// Check that the event has the correct type and source
	assert.Equal(t, EventTypeOutput, event.Event.Type)
	assert.Equal(t, "test", event.Event.Source)

	// Check that the event has the correct data
	assert.Equal(t, "Hello, world!", event.View)
	assert.Equal(t, "Hello, world!", event.Data["view_summary"])
	assert.Equal(t, float64(13), event.Data["view_length"])

	// Create a new output event with a long view
	longView := "This is a very long view that should be truncated in the summary."
	event = NewOutputEvent("test", longView)

	// Check that the summary is truncated
	expectedSummary := "This is a very long view that should be truncated in the summary." + "..."
	assert.Equal(t, expectedSummary, event.Data["view_summary"])
	assert.Equal(t, float64(len(longView)), event.Data["view_length"])
}

func TestStateChangeEvent(t *testing.T) {
	// Create before and after states
	before := map[string]interface{}{"count": 1}
	after := map[string]interface{}{"count": 2}

	// Create a new state change event
	event := NewStateChangeEvent("test", before, after)

	// Check that the event has the correct type and source
	assert.Equal(t, EventTypeStateChange, event.Event.Type)
	assert.Equal(t, "test", event.Event.Source)

	// Check that the event has the correct data
	assert.Equal(t, before, event.Before)
	assert.Equal(t, after, event.After)
	assert.Equal(t, "map[string]interface {}", event.Data["before_type"])
	assert.Equal(t, "map[string]interface {}", event.Data["after_type"])

	// Check that the JSON data is correct
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	assert.Equal(t, string(beforeJSON), event.Data["before_json"])
	assert.Equal(t, string(afterJSON), event.Data["after_json"])
}

func TestCommandEvent(t *testing.T) {
	// Create a new command event
	event := NewCommandEvent("test", "echo", []string{"hello"})

	// Check that the event has the correct type and source
	assert.Equal(t, EventTypeCommand, event.Event.Type)
	assert.Equal(t, "test", event.Event.Source)

	// Check that the event has the correct data
	assert.Equal(t, "echo", event.Command)
	assert.Equal(t, []string{"hello"}, event.Args)
	assert.Equal(t, "echo", event.Data["command"])
	assert.Equal(t, []interface{}{"hello"}, event.Data["args"])

	// Add a result to the event
	result := map[string]interface{}{"output": "hello"}
	event.WithResult(result)

	// Check that the result was added
	assert.Equal(t, result, event.Result)
	resultJSON, _ := json.Marshal(result)
	assert.Equal(t, string(resultJSON), event.Data["result"])
}
