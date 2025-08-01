package events

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestEventRecorder(t *testing.T) {
	// Create a new recorder
	recorder := NewEventRecorder()

	// Check that the recorder is not active by default
	assert.False(t, recorder.IsActive())

	// Start the recorder
	recorder.Start()

	// Check that the recorder is now active
	assert.True(t, recorder.IsActive())

	// Record some events
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	recorder.RecordInput("test", keyMsg)
	recorder.RecordOutput("test", "Hello, world!")
	recorder.RecordStateChange("test", map[string]interface{}{"count": 1}, map[string]interface{}{"count": 2})
	recorder.RecordCommand("test", "echo", []string{"hello"})

	// Check that the events were recorded
	events := recorder.GetEvents()
	assert.Equal(t, 4, len(events))

	// Check the types of the events
	assert.Equal(t, EventTypeInput, events[0].Type)
	assert.Equal(t, EventTypeOutput, events[1].Type)
	assert.Equal(t, EventTypeStateChange, events[2].Type)
	assert.Equal(t, EventTypeCommand, events[3].Type)

	// Stop the recorder
	recorder.Stop()

	// Check that the recorder is no longer active
	assert.False(t, recorder.IsActive())

	// Try to record an event after stopping
	recorder.RecordInput("test", keyMsg)

	// Check that no new event was recorded
	events = recorder.GetEvents()
	assert.Equal(t, 4, len(events))

	// Clear the recorder
	recorder.Clear()

	// Check that all events were cleared
	events = recorder.GetEvents()
	assert.Equal(t, 0, len(events))
}

func TestEventRecorderSaveLoad(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "events_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new recorder
	recorder := NewEventRecorder()
	recorder.Start()

	// Record some events
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	recorder.RecordInput("test", keyMsg)
	recorder.RecordOutput("test", "Hello, world!")

	// Save the events to a file
	filePath := filepath.Join(tempDir, "events.json")
	err = recorder.SaveToFile(filePath)
	assert.NoError(t, err)

	// Check that the file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Load the events from the file
	loadedEvents, err := LoadEventsFromFile(filePath)
	assert.NoError(t, err)

	// Check that the loaded events match the original events
	originalEvents := recorder.GetEvents()
	assert.Equal(t, len(originalEvents), len(loadedEvents))

	// Check the types of the loaded events
	assert.Equal(t, EventTypeInput, loadedEvents[0].Type)
	assert.Equal(t, EventTypeOutput, loadedEvents[1].Type)

	// Set an output file for continuous recording
	outputFilePath := filepath.Join(tempDir, "continuous.json")
	recorder.SetOutputFile(outputFilePath)

	// Record another event
	recorder.RecordStateChange("test", map[string]interface{}{"count": 1}, map[string]interface{}{"count": 2})

	// Flush to the output file
	err = recorder.FlushToOutputFile()
	assert.NoError(t, err)

	// Check that the file exists
	_, err = os.Stat(outputFilePath)
	assert.NoError(t, err)

	// Load the events from the output file
	continuousEvents, err := LoadEventsFromFile(outputFilePath)
	assert.NoError(t, err)

	// Check that all events were saved
	assert.Equal(t, 3, len(continuousEvents))
}

func TestEventFilters(t *testing.T) {
	// Create some events with different types and sources
	events := []Event{
		*NewEvent(EventTypeInput, "source1"),
		*NewEvent(EventTypeOutput, "source1"),
		*NewEvent(EventTypeStateChange, "source2"),
		*NewEvent(EventTypeCommand, "source2"),
	}

	// Set timestamps
	now := time.Now()
	events[0].Timestamp = now.Add(-3 * time.Hour)
	events[1].Timestamp = now.Add(-2 * time.Hour)
	events[2].Timestamp = now.Add(-1 * time.Hour)
	events[3].Timestamp = now

	// Test filtering by type
	typeFilter := FilterEventsByType(EventTypeInput, EventTypeOutput)
	filteredEvents := make([]Event, 0)
	for _, e := range events {
		if typeFilter(e) {
			filteredEvents = append(filteredEvents, e)
		}
	}
	assert.Equal(t, 2, len(filteredEvents))
	assert.Equal(t, EventTypeInput, filteredEvents[0].Type)
	assert.Equal(t, EventTypeOutput, filteredEvents[1].Type)

	// Test filtering by source
	sourceFilter := FilterEventsBySource("source2")
	filteredEvents = make([]Event, 0)
	for _, e := range events {
		if sourceFilter(e) {
			filteredEvents = append(filteredEvents, e)
		}
	}
	assert.Equal(t, 2, len(filteredEvents))
	assert.Equal(t, EventTypeStateChange, filteredEvents[0].Type)
	assert.Equal(t, EventTypeCommand, filteredEvents[1].Type)

	// Test filtering by time
	timeFilter := FilterEventsAfterTime(now.Add(-90 * time.Minute))
	filteredEvents = make([]Event, 0)
	for _, e := range events {
		if timeFilter(e) {
			filteredEvents = append(filteredEvents, e)
		}
	}
	assert.Equal(t, 2, len(filteredEvents))
	assert.Equal(t, EventTypeStateChange, filteredEvents[0].Type)
	assert.Equal(t, EventTypeCommand, filteredEvents[1].Type)

	// Test combining filters
	combinedFilter := CombineFilters(
		FilterEventsBySource("source2"),
		FilterEventsAfterTime(now.Add(-90*time.Minute)),
	)
	filteredEvents = make([]Event, 0)
	for _, e := range events {
		if combinedFilter(e) {
			filteredEvents = append(filteredEvents, e)
		}
	}
	assert.Equal(t, 2, len(filteredEvents))
	assert.Equal(t, EventTypeStateChange, filteredEvents[0].Type)
	assert.Equal(t, EventTypeCommand, filteredEvents[1].Type)
}
