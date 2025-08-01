package debug

import (
	"testing"
)

// MockSnapshottable implements the Snapshottable interface for testing
type MockSnapshottable struct {
	ID    string
	State map[string]interface{}
}

func (m *MockSnapshottable) GetSnapshotState() interface{} {
	return m.State
}

func (m *MockSnapshottable) GetSnapshotID() string {
	return m.ID
}

// MockVisualizable implements the Visualizable interface for testing
type MockVisualizable struct {
	ID             string
	Representation string
	Width, Height  int
}

func (m *MockVisualizable) GetVisualRepresentation(detailLevel DetailLevel) string {
	return m.Representation
}

func (m *MockVisualizable) GetVisualDimensions() (width, height int) {
	return m.Width, m.Height
}

func (m *MockVisualizable) GetVisualID() string {
	return m.ID
}

// MockDebugCapable implements both Snapshottable and Visualizable for testing
type MockDebugCapable struct {
	MockSnapshottable
	MockVisualizable
}

// MockLayoutProvider implements the LayoutProvider interface for testing
type MockLayoutProvider struct {
	Layout string
}

func (m *MockLayoutProvider) GetLayoutDescription() string {
	return m.Layout
}

// MockMetadataProvider implements the MetadataProvider interface for testing
type MockMetadataProvider struct {
	Metadata map[string]string
}

func (m *MockMetadataProvider) GetSnapshotMetadata() map[string]string {
	return m.Metadata
}

// TestStruct is a simple struct for testing reflection-based state extraction
type TestStruct struct {
	PublicField  string
	privateField string
	NestedStruct struct {
		Field string
	}
}

func TestCaptureSnapshotWithSnapshottable(t *testing.T) {
	mock := &MockSnapshottable{
		ID: "test-component",
		State: map[string]interface{}{
			"key": "value",
		},
	}

	snapshot, err := CaptureSnapshot(mock)
	if err != nil {
		t.Fatalf("Failed to capture snapshot: %v", err)
	}

	// Check that the snapshot contains the state from the mock
	if state, ok := snapshot.AppState["test-component"]; !ok {
		t.Error("Expected snapshot to contain state with ID 'test-component'")
	} else {
		stateMap, ok := state.(map[string]interface{})
		if !ok {
			t.Errorf("Expected state to be a map, got %T", state)
		} else if stateMap["key"] != "value" {
			t.Errorf("Expected state to contain key 'key' with value 'value', got %v", stateMap["key"])
		}
	}
}

func TestCaptureSnapshotWithMetadataProvider(t *testing.T) {
	mock := &MockMetadataProvider{
		Metadata: map[string]string{
			"version": "1.0",
			"user":    "test",
		},
	}

	snapshot, err := CaptureSnapshot(mock)
	if err != nil {
		t.Fatalf("Failed to capture snapshot: %v", err)
	}

	// Check that the snapshot contains the metadata from the mock
	if snapshot.Metadata["version"] != "1.0" {
		t.Errorf("Expected metadata to contain key 'version' with value '1.0', got %v", snapshot.Metadata["version"])
	}

	if snapshot.Metadata["user"] != "test" {
		t.Errorf("Expected metadata to contain key 'user' with value 'test', got %v", snapshot.Metadata["user"])
	}
}

func TestCaptureSnapshotWithStruct(t *testing.T) {
	testStruct := TestStruct{
		PublicField:  "public",
		privateField: "private",
	}
	testStruct.NestedStruct.Field = "nested"

	snapshot, err := CaptureSnapshot(testStruct)
	if err != nil {
		t.Fatalf("Failed to capture snapshot: %v", err)
	}

	// Check that the snapshot contains the public fields but not the private ones
	if val, ok := snapshot.AppState["PublicField"]; !ok {
		t.Error("Expected snapshot to contain PublicField")
	} else if val != "public" {
		t.Errorf("Expected PublicField to be 'public', got %v", val)
	}

	if _, ok := snapshot.AppState["privateField"]; ok {
		t.Error("Expected snapshot not to contain privateField")
	}
}

