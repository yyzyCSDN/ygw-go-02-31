package timeline

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Change struct {
	Sequence uint64
	Subject  string
	Kind     string
	Actor    string
	At       time.Time
	Fields   map[string]string
}

func (change Change) Clone() Change {
	change.Fields = cloneFields(change.Fields)
	return change
}

func (change Change) Validate() error {
	if strings.TrimSpace(change.Subject) == "" {
		return fmt.Errorf("change subject is empty")
	}
	if strings.TrimSpace(change.Kind) == "" {
		return fmt.Errorf("change kind is empty")
	}
	if change.At.IsZero() {
		return fmt.Errorf("change timestamp is empty")
	}
	return nil
}

type Journal struct {
	mu      sync.RWMutex
	limit   int
	next    uint64
	changes []Change
}

func New(limit int) *Journal {
	if limit < 1 {
		limit = 1
	}
	return &Journal{limit: limit, next: 1}
}

func (journal *Journal) Append(change Change) (Change, error) {
	if err := change.Validate(); err != nil {
		return Change{}, err
	}
	journal.mu.Lock()
	defer journal.mu.Unlock()
	change = change.Clone()
	change.Sequence = journal.next
	journal.next++
	journal.changes = append(journal.changes, change)
	if overflow := len(journal.changes) - journal.limit; overflow > 0 {
		journal.changes = append([]Change(nil), journal.changes[overflow:]...)
	}
	return change.Clone(), nil
}

func (journal *Journal) List() []Change {
	journal.mu.RLock()
	defer journal.mu.RUnlock()
	return cloneChanges(journal.changes)
}

func (journal *Journal) Subject(subject string) []Change {
	journal.mu.RLock()
	defer journal.mu.RUnlock()
	result := make([]Change, 0)
	for _, change := range journal.changes {
		if change.Subject == subject {
			result = append(result, change.Clone())
		}
	}
	return result
}

func (journal *Journal) Since(cursor uint64) []Change {
	journal.mu.RLock()
	defer journal.mu.RUnlock()
	result := make([]Change, 0)
	for _, change := range journal.changes {
		if change.Sequence > cursor {
			result = append(result, change.Clone())
		}
	}
	return result
}

type Digest struct {
	Total    int
	First    time.Time
	Last     time.Time
	ByKind   map[string]int
	ByActor  map[string]int
	Subjects []string
}

func Summarize(changes []Change) Digest {
	result := Digest{ByKind: map[string]int{}, ByActor: map[string]int{}}
	subjects := map[string]struct{}{}
	for index, change := range changes {
		result.Total++
		result.ByKind[change.Kind]++
		result.ByActor[change.Actor]++
		subjects[change.Subject] = struct{}{}
		if index == 0 || change.At.Before(result.First) {
			result.First = change.At
		}
		if index == 0 || change.At.After(result.Last) {
			result.Last = change.At
		}
	}
	for subject := range subjects {
		result.Subjects = append(result.Subjects, subject)
	}
	sort.Strings(result.Subjects)
	return result
}

func Merge(groups ...[]Change) []Change {
	result := make([]Change, 0)
	for _, group := range groups {
		result = append(result, cloneChanges(group)...)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].At.Equal(result[j].At) {
			return result[i].Sequence < result[j].Sequence
		}
		return result[i].At.Before(result[j].At)
	})
	return result
}

func cloneChanges(input []Change) []Change {
	result := make([]Change, len(input))
	for index, change := range input {
		result[index] = change.Clone()
	}
	return result
}

func cloneFields(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}
