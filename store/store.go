package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"example.com/eventledger/model"
)

type Change struct {
	Revision uint64
	Action   string
	ID       string
	At       time.Time
}

type Store struct {
	mu       sync.RWMutex
	clock    func() time.Time
	revision uint64
	items    map[string]model.EventRecord
	history  []Change
	limit    int
}

func New(historyLimit int) *Store {
	if historyLimit < 1 {
		historyLimit = 1
	}
	return &Store{clock: time.Now, items: map[string]model.EventRecord{}, limit: historyLimit}
}

func (store *Store) Put(value model.EventRecord) (model.EventRecord, error) {
	if err := value.Validate(); err != nil {
		return model.EventRecord{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	current, exists := store.items[value.ID]
	if exists {
		value.CreatedAt = current.CreatedAt
		value.Version = current.Version
	}
	value = value.Touch(store.clock())
	store.items[value.ID] = value.Clone()
	action := "created"
	if exists {
		action = "updated"
	}
	store.recordLocked(action, value.ID)
	return value.Clone(), nil
}

func (store *Store) Create(value model.EventRecord) (model.EventRecord, error) {
	store.mu.RLock()
	_, exists := store.items[value.ID]
	store.mu.RUnlock()
	if exists {
		return model.EventRecord{}, fmt.Errorf("eventRecord %q already exists", value.ID)
	}
	return store.Put(value)
}

func (store *Store) Get(id string) (model.EventRecord, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, exists := store.items[id]
	return value.Clone(), exists
}

func (store *Store) Delete(id string) (model.EventRecord, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	value, exists := store.items[id]
	if !exists {
		return model.EventRecord{}, fmt.Errorf("eventRecord %q does not exist", id)
	}
	delete(store.items, id)
	store.recordLocked("deleted", id)
	return value.Clone(), nil
}

func (store *Store) List() []model.EventRecord {
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]model.EventRecord, 0, len(store.items))
	for _, value := range store.items {
		result = append(result, value.Clone())
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (store *Store) Count() int {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return len(store.items)
}

func (store *Store) Revision() uint64 {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.revision
}

func (store *Store) History() []Change {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return append([]Change(nil), store.history...)
}

func (store *Store) FindByName(text string) []model.EventRecord {
	text = strings.ToLower(strings.TrimSpace(text))
	result := make([]model.EventRecord, 0)
	for _, value := range store.List() {
		if strings.Contains(strings.ToLower(value.Name), text) {
			result = append(result, value)
		}
	}
	return result
}

func (store *Store) ReplaceAll(values []model.EventRecord) error {
	replacement := make(map[string]model.EventRecord, len(values))
	now := store.clock()
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
		if _, duplicate := replacement[value.ID]; duplicate {
			return fmt.Errorf("duplicate eventRecord id %q", value.ID)
		}
		value = value.Touch(now)
		replacement[value.ID] = value.Clone()
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	store.items = replacement
	store.recordLocked("replaced", "*")
	return nil
}

func (store *Store) recordLocked(action, id string) {
	store.revision++
	store.history = append(store.history, Change{Revision: store.revision, Action: action, ID: id, At: store.clock()})
	if len(store.history) > store.limit {
		store.history = append([]Change(nil), store.history[len(store.history)-store.limit:]...)
	}
}
