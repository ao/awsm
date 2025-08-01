package events

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ao/awsm/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// EventRecorder records events from the TUI application
type EventRecorder struct {
	events     []Event
	isActive   bool
	startTime  time.Time
	mu         sync.Mutex
	outputFile string
	component  string
}

// NewEventRecorder creates a new event recorder
func NewEventRecorder() *EventRecorder {
	return &EventRecorder{
		events:    make([]Event, 0),
		isActive:  false,
		component: "event_recorder",
	}
}

// Start begins recording events
func (r *EventRecorder) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.isActive = true
	r.startTime = time.Now()
	r.events = make([]Event, 0)
	logger.InfoWithComponent(r.component, "Event recording started")
}

// Stop stops recording events
func (r *EventRecorder) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.isActive = false
	duration := time.Since(r.startTime)
	logger.InfoWithComponent(r.component, "Event recording stopped after %v, captured %d events",
		duration, len(r.events))
}

// IsActive returns whether the recorder is currently active
func (r *EventRecorder) IsActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.isActive
}

// RecordEvent records a generic event
func (r *EventRecorder) RecordEvent(event Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isActive {
		return
	}

	r.events = append(r.events, event)
	logger.DebugWithComponent(r.component, "Recorded %s event from %s", event.Type, event.Source)
}

// RecordInput records an input event
func (r *EventRecorder) RecordInput(source string, msg tea.Msg) {
	if !r.IsActive() {
		return
	}

	event := NewInputEvent(source, msg)
	r.RecordEvent(*event.Event)
}

// RecordOutput records an output event
func (r *EventRecorder) RecordOutput(source string, view string) {
	if !r.IsActive() {
		return
	}

	event := NewOutputEvent(source, view)
	r.RecordEvent(*event.Event)
}

// RecordStateChange records a state change event
func (r *EventRecorder) RecordStateChange(source string, before, after interface{}) {
	if !r.IsActive() {
		return
	}

	event := NewStateChangeEvent(source, before, after)
	r.RecordEvent(*event.Event)
}

// RecordCommand records a command event
func (r *EventRecorder) RecordCommand(source string, command string, args []string) *CommandEvent {
	if !r.IsActive() {
		return nil
	}

	event := NewCommandEvent(source, command, args)
	r.RecordEvent(*event.Event)
	return event
}

// GetEvents returns a copy of all recorded events
func (r *EventRecorder) GetEvents() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy to avoid race conditions
	eventsCopy := make([]Event, len(r.events))
	copy(eventsCopy, r.events)
	return eventsCopy
}

// Clear clears all recorded events
func (r *EventRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events = make([]Event, 0)
	logger.InfoWithComponent(r.component, "Event recorder cleared")
}

// SaveToFile saves all recorded events to a file
func (r *EventRecorder) SaveToFile(path string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create metadata
	metadata := map[string]interface{}{
		"version":     "1.0",
		"timestamp":   time.Now().Format(time.RFC3339),
		"start_time":  r.startTime.Format(time.RFC3339),
		"event_count": len(r.events),
		"duration_ms": time.Since(r.startTime).Milliseconds(),
	}

	// Create the recording object
	recording := map[string]interface{}{
		"metadata": metadata,
		"events":   r.events,
	}

	// Write to file as JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(recording); err != nil {
		return fmt.Errorf("failed to encode events: %w", err)
	}

	logger.InfoWithComponent(r.component, "Saved %d events to %s", len(r.events), path)
	return nil
}

// LoadEventsFromFile loads events from a file
func LoadEventsFromFile(path string) ([]Event, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return LoadEventsFromReader(file)
}

// LoadEventsFromReader loads events from an io.Reader
func LoadEventsFromReader(reader io.Reader) ([]Event, error) {
	// Decode the JSON
	var recording struct {
		Metadata map[string]interface{} `json:"metadata"`
		Events   []Event                `json:"events"`
	}

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&recording); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	logger.InfoWithComponent("event_recorder", "Loaded %d events", len(recording.Events))
	return recording.Events, nil
}

// SetOutputFile sets the output file for continuous recording
func (r *EventRecorder) SetOutputFile(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.outputFile = path
}

// FlushToOutputFile writes all current events to the output file if set
func (r *EventRecorder) FlushToOutputFile() error {
	r.mu.Lock()
	outputFile := r.outputFile
	r.mu.Unlock()

	if outputFile == "" {
		return nil
	}

	return r.SaveToFile(outputFile)
}
