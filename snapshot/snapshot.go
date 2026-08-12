package snapshot

import (
	"fmt"
	"sync"
	"time"

	"example.com/eventledger/model"
)

type Snapshot struct {
	ID      uint64
	Reason  string
	Created time.Time
	Values  []model.EventRecord
}

type Manager struct {
	mu        sync.RWMutex
	clock     func() time.Time
	next      uint64
	snapshots []Snapshot
	limit     int
}

func New(limit int) *Manager {
	if limit < 1 {
		limit = 1
	}
	return &Manager{clock: time.Now, limit: limit}
}

func (manager *Manager) Capture(reason string, values []model.EventRecord) Snapshot {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.next++
	snapshot := Snapshot{ID: manager.next, Reason: reason, Created: manager.clock(), Values: cloneValues(values)}
	manager.snapshots = append(manager.snapshots, snapshot)
	if len(manager.snapshots) > manager.limit {
		manager.snapshots = append([]Snapshot(nil), manager.snapshots[len(manager.snapshots)-manager.limit:]...)
	}
	return cloneSnapshot(snapshot)
}

func (manager *Manager) Get(id uint64) (Snapshot, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	for _, snapshot := range manager.snapshots {
		if snapshot.ID == id {
			return cloneSnapshot(snapshot), true
		}
	}
	return Snapshot{}, false
}

func (manager *Manager) Latest() (Snapshot, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if len(manager.snapshots) == 0 {
		return Snapshot{}, false
	}
	return cloneSnapshot(manager.snapshots[len(manager.snapshots)-1]), true
}

func (manager *Manager) Restore(id uint64) ([]model.EventRecord, error) {
	snapshot, ok := manager.Get(id)
	if !ok {
		return nil, fmt.Errorf("snapshot %d does not exist", id)
	}
	return cloneValues(snapshot.Values), nil
}

func (manager *Manager) List() []Snapshot {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	result := make([]Snapshot, len(manager.snapshots))
	for index, snapshot := range manager.snapshots {
		result[index] = cloneSnapshot(snapshot)
	}
	return result
}

func cloneValues(values []model.EventRecord) []model.EventRecord {
	result := make([]model.EventRecord, len(values))
	for index, value := range values {
		result[index] = value.Clone()
	}
	return result
}

func cloneSnapshot(value Snapshot) Snapshot {
	value.Values = cloneValues(value.Values)
	return value
}
