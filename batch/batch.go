package batch

import (
	"fmt"
	"sort"

	"example.com/eventledger/model"
)

type OperationKind string

const (
	Upsert OperationKind = "upsert"
	Delete OperationKind = "delete"
)

type Operation struct {
	Kind  OperationKind
	ID    string
	Value model.EventRecord
}

type Result struct {
	Values  []model.EventRecord
	Created int
	Updated int
	Deleted int
}

func Apply(current []model.EventRecord, operations []Operation) (Result, error) {
	working := map[string]model.EventRecord{}
	for _, value := range current {
		working[value.ID] = value.Clone()
	}
	result := Result{}
	for index, operation := range operations {
		switch operation.Kind {
		case Upsert:
			if operation.Value.ID == "" {
				operation.Value.ID = operation.ID
			}
			if operation.ID != "" && operation.Value.ID != operation.ID {
				return Result{}, fmt.Errorf("operation %d id mismatch", index+1)
			}
			if err := operation.Value.Validate(); err != nil {
				return Result{}, fmt.Errorf("operation %d: %w", index+1, err)
			}
			if _, exists := working[operation.Value.ID]; exists {
				result.Updated++
			} else {
				result.Created++
			}
			working[operation.Value.ID] = operation.Value.Clone()
		case Delete:
			if _, exists := working[operation.ID]; !exists {
				return Result{}, fmt.Errorf("operation %d: eventRecord %q does not exist", index+1, operation.ID)
			}
			delete(working, operation.ID)
			result.Deleted++
		default:
			return Result{}, fmt.Errorf("operation %d has unknown kind %q", index+1, operation.Kind)
		}
	}
	ids := make([]string, 0, len(working))
	for id := range working {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result.Values = make([]model.EventRecord, 0, len(ids))
	for _, id := range ids {
		result.Values = append(result.Values, working[id].Clone())
	}
	return result, nil
}

func ValidateUnique(values []model.EventRecord) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if _, exists := seen[value.ID]; exists {
			return fmt.Errorf("duplicate eventRecord id %q", value.ID)
		}
		seen[value.ID] = struct{}{}
	}
	return nil
}
