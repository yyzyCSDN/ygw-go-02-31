package model

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// EventRecord is the canonical record managed by Event Ledger.
type EventRecord struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Priority  int               `json:"priority"`
	Amount    int64             `json:"amount"`
	Active    bool              `json:"active"`
	Version   uint64            `json:"version"`
	Tags      []string          `json:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func (value EventRecord) Clone() EventRecord {
	value.Tags = append([]string(nil), value.Tags...)
	value.Metadata = cloneMetadata(value.Metadata)
	return value
}

func (value EventRecord) Validate() error {
	if strings.TrimSpace(value.ID) == "" {
		return fmt.Errorf("eventRecord id is empty")
	}
	if strings.TrimSpace(value.Name) == "" {
		return fmt.Errorf("eventRecord name is empty")
	}
	if value.Priority < 0 {
		return fmt.Errorf("eventRecord priority cannot be negative")
	}
	if value.Amount < 0 {
		return fmt.Errorf("eventRecord amount cannot be negative")
	}
	for key := range value.Metadata {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("eventRecord metadata contains a blank key")
		}
	}
	return nil
}

func (value EventRecord) HasTag(tag string) bool {
	tag = normalizeToken(tag)
	for _, candidate := range value.Tags {
		if normalizeToken(candidate) == tag {
			return true
		}
	}
	return false
}

func (value EventRecord) MetadataValue(key string) (string, bool) {
	result, ok := value.Metadata[key]
	return result, ok
}

func (value EventRecord) WithTag(tag string) EventRecord {
	copy := value.Clone()
	tag = normalizeToken(tag)
	if tag == "" || copy.HasTag(tag) {
		return copy
	}
	copy.Tags = append(copy.Tags, tag)
	sort.Strings(copy.Tags)
	return copy
}

func (value EventRecord) WithoutTag(tag string) EventRecord {
	copy := value.Clone()
	tag = normalizeToken(tag)
	filtered := copy.Tags[:0]
	for _, candidate := range copy.Tags {
		if normalizeToken(candidate) != tag {
			filtered = append(filtered, candidate)
		}
	}
	copy.Tags = append([]string(nil), filtered...)
	return copy
}

func (value EventRecord) WithMetadata(key, content string) EventRecord {
	copy := value.Clone()
	if copy.Metadata == nil {
		copy.Metadata = map[string]string{}
	}
	copy.Metadata[strings.TrimSpace(key)] = content
	return copy
}

func (value EventRecord) Touch(now time.Time) EventRecord {
	copy := value.Clone()
	if copy.CreatedAt.IsZero() {
		copy.CreatedAt = now
	}
	copy.UpdatedAt = now
	copy.Version++
	return copy
}

func NormalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = normalizeToken(tag)
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}

func cloneMetadata(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func normalizeToken(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
