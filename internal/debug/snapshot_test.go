package debug

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSnapshot(t *testing.T) {
	snapshot := NewSnapshot()

	if snapshot.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if snapshot.AppState == nil {
		t.Error("Expected non-nil AppState")
	}

	if snapshot.Metadata == nil {
		t.Error("Expected non-nil Metadata")
	}
}

func TestAddState(t *testing.T) {
	snapshot := NewSnapshot()
	snapshot.AddState("test", "value")

	if val, ok := snapshot.AppState["test"]; !ok {
		t.Error("Expected state to be added")
	} else if val != "value" {
		t.Errorf("Expected value to be 'value', got %v", val)
	}
}

func TestAddMetadata(t *testing.T) {
	snapshot := NewSnapshot()
	snapshot.AddMetadata("test", "value")

	if val, ok := snapshot.Metadata["test"]; !ok {
		t.Error("Expected metadata to be added")
	} else if val != "value" {
		t.Errorf("Expected value to be 'value', got %v", val)
	}
}

func TestToJSON(t *testing.T) {
	snapshot := NewSnapshot()
	snapshot.AddState("test", "value")
	snapshot.AddMetadata("test", "value")

	data, err := snapshot.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize snapshot: %v", err)
	}

	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if unmarshalled["id"] == nil {
		t.Error("Expected ID in JSON")
	}

	if unmarshalled["timestamp"] == nil {
		t.Error("Expected timestamp in JSON")
	}

	if appState, ok := unmarshalled["app_state"].(map[string]interface{}); !ok {
		t.Error("Expected app_state in JSON")
	} else if appState["test"] != "value" {
		t.Errorf("Expected test value to be 'value', got %v", appState["test"])
	}

	if metadata, ok := unmarshalled["metadata"].(map[string]interface{}); !ok {
		t.Error("Expected metadata in JSON")
	} else if metadata["test"] != "value" {
		t.Errorf("Expected test value to be 'value', got %v", metadata["test"])
	}
}

func TestFromJSON(t *testing.T) {
	original := NewSnapshot()
	original.AddState("test", "value")
	original.AddMetadata("test", "value")

	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize snapshot: %v", err)
	}

	snapshot, err := FromJSON(data)
	if err != nil {
		t.Fatalf("Failed to deserialize snapshot: %v", err)
	}

	if snapshot.ID != original.ID {
		t.Errorf("Expected ID to be %s, got %s", original.ID, snapshot.ID)
	}

	if !snapshot.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Expected timestamp to be %v, got %v", original.Timestamp, snapshot.Timestamp)
	}

	if val, ok := snapshot.AppState["test"]; !ok {
		t.Error("Expected state to be present")
	} else if val != "value" {
		t.Errorf("Expected value to be 'value', got %v", val)
	}

	if val, ok := snapshot.Metadata["test"]; !ok {
		t.Error("Expected metadata to be present")
	} else if val != "value" {
		t.Errorf("Expected value to be 'value', got %v", val)
	}
}

func TestSaveAndLoadSnapshot(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "snapshot_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a snapshot
	snapshot := NewSnapshot()
	snapshot.AddState("test", "value")
	snapshot.AddMetadata("test", "value")

	// Save the snapshot
	path := filepath.Join(tempDir, "snapshot.json")
	if err := SaveSnapshot(snapshot, path); err != nil {
		t.Fatalf("Failed to save snapshot: %v", err)
	}

	// Load the snapshot
	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("Failed to load snapshot: %v", err)
	}

	// Verify the loaded snapshot
	if loaded.ID != snapshot.ID {
		t.Errorf("Expected ID to be %s, got %s", snapshot.ID, loaded.ID)
	}

	if !loaded.Timestamp.Equal(snapshot.Timestamp) {
		t.Errorf("Expected timestamp to be %v, got %v", snapshot.Timestamp, loaded.Timestamp)
	}

	if val, ok := loaded.AppState["test"]; !ok {
		t.Error("Expected state to be present")
	} else if val != "value" {
		t.Errorf("Expected value to be 'value', got %v", val)
	}

	if val, ok := loaded.Metadata["test"]; !ok {
		t.Error("Expected metadata to be present")
	} else if val != "value" {
		t.Errorf("Expected value to be 'value', got %v", val)
	}
}

func TestGenerateFilename(t *testing.T) {
	snapshot := NewSnapshot()
	snapshot.Timestamp = time.Date(2023, 1, 2, 3, 4, 5, 0, time.UTC)

	filename := snapshot.GenerateFilename("test")
	expected := "test-20230102-030405-" + snapshot.ID[:8] + ".json"

	if filename != expected {
		t.Errorf("Expected filename to be %s, got %s", expected, filename)
	}
}
