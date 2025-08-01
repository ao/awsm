// Package examples provides examples of how to use the debug package.
// This file demonstrates how to implement the debug interfaces in a TUI application.
package examples

import (
	"fmt"
	"time"
)

// This is a simplified example that shows how to implement the debug interfaces.
// In a real application, you would import the debug package and implement these interfaces.

// DetailLevel represents the level of detail for visual state representation.
type DetailLevel int

const (
	// MinimalDetail provides a minimal representation with just essential information.
	MinimalDetail DetailLevel = iota
	// NormalDetail provides a standard representation with moderate details.
	NormalDetail
	// DetailedDetail provides a comprehensive representation with all available details.
	DetailedDetail
)

// Snapshot represents a complete capture of the application state at a specific point in time.
type Snapshot struct {
	ID        string
	Timestamp time.Time
	AppState  map[string]interface{}
	Metadata  map[string]string
}

// VisualState represents a text-based visual representation of the TUI application state.
type VisualState struct {
	Components  map[string]string
	Layout      string
	DetailLevel DetailLevel
	Width       int
	Height      int
}

// SimpleModel is a simple TUI model that implements the debug interfaces
type SimpleModel struct {
	Items    []Item
	Width    int
	Height   int
	Selected string
}

// Item represents an item in our list
type Item struct {
	Title       string
	Description string
}

// GetSnapshotState implements Snapshottable interface
func (m SimpleModel) GetSnapshotState() interface{} {
	items := make([]map[string]string, 0, len(m.Items))
	for _, item := range m.Items {
		items = append(items, map[string]string{
			"title":       item.Title,
			"description": item.Description,
		})
	}

	return map[string]interface{}{
		"items":    items,
		"selected": m.Selected,
		"width":    m.Width,
		"height":   m.Height,
	}
}

// GetSnapshotID implements Snapshottable interface
func (m SimpleModel) GetSnapshotID() string {
	return "simple-model"
}

// GetVisualRepresentation implements Visualizable interface
func (m SimpleModel) GetVisualRepresentation(detailLevel DetailLevel) string {
	var representation string

	// Header
	representation += "Example TUI Application\n\n"

	// List items
	representation += "Items:\n"
	for _, item := range m.Items {
		prefix := "  "
		if item.Title == m.Selected {
			prefix = "> "
		}
		representation += fmt.Sprintf("%s%s - %s\n", prefix, item.Title, item.Description)
	}

	// Selected item
	representation += fmt.Sprintf("\nSelected: %s\n", m.Selected)

	return representation
}

// GetVisualDimensions implements Visualizable interface
func (m SimpleModel) GetVisualDimensions() (width, height int) {
	return m.Width, m.Height
}

// GetVisualID implements Visualizable interface
func (m SimpleModel) GetVisualID() string {
	return "simple-model"
}

// GetLayoutDescription implements LayoutProvider interface
func (m SimpleModel) GetLayoutDescription() string {
	return fmt.Sprintf("Simple layout with a list (%dx%d)", m.Width, m.Height)
}

// GetSnapshotMetadata implements MetadataProvider interface
func (m SimpleModel) GetSnapshotMetadata() map[string]string {
	return map[string]string{
		"app_name":   "Example TUI",
		"item_count": fmt.Sprintf("%d", len(m.Items)),
		"selected":   m.Selected,
		"dimensions": fmt.Sprintf("%dx%d", m.Width, m.Height),
		"timestamp":  time.Now().Format(time.RFC3339),
	}
}

// OnSnapshotTaken implements DebugHandler interface
func (m SimpleModel) OnSnapshotTaken(snapshot *Snapshot) {
	fmt.Printf("Snapshot taken: %s\n", snapshot.ID)
}

// OnVisualStateGenerated implements DebugHandler interface
func (m SimpleModel) OnVisualStateGenerated(visualState *VisualState) {
	fmt.Printf("Visual state generated with %d components\n", len(visualState.Components))
}

// ExampleUsage demonstrates how to use the debug package with a TUI application
func ExampleUsage() {
	// Create a model
	model := SimpleModel{
		Items: []Item{
			{Title: "Item 1", Description: "First item"},
			{Title: "Item 2", Description: "Second item"},
			{Title: "Item 3", Description: "Third item"},
		},
		Width:    80,
		Height:   24,
		Selected: "Item 1",
	}

	// In a real application, you would:
	// 1. Create a snapshot directory
	// 2. Start periodic snapshots
	// 3. Generate visual states as needed
	// 4. Use the debug package to help understand the application state

	// Example code (not actually executed):
	/*
		// Create a snapshot directory
		snapshotDir := "snapshots"
		if err := os.MkdirAll(snapshotDir, 0755); err != nil {
			fmt.Printf("Failed to create snapshot directory: %v\n", err)
			return
		}

		// Start periodic snapshots
		manager, err := debug.StartPeriodicSnapshots(model, 5*time.Second, snapshotDir)
		if err != nil {
			fmt.Printf("Failed to start periodic snapshots: %v\n", err)
			return
		}
		defer manager.Stop()

		// Generate a visual state
		visualState, err := debug.GenerateVisualState(model, debug.NormalDetail)
		if err != nil {
			fmt.Printf("Failed to generate visual state: %v\n", err)
			return
		}

		// Print the visual state
		fmt.Println("Visual State:")
		fmt.Println(visualState.String())
	*/

	// Just to use the model variable and avoid the "declared and not used" error
	_ = model
}
