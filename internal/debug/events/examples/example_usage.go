package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ao/awsm/internal/debug/events"
	"github.com/ao/awsm/internal/logger"
	tea "github.com/charmbracelet/bubbletea"
)

// Simple counter model for demonstration
type counterModel struct {
	count int
}

func (m counterModel) Init() tea.Cmd {
	return nil
}

func (m counterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "+":
			m.count++
		case "down", "-":
			m.count--
		}
	}
	return m, nil
}

func (m counterModel) View() string {
	return fmt.Sprintf("\n Count: %d\n\n ↑/+: increment\n ↓/-: decrement\n q: quit\n", m.count)
}

func main() {
	// Initialize logger
	if err := logger.Initialize(); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Create an event recorder
	recorder := events.NewEventRecorder()

	// Start recording events
	recorder.Start()
	defer recorder.Stop()

	// Set up automatic saving of events every 5 seconds
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if recorder.IsActive() {
				if err := recorder.SaveToFile("events.json"); err != nil {
					logger.Error("Failed to save events: %v", err)
				}
			}
		}
	}()

	// Create a simple counter model
	model := counterModel{}

	// Wrap the model with recording middleware
	recordingMiddleware := events.NewRecordingMiddleware(recorder, "counter_app")

	// Create and run the Bubble Tea program with the middleware
	p := tea.NewProgram(recordingMiddleware(model))

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Save the final events
	if err := recorder.SaveToFile("events_final.json"); err != nil {
		fmt.Printf("Failed to save final events: %v\n", err)
	} else {
		fmt.Println("Events saved to events_final.json")
	}

	// Example of loading and replaying events
	fmt.Println("Loading events for replay...")
	loadedEvents, err := events.LoadEventsFromFile("events_final.json")
	if err != nil {
		fmt.Printf("Failed to load events: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d events\n", len(loadedEvents))

	// Filter events if needed
	inputEvents := make([]events.Event, 0)
	for _, e := range loadedEvents {
		if e.Type == events.EventTypeInput {
			inputEvents = append(inputEvents, e)
		}
	}

	fmt.Printf("Found %d input events for replay\n", len(inputEvents))

	// Create a new program for replay
	replayModel := counterModel{}
	replayMiddleware := events.NewReplayMiddleware(inputEvents)

	fmt.Println("Press Enter to start replay...")
	fmt.Scanln()

	// Run the replay
	replayProgram := tea.NewProgram(replayMiddleware(replayModel))
	if _, err := replayProgram.Run(); err != nil {
		fmt.Printf("Error running replay: %v\n", err)
		os.Exit(1)
	}
}

// Example of how to use the event recorder in an existing application
func ExampleIntegration() {
	// Create an event recorder
	recorder := events.NewEventRecorder()

	// Start recording
	recorder.Start()

	// Create middleware
	middleware := events.NewRecordingMiddleware(recorder, "my_app")

	// Create your model
	model := yourModel{} // Replace with your actual model

	// Create and run the program with middleware
	p := tea.NewProgram(middleware(model))
	p.Run()

	// Save events when done
	recorder.SaveToFile("my_app_events.json")
}

// Example of how to record specific events manually
func ExampleManualRecording() {
	// Create an event recorder
	recorder := events.NewEventRecorder()

	// Start recording
	recorder.Start()

	// Record input event
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	recorder.RecordInput("manual_example", keyMsg)

	// Record output event
	recorder.RecordOutput("manual_example", "Hello, world!")

	// Record state change
	before := map[string]interface{}{"count": 1}
	after := map[string]interface{}{"count": 2}
	recorder.RecordStateChange("manual_example", before, after)

	// Record command
	recorder.RecordCommand("manual_example", "echo", []string{"hello"})

	// Save events
	recorder.SaveToFile("manual_events.json")
}

// Placeholder for your actual model
type yourModel struct{}

func (m yourModel) Init() tea.Cmd                           { return nil }
func (m yourModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m yourModel) View() string                            { return "" }
