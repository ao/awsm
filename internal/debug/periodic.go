package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// SnapshotManager handles periodic snapshots of the application state.
type SnapshotManager struct {
	// app is the application to snapshot
	app interface{}

	// interval is the time between snapshots
	interval time.Duration

	// dir is the directory to save snapshots to
	dir string

	// maxSnapshots is the maximum number of snapshots to keep
	maxSnapshots int

	// filenamePrefix is the prefix for snapshot filenames
	filenamePrefix string

	// stopChan is used to signal the snapshot goroutine to stop
	stopChan chan struct{}

	// wg is used to wait for the snapshot goroutine to finish
	wg sync.WaitGroup

	// mu protects the fields below
	mu sync.Mutex

	// running indicates whether the snapshot goroutine is running
	running bool

	// lastSnapshot is the last snapshot taken
	lastSnapshot *Snapshot

	// snapshotCount is the number of snapshots taken
	snapshotCount int
}

// NewSnapshotManager creates a new SnapshotManager.
func NewSnapshotManager(app interface{}, interval time.Duration, dir string) *SnapshotManager {
	return &SnapshotManager{
		app:            app,
		interval:       interval,
		dir:            dir,
		maxSnapshots:   100, // Default to keeping 100 snapshots
		filenamePrefix: "snapshot",
		stopChan:       make(chan struct{}),
	}
}

// SetMaxSnapshots sets the maximum number of snapshots to keep.
func (sm *SnapshotManager) SetMaxSnapshots(max int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.maxSnapshots = max
}

// SetFilenamePrefix sets the prefix for snapshot filenames.
func (sm *SnapshotManager) SetFilenamePrefix(prefix string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.filenamePrefix = prefix
}

// Start starts taking periodic snapshots.
func (sm *SnapshotManager) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("snapshot manager is already running")
	}

	// Ensure the snapshot directory exists
	if err := os.MkdirAll(sm.dir, 0755); err != nil {
		return fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	sm.running = true
	sm.wg.Add(1)

	go sm.run()

	return nil
}

// Stop stops taking periodic snapshots.
func (sm *SnapshotManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return
	}

	close(sm.stopChan)
	sm.wg.Wait()

	sm.running = false
	sm.stopChan = make(chan struct{})
}

// run is the main loop for taking periodic snapshots.
func (sm *SnapshotManager) run() {
	defer sm.wg.Done()

	ticker := time.NewTicker(sm.interval)
	defer ticker.Stop()

	// Take an initial snapshot
	sm.takeSnapshot()

	for {
		select {
		case <-ticker.C:
			sm.takeSnapshot()
		case <-sm.stopChan:
			return
		}
	}
}

// takeSnapshot takes a snapshot and saves it to disk.
func (sm *SnapshotManager) takeSnapshot() {
	snapshot, err := CaptureSnapshot(sm.app)
	if err != nil {
		fmt.Printf("Failed to capture snapshot: %v\n", err)
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.lastSnapshot = snapshot
	sm.snapshotCount++

	// Generate a filename for the snapshot
	filename := snapshot.GenerateFilename(sm.filenamePrefix)
	path := filepath.Join(sm.dir, filename)

	// Save the snapshot to disk
	if err := SaveSnapshot(snapshot, path); err != nil {
		fmt.Printf("Failed to save snapshot: %v\n", err)
		return
	}

	// If we have a DebugHandler, notify it
	if handler, ok := sm.app.(DebugHandler); ok {
		handler.OnSnapshotTaken(snapshot)
	}

	// Clean up old snapshots if we have too many
	if sm.maxSnapshots > 0 {
		sm.cleanupOldSnapshots()
	}
}

// cleanupOldSnapshots removes old snapshots if we have too many.
func (sm *SnapshotManager) cleanupOldSnapshots() {
	// List all snapshot files
	pattern := filepath.Join(sm.dir, sm.filenamePrefix+"-*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Failed to list snapshot files: %v\n", err)
		return
	}

	// If we have fewer snapshots than the maximum, we're done
	if len(files) <= sm.maxSnapshots {
		return
	}

	// Get file info for each snapshot
	type fileInfo struct {
		path    string
		modTime time.Time
	}

	fileInfos := make([]fileInfo, 0, len(files))
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			fmt.Printf("Failed to stat file %s: %v\n", file, err)
			continue
		}

		fileInfos = append(fileInfos, fileInfo{
			path:    file,
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time (oldest first)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.Before(fileInfos[j].modTime)
	})

	// Remove the oldest snapshots
	for i := 0; i < len(fileInfos)-sm.maxSnapshots; i++ {
		if err := os.Remove(fileInfos[i].path); err != nil {
			fmt.Printf("Failed to remove old snapshot %s: %v\n", fileInfos[i].path, err)
		}
	}
}

// GetLastSnapshot returns the last snapshot taken.
func (sm *SnapshotManager) GetLastSnapshot() *Snapshot {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.lastSnapshot
}

// GetSnapshotCount returns the number of snapshots taken.
func (sm *SnapshotManager) GetSnapshotCount() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.snapshotCount
}

// IsRunning returns whether the snapshot manager is running.
func (sm *SnapshotManager) IsRunning() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.running
}

// StartPeriodicSnapshots is a convenience function to start periodic snapshots.
// It creates a new SnapshotManager and starts it.
func StartPeriodicSnapshots(app interface{}, interval time.Duration, dir string) (*SnapshotManager, error) {
	manager := NewSnapshotManager(app, interval, dir)
	if err := manager.Start(); err != nil {
		return nil, err
	}
	return manager, nil
}
