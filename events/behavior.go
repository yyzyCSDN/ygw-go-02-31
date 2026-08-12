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
	for _, event := range events {
		s.Values[event.Key] += event.Delta
		if s.Values[event.Key] < 0 {
			return fmt.Errorf("negative state")
		}
	}
	return nil
}
