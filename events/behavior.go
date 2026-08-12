package events

import (
	"fmt"
)

type State struct{ Values map[string]int }
type Event struct {
	Key   string
	Delta int
}

func (s *State) Apply(events []Event) error {
	next := make(map[string]int, len(s.Values))
	for key, value := range s.Values {
		next[key] = value
	}
	for _, event := range events {
		next[event.Key] += event.Delta
		if next[event.Key] < 0 {
			return fmt.Errorf("negative state")
		}
	}
	s.Values = next
	return nil
}
