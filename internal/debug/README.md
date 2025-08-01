# Debug Package

The `debug` package provides tools for capturing, visualizing, and analyzing the state of AWSM's TUI application. It's designed to help AI tools and developers understand the application state through snapshots and visual representations.

## Components

### State Snapshots

The package includes functionality to capture the entire application state as snapshots:

- `Snapshot` struct that stores application state with timestamps and unique IDs
- Serialization/deserialization to/from JSON
- Disk storage with automatic directory management
- Periodic snapshot capture with configurable intervals

### Visual State Representation

The package provides tools to generate text-based visual representations of the TUI:

- `VisualState` struct that represents the TUI visually as text
- ASCII/text representations of UI components
- Support for different detail levels (minimal, normal, detailed)
- Box drawing and table generation utilities

### Integration Interfaces

The package defines interfaces that TUI components can implement:

- `Snapshottable` for components that can provide state for snapshots
- `Visualizable` for components that can provide visual representations
- `DebugCapable` combining both interfaces
- `LayoutProvider` for components that can provide layout information
- `MetadataProvider` for components that can provide metadata
- `DebugHandler` for components that can handle debug events

## Usage

### Capturing a Snapshot

```go
// Capture a snapshot of the application
snapshot, err := debug.CaptureSnapshot(app)
if err != nil {
    log.Fatalf("Failed to capture snapshot: %v", err)
}

// Save the snapshot to disk
err = debug.SaveSnapshot(snapshot, "snapshots/latest.json")
if err != nil {
    log.Fatalf("Failed to save snapshot: %v", err)
}
```

### Generating a Visual Representation

```go
// Generate a visual representation of the application
visualState, err := debug.GenerateVisualState(app, debug.NormalDetail)
if err != nil {
    log.Fatalf("Failed to generate visual state: %v", err)
}

// Print the visual representation
fmt.Println(visualState.String())
```

### Starting Periodic Snapshots

```go
// Start taking periodic snapshots every 5 seconds
manager, err := debug.StartPeriodicSnapshots(app, 5*time.Second, "snapshots")
if err != nil {
    log.Fatalf("Failed to start periodic snapshots: %v", err)
}

// Stop taking snapshots when done
defer manager.Stop()
```

### Implementing the Interfaces

To fully integrate a component with the debug package, implement the `DebugCapable` interface:

```go
// MyComponent implements debug.DebugCapable
type MyComponent struct {
    // Component fields
}

// GetSnapshotState returns a serializable representation of the component's state
func (c *MyComponent) GetSnapshotState() interface{} {
    return map[string]interface{}{
        "field1": c.Field1,
        "field2": c.Field2,
    }
}

// GetSnapshotID returns a unique identifier for this component
func (c *MyComponent) GetSnapshotID() string {
    return "my-component"
}

// GetVisualRepresentation returns a text representation of the component
func (c *MyComponent) GetVisualRepresentation(detailLevel debug.DetailLevel) string {
    // Generate a visual representation based on the detail level
    return fmt.Sprintf("MyComponent: %s", c.Field1)
}

// GetVisualDimensions returns the width and height of the component
func (c *MyComponent) GetVisualDimensions() (width, height int) {
    return 80, 10
}

// GetVisualID returns a unique identifier for this component
func (c *MyComponent) GetVisualID() string {
    return "my-component"
}
```

## Benefits for AI Tools

This package is specifically designed to help AI tools "see" and understand the TUI application:

1. **State Transparency**: Snapshots provide a clear view of the application's internal state
2. **Visual Context**: Text representations help AI tools understand the visual layout
3. **Temporal Analysis**: Periodic snapshots enable tracking state changes over time
4. **Structured Data**: JSON serialization makes the state easy to parse and analyze
5. **Minimal Impact**: Integration interfaces minimize changes to the main application code

By implementing these interfaces in your TUI components, you enable AI tools to better understand and interact with your application.