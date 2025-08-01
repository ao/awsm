package events

import (
	"time"

	"github.com/ao/awsm/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// ReplayOptions configures how events are replayed
type ReplayOptions struct {
	// Speed multiplier for replay (1.0 = original speed, 2.0 = twice as fast)
	SpeedMultiplier float64

	// Filter function to determine which events to replay
	Filter func(Event) bool

	// Whether to simulate timing between events
	SimulateTiming bool

	// Maximum duration to wait between events
	MaxWaitDuration time.Duration
}

// DefaultReplayOptions returns the default replay options
func DefaultReplayOptions() ReplayOptions {
	return ReplayOptions{
		SpeedMultiplier: 1.0,
		Filter:          func(e Event) bool { return true }, // Include all events
		SimulateTiming:  true,
		MaxWaitDuration: 2 * time.Second, // Don't wait more than 2 seconds between events
	}
}

// EventReplayMsg is a message sent during replay
type EventReplayMsg struct {
	Event Event
}

// ReplayEvents replays a sequence of events through a Bubble Tea program
func ReplayEvents(events []Event, options ReplayOptions) tea.Cmd {
	return func() tea.Msg {
		// Return the first event immediately
		if len(events) == 0 {
			return nil
		}

		// Filter events
		filteredEvents := make([]Event, 0, len(events))
		for _, event := range events {
			if options.Filter(event) {
				filteredEvents = append(filteredEvents, event)
			}
		}

		if len(filteredEvents) == 0 {
			return nil
		}

		// Return the first event
		return EventReplayMsg{Event: filteredEvents[0]}
	}
}

// ReplayNextEvent returns a command that will replay the next event
func ReplayNextEvent(events []Event, currentIndex int, options ReplayOptions) tea.Cmd {
	return func() tea.Msg {
		// Check if we've reached the end
		if currentIndex >= len(events)-1 {
			logger.InfoWithComponent("event_replay", "Replay completed, %d events replayed", currentIndex+1)
			return nil
		}

		// Get the next event
		nextIndex := currentIndex + 1
		nextEvent := events[nextIndex]

		// Calculate delay if simulating timing
		var delay time.Duration
		if options.SimulateTiming && currentIndex >= 0 {
			currentEvent := events[currentIndex]
			delay = nextEvent.Timestamp.Sub(currentEvent.Timestamp)

			// Apply speed multiplier
			if options.SpeedMultiplier > 0 {
				delay = time.Duration(float64(delay) / options.SpeedMultiplier)
			}

			// Cap the delay
			if delay > options.MaxWaitDuration {
				delay = options.MaxWaitDuration
			}
		}

		// Wait for the delay
		if delay > 0 {
			time.Sleep(delay)
		}

		logger.DebugWithComponent("event_replay", "Replaying event %d/%d: %s from %s",
			nextIndex+1, len(events), nextEvent.Type, nextEvent.Source)

		// Return the next event
		return EventReplayMsg{Event: nextEvent}
	}
}

// NewReplayMiddleware creates middleware for replaying events
func NewReplayMiddleware(events []Event) func(tea.Model, ...tea.Cmd) tea.Model {
	options := DefaultReplayOptions()
	currentIndex := -1

	return func(next tea.Model, cmds ...tea.Cmd) tea.Model {
		return replayModel{
			next:         next,
			events:       events,
			currentIndex: &currentIndex,
			options:      options,
			cmds:         cmds,
		}
	}
}

// replayModel is a model that wraps another model and replays events to it
type replayModel struct {
	next         tea.Model
	events       []Event
	currentIndex *int
	options      ReplayOptions
	cmds         []tea.Cmd
}

// Init initializes the replay model
func (m replayModel) Init() tea.Cmd {
	cmds := []tea.Cmd{m.next.Init()}

	// Start the replay if we have events
	if len(m.events) > 0 {
		cmds = append(cmds, ReplayEvents(m.events, m.options))
	}

	return tea.Batch(cmds...)
}

// Update updates the replay model
func (m replayModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle replay messages
	switch msg := msg.(type) {
	case EventReplayMsg:
		// Convert the event to a tea.Msg
		var teaMsg tea.Msg

		switch msg.Event.Type {
		case EventTypeInput:
			// Try to reconstruct the input message
			if keyStr, ok := msg.Event.Data["key"].(string); ok {
				// This is a simplified reconstruction - in a real implementation,
				// you would need more sophisticated logic to recreate the exact message
				teaMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)}
			}
		case EventTypeOutput:
			// Output events don't generate messages
			return m.next, nil
		case EventTypeStateChange:
			// State change events don't generate messages
			return m.next, nil
		case EventTypeCommand:
			// Command events don't generate messages
			return m.next, nil
		}

		if teaMsg != nil {
			// Update the next model with the reconstructed message
			nextModel, cmd := m.next.Update(teaMsg)
			m.next = nextModel
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		// Increment the current index
		*m.currentIndex++

		// Queue up the next event
		cmds = append(cmds, ReplayNextEvent(m.events, *m.currentIndex, m.options))

		return m, tea.Batch(cmds...)
	}

	// Pass other messages to the next model
	nextModel, cmd := m.next.Update(msg)
	m.next = nextModel
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the view of the next model
func (m replayModel) View() string {
	return m.next.View()
}

// FilterEventsByType returns a filter function that only includes events of the specified types
func FilterEventsByType(types ...EventType) func(Event) bool {
	return func(e Event) bool {
		for _, t := range types {
			if e.Type == t {
				return true
			}
		}
		return false
	}
}

// FilterEventsBySource returns a filter function that only includes events from the specified sources
func FilterEventsBySource(sources ...string) func(Event) bool {
	return func(e Event) bool {
		for _, s := range sources {
			if e.Source == s {
				return true
			}
		}
		return false
	}
}

// FilterEventsAfterTime returns a filter function that only includes events after the specified time
func FilterEventsAfterTime(t time.Time) func(Event) bool {
	return func(e Event) bool {
		return e.Timestamp.After(t)
	}
}

// FilterEventsBeforeTime returns a filter function that only includes events before the specified time
func FilterEventsBeforeTime(t time.Time) func(Event) bool {
	return func(e Event) bool {
		return e.Timestamp.Before(t)
	}
}

// CombineFilters combines multiple filter functions with AND logic
func CombineFilters(filters ...func(Event) bool) func(Event) bool {
	return func(e Event) bool {
		for _, filter := range filters {
			if !filter(e) {
				return false
			}
		}
		return true
	}
}
