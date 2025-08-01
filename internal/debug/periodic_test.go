package debug

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockDebugHandler implements the DebugHandler interface for testing
type MockDebugHandler struct {
	SnapshotTaken        bool
	VisualStateGenerated bool
	LastSnapshot         *Snapshot
	LastVisualState      *VisualState
}

func (m *MockDebugHandler) OnSnapshotTaken(snapshot *Snapshot) {
	m.SnapshotTaken = true
	m.LastSnapshot = snapshot
}

func (m *MockDebugHandler) OnVisualStateGenerated(visualState *VisualState) {
	m.VisualStateGenerated = true
	m.LastVisualState = visualState
}

// TestApp combines all the mock interfaces for testing
type TestApp struct {
	MockSnapshottable
	MockVisualizable
	MockLayoutProvider
	MockMetadataProvider
	MockDebugHandler
}

func TestNewSnapshotManager(t *testing.T) {
	app := &TestApp{}
	interval := 5 * time.Second
	dir := "test-snapshots"

	manager := NewSnapshotManager(app, interval, dir)

	if manager.app != app {
		t.Error("Expected app to be set correctly")
	}

	if manager.interval != interval {
		t.Errorf("Expected interval to be %v, got %v", interval, manager.interval)
	}

	if manager.dir != dir {
		t.Errorf("Expected dir to be %s, got %s", dir, manager.dir)
	}

	if manager.running {
		t.Error("Expected manager to not be running initially")
	}
}

func TestSetMaxSnapshots(t *testing.T) {
	manager := NewSnapshotManager(nil, 0, "")
	manager.SetMaxSnapshots(50)

	if manager.maxSnapshots != 50 {
		t.Errorf("Expected maxSnapshots to be 50, got %d", manager.maxSnapshots)
	}
}

func TestSetFilenamePrefix(t *testing.T) {
	manager := NewSnapshotManager(nil, 0, "")
	manager.SetFilenamePrefix("test-prefix")

	if manager.filenamePrefix != "test-prefix" {
		t.Errorf("Expected filenamePrefix to be 'test-prefix', got %s", manager.filenamePrefix)
	}
}

func TestStartAndStop(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "snapshot_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	app := &TestApp{
		MockSnapshottable: MockSnapshottable{
			ID: "test-app",
			State: map[string]interface{}{
				"key": "value",
			},
		},
	}

	// Use a short interval for testing
	manager := NewSnapshotManager(app, 10*time.Millisecond, tempDir)

	// Create a channel to signal when a snapshot is taken
	doneCh := make(chan struct{})

	// Override the takeSnapshot method to signal when done
	originalCount := manager.snapshotCount
	go func() {
		// Wait until we have at least 1 snapshot
		for {
			manager.mu.Lock()
			count := manager.snapshotCount
			manager.mu.Unlock()

			if count > originalCount {
				close(doneCh)
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Start the manager
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Check that the manager is running
	if !manager.IsRunning() {
		t.Error("Expected manager to be running after Start")
	}

	// Wait for the snapshot to be taken or timeout
	select {
	case <-doneCh:
		// Success, continue
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for snapshot")
	}

	// Stop the manager
	manager.Stop()

	// Check that the manager is not running
	if manager.IsRunning() {
		t.Error("Expected manager to not be running after Stop")
	}

	// Check that at least one snapshot was taken
	if manager.GetSnapshotCount() < 1 {
		t.Errorf("Expected at least 1 snapshot to be taken, got %d", manager.GetSnapshotCount())
	}

	// Check that the snapshot file was created
	files, err := filepath.Glob(filepath.Join(tempDir, "snapshot-*.json"))
	if err != nil {
		t.Fatalf("Failed to list snapshot files: %v", err)
	}

	if len(files) < 1 {
		t.Error("Expected at least one snapshot file to be created")
	}
}

func TestStartPeriodicSnapshots(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "periodic_snapshots_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	app := &TestApp{
		MockSnapshottable: MockSnapshottable{
			ID: "test-app",
			State: map[string]interface{}{
				"key": "value",
			},
		},
	}

	// Start periodic snapshots
	manager, err := StartPeriodicSnapshots(app, 10*time.Millisecond, tempDir)
	if err != nil {
		t.Fatalf("Failed to start periodic snapshots: %v", err)
	}

	// Check that the manager is running
	if !manager.IsRunning() {
		t.Error("Expected manager to be running")
	}

	// Create a channel to signal when a snapshot is taken
	doneCh := make(chan struct{})

	// Wait for at least one snapshot to be taken
	go func() {
		for {
			if manager.GetSnapshotCount() > 0 {
				close(doneCh)
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Wait for the snapshot to be taken or timeout
	select {
	case <-doneCh:
		// Success, continue
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for snapshot")
	}

	// Stop the manager
	manager.Stop()

	// Check that at least one snapshot was taken
	if manager.GetSnapshotCount() < 1 {
		t.Errorf("Expected at least 1 snapshot to be taken, got %d", manager.GetSnapshotCount())
	}
}

func TestDebugHandlerNotification(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "debug_handler_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	app := &TestApp{
		MockSnapshottable: MockSnapshottable{
			ID: "test-app",
			State: map[string]interface{}{
				"key": "value",
			},
		},
	}

	// Start periodic snapshots
	manager := NewSnapshotManager(app, 10*time.Millisecond, tempDir)
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Create a channel to signal when the debug handler is notified
	doneCh := make(chan struct{})

	// Wait for the debug handler to be notified
	go func() {
		for {
			if app.SnapshotTaken {
				close(doneCh)
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Wait for the notification or timeout
	select {
	case <-doneCh:
		// Success, continue
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for debug handler notification")
	}

	// Stop the manager
	manager.Stop()

	// Check that the debug handler was notified
	if !app.SnapshotTaken {
		t.Error("Expected debug handler to be notified of snapshot")
	}

	if app.LastSnapshot == nil {
		t.Error("Expected debug handler to receive the snapshot")
	}
}

func TestCleanupOldSnapshots(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "cleanup_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	app := &TestApp{
		MockSnapshottable: MockSnapshottable{
			ID: "test-app",
			State: map[string]interface{}{
				"key": "value",
			},
		},
	}

	// Create a manager with a max of 2 snapshots
	manager := NewSnapshotManager(app, 10*time.Millisecond, tempDir)
	manager.SetMaxSnapshots(2)

	// Create a channel to signal when snapshots are taken
	doneCh := make(chan struct{})

	// Override the takeSnapshot method to signal when done
	originalCount := manager.snapshotCount
	go func() {
		// Wait until we have at least 3 snapshots
		for {
			manager.mu.Lock()
			count := manager.snapshotCount
			manager.mu.Unlock()

			if count >= originalCount+3 {
				close(doneCh)
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Start the manager
	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}

	// Wait for the snapshots to be taken or timeout
	select {
	case <-doneCh:
		// Success, continue
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out waiting for snapshots")
	}

	// Stop the manager
	manager.Stop()

	// Check that at least 3 snapshots were taken
	if manager.GetSnapshotCount() < 3 {
		t.Errorf("Expected at least 3 snapshots to be taken, got %d", manager.GetSnapshotCount())
	}

	// Check that only 2 snapshot files remain
	files, err := filepath.Glob(filepath.Join(tempDir, "snapshot-*.json"))
	if err != nil {
		t.Fatalf("Failed to list snapshot files: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 snapshot files to remain, got %d", len(files))
	}
}
