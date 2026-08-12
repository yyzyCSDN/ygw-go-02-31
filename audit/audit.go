package audit

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	Sequence uint64
	Action   string
	Subject  string
	Actor    string
	Detail   string
	At       time.Time
}

type Log struct {
	mu      sync.RWMutex
	clock   func() time.Time
	limit   int
	next    uint64
	entries []Entry
}

func New(limit int) *Log {
	if limit < 1 {
		limit = 1
	}
	return &Log{clock: time.Now, limit: limit}
}

func (log *Log) Append(action, subject, actor, detail string) Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.next++
	entry := Entry{Sequence: log.next, Action: strings.TrimSpace(action), Subject: strings.TrimSpace(subject), Actor: strings.TrimSpace(actor), Detail: detail, At: log.clock()}
	log.entries = append(log.entries, entry)
	if len(log.entries) > log.limit {
		log.entries = append([]Entry(nil), log.entries[len(log.entries)-log.limit:]...)
	}
	return entry
}

func (log *Log) Entries() []Entry {
	log.mu.RLock()
	defer log.mu.RUnlock()
	return append([]Entry(nil), log.entries...)
}

func (log *Log) ByAction(action string) []Entry {
	result := make([]Entry, 0)
	for _, entry := range log.Entries() {
		if entry.Action == action {
			result = append(result, entry)
		}
	}
	return result
}

func (log *Log) Since(sequence uint64) []Entry {
	result := make([]Entry, 0)
	for _, entry := range log.Entries() {
		if entry.Sequence > sequence {
			result = append(result, entry)
		}
	}
	return result
}

func CountByAction(entries []Entry) map[string]int {
	result := map[string]int{}
	for _, entry := range entries {
		result[entry.Action]++
	}
	return result
}

func Actors(entries []Entry) []string {
	set := map[string]struct{}{}
	for _, entry := range entries {
		if entry.Actor != "" {
			set[entry.Actor] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for actor := range set {
		result = append(result, actor)
	}
	sort.Strings(result)
	return result
}
