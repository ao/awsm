# Events Package

The `events` package provides tools for capturing, recording, and replaying events in the AWSM TUI application. It's designed to help AI tools understand the flow of events and state changes in the application.

## Components

### Event Types

The package defines several types of events:

- **InputEvent**: Represents user input (key presses, mouse events)
- **OutputEvent**: Represents application output (screen updates)
- **StateChangeEvent**: Represents internal state changes
- **CommandEvent**: Represents commands being executed

### Event Recording

The `EventRecorder` provides functionality to capture and store events:

- Recording of different event types
- Serialization/deserialization to/from JSON
- Saving recordings to disk
- Filtering options for specific event types

### Bubble Tea Integration

The package includes middleware for Bubble Tea to intercept messages:

- `RecorderMiddleware`: Records input and output events
- `StateChangeMiddleware`: Records state changes
- `NewRecordingMiddleware`: Combines both middleware types
- `NewReplayMiddleware`: Creates middleware for replaying events

## Usage

### Basic Recording

```go
// Create a new event recorder
recorder := events.NewEventRecorder()

// Start recording events
recorder.Start()

// Record events
recorder.RecordInput("component", msg)
recorder.RecordOutput("component", view)
recorder.RecordStateChange("component", before, after)
recorder.RecordCommand("component", "command", []string{"arg1", "arg2"})

// Stop recording
recorder.Stop()

// Save events to disk
recorder.SaveToFile("events.json")
```

### Using Middleware

```go
// Create a recorder
recorder := events.NewEventRecorder()
recorder.Start()

// Create middleware
middleware := events.NewRecordingMiddleware(recorder, "my_app")

// Create your model
model := yourModel{}

// Create and run the program with middleware
p := tea.NewProgram(middleware(model))
p.Run()

// Save events when done
recorder.SaveToFile("my_app_events.json")
```

### Replaying Events

```go
// Load events from disk
loadedEvents, err := events.LoadEventsFromFile("events.json")
if err != nil {
    // Handle error
}

// Filter events if needed
inputEvents := make([]events.Event, 0)
for _, e := range loadedEvents {
    if e.Type == events.EventTypeInput {
        inputEvents = append(inputEvents, e)
    }
}

// Create a model for replay
replayModel := yourModel{}

// Create replay middleware
replayMiddleware := events.NewReplayMiddleware(inputEvents)

// Run the replay
replayProgram := tea.NewProgram(replayMiddleware(replayModel))
replayProgram.Run()
```

### Filtering Events

The package provides several filter functions:

```go
// Filter by event type
typeFilter := events.FilterEventsByType(events.EventTypeInput, events.EventTypeOutput)

// Filter by source
sourceFilter := events.FilterEventsBySource("component1", "component2")

// Filter by time
timeFilter := events.FilterEventsAfterTime(startTime)

// Combine filters
combinedFilter := events.CombineFilters(
    events.FilterEventsByType(events.EventTypeInput),
    events.FilterEventsBySource("component1"),
)
```

## Benefits for AI Tools

This package is specifically designed to help AI tools understand the TUI application:

1. **Event Transparency**: Recordings provide a clear view of user interactions
2. **Flow Analysis**: Event sequences help understand application flow
3. **State Tracking**: State change events enable tracking changes over time
4. **Structured Data**: JSON serialization makes events easy to parse and analyze
5. **Minimal Impact**: Middleware integration minimizes changes to the main application code

By recording and analyzing events, AI tools can better understand how users interact with the application and how the application responds to those interactions.