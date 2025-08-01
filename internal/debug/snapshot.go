package debug

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Snapshot represents a complete capture of the application state at a specific point in time.
// It can be serialized to JSON and saved to disk for later analysis.
type Snapshot struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	AppState  map[string]interface{} `json:"app_state"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// NewSnapshot creates a new snapshot with a unique ID and the current timestamp.
func NewSnapshot() *Snapshot {
	return &Snapshot{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		AppState:  make(map[string]interface{}),
		Metadata:  make(map[string]string),
	}
}

// AddState adds a named state component to the snapshot.
func (s *Snapshot) AddState(name string, state interface{}) {
	s.AppState[name] = state
}

// AddMetadata adds metadata information to the snapshot.
func (s *Snapshot) AddMetadata(key, value string) {
	s.Metadata[key] = value
}

// ToJSON serializes the snapshot to JSON.
func (s *Snapshot) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON deserializes a snapshot from JSON.
func FromJSON(data []byte) (*Snapshot, error) {
	var snapshot Snapshot
	err := json.Unmarshal(data, &snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}
	return &snapshot, nil
}

// SaveSnapshot saves a snapshot to disk at the specified path.
func SaveSnapshot(snapshot *Snapshot, path string) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Serialize the snapshot to JSON
	data, err := snapshot.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize snapshot: %w", err)
	}

	// Write the data to the file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot to %s: %w", path, err)
	}

	return nil
}

// LoadSnapshot loads a snapshot from disk at the specified path.
func LoadSnapshot(path string) (*Snapshot, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read snapshot from %s: %w", path, err)
	}

	// Deserialize the snapshot from JSON
	return FromJSON(data)
}

// GenerateFilename generates a filename for a snapshot based on its ID and timestamp.
func (s *Snapshot) GenerateFilename(prefix string) string {
	timestamp := s.Timestamp.Format("20060102-150405")
	return fmt.Sprintf("%s-%s-%s.json", prefix, timestamp, s.ID[:8])
}
