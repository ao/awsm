package debug

// Snapshottable is an interface that components can implement to provide state for snapshots.
// Any component that wants to be included in state snapshots should implement this interface.
type Snapshottable interface {
	// GetSnapshotState returns a serializable representation of the component's state.
	// The returned value should be JSON-serializable.
	GetSnapshotState() interface{}

	// GetSnapshotID returns a unique identifier for this component in the snapshot.
	// This ID will be used as the key in the snapshot's AppState map.
	GetSnapshotID() string
}

// Visualizable is an interface that components can implement to provide visual representation.
// Any component that wants to be included in visual state representations should implement this interface.
type Visualizable interface {
	// GetVisualRepresentation returns a text representation of the component.
	// The detail level parameter controls how much detail to include.
	GetVisualRepresentation(detailLevel DetailLevel) string

	// GetVisualDimensions returns the width and height of the component in characters.
	GetVisualDimensions() (width, height int)

	// GetVisualID returns a unique identifier for this component in the visual representation.
	// This ID will be used as the key in the VisualState's Components map.
	GetVisualID() string
}

// DebugCapable is an interface that combines both Snapshottable and Visualizable.
// Components that implement this interface can be fully integrated with the debug package.
type DebugCapable interface {
	Snapshottable
	Visualizable
}

// LayoutProvider is an interface for components that can provide layout information.
// This is typically implemented by the main application or top-level components.
type LayoutProvider interface {
	// GetLayoutDescription returns a text description of the component layout.
	GetLayoutDescription() string
}

// MetadataProvider is an interface for components that can provide metadata for snapshots.
// This is typically implemented by the main application or components with important metadata.
type MetadataProvider interface {
	// GetSnapshotMetadata returns a map of metadata key-value pairs for the snapshot.
	GetSnapshotMetadata() map[string]string
}

// DebugHandler is an interface for components that can handle debug operations.
// This is typically implemented by the main application.
type DebugHandler interface {
	// OnSnapshotTaken is called when a snapshot is taken.
	// The component can use this to perform additional operations or logging.
	OnSnapshotTaken(snapshot *Snapshot)

	// OnVisualStateGenerated is called when a visual state is generated.
	// The component can use this to perform additional operations or logging.
	OnVisualStateGenerated(visualState *VisualState)
}