func TestGenerateVisualStateWithVisualizable(t *testing.T) {
	mock := &MockVisualizable{
		ID:             "test-component",
		Representation: "test representation",
		Width:          100,
		Height:         50,
	}

	visualState, err := GenerateVisualState(mock, NormalDetail)
	if err != nil {
		t.Fatalf("Failed to generate visual state: %v", err)
	}

	// Check that the visual state has the correct dimensions
	if visualState.Width != 100 {
		t.Errorf("Expected Width to be 100, got %d", visualState.Width)
	}

	if visualState.Height != 50 {
		t.Errorf("Expected Height to be 50, got %d", visualState.Height)
	}

	// Check that the visual state contains the representation from the mock
	if rep, ok := visualState.Components["test-component"]; !ok {
		t.Error("Expected visual state to contain component with ID 'test-component'")
	} else if rep != "test representation" {
		t.Errorf("Expected representation to be 'test representation', got %s", rep)
	}
}

func TestGenerateVisualStateWithLayoutProvider(t *testing.T) {
	mock := &struct {
		MockVisualizable
		MockLayoutProvider
	}{
		MockVisualizable: MockVisualizable{
			ID:             "test-component",
			Representation: "test representation",
			Width:          100,
			Height:         50,
		},
		MockLayoutProvider: MockLayoutProvider{
			Layout: "test layout",
		},
	}

	visualState, err := GenerateVisualState(mock, NormalDetail)
	if err != nil {
		t.Fatalf("Failed to generate visual state: %v", err)
	}

	// Check that the visual state contains the layout from the mock
	if visualState.Layout != "test layout" {
		t.Errorf("Expected Layout to be 'test layout', got %s", visualState.Layout)
	}
}

func TestFindVisualizableComponents(t *testing.T) {
	// Create a struct with nested Visualizable components
	root := &struct {
		Component1      *MockVisualizable
		Component2      *MockVisualizable
		NotVisualizable string
		Nested          struct {
			Component3 *MockVisualizable
		}
	}{
		Component1:      &MockVisualizable{ID: "component1"},
		Component2:      &MockVisualizable{ID: "component2"},
		NotVisualizable: "not visualizable",
	}
	root.Nested.Component3 = &MockVisualizable{ID: "component3"}

	// Find all Visualizable components
	components := FindVisualizableComponents(root)

	// Check that the components have the correct IDs
	ids := make(map[string]bool)
	for _, c := range components {
		ids[c.GetVisualID()] = true
	}

	// Check that we found the expected components
	if !ids["component1"] {
		t.Error("Expected to find component1")
	}

	if !ids["component2"] {
		t.Error("Expected to find component2")
	}

	if !ids["component3"] {
		t.Error("Expected to find component3")
	}
}

func TestFindSnapshottableComponents(t *testing.T) {
	// Create a struct with nested Snapshottable components
	root := &struct {
		Component1       *MockSnapshottable
		Component2       *MockSnapshottable
		NotSnapshottable string
		Nested           struct {
			Component3 *MockSnapshottable
		}
	}{
		Component1:       &MockSnapshottable{ID: "component1"},
		Component2:       &MockSnapshottable{ID: "component2"},
		NotSnapshottable: "not snapshottable",
	}
	root.Nested.Component3 = &MockSnapshottable{ID: "component3"}

	// Find all Snapshottable components
	components := FindSnapshottableComponents(root)

	// Check that the components have the correct IDs
	ids := make(map[string]bool)
	for _, c := range components {
		ids[c.GetSnapshotID()] = true
	}

	// Check that we found the expected components
	if !ids["component1"] {
		t.Error("Expected to find component1")
	}

	if !ids["component2"] {
		t.Error("Expected to find component2")
	}

	if !ids["component3"] {
		t.Error("Expected to find component3")
	}
}
