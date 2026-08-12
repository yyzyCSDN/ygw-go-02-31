package events

import (
	"testing"
)

func TestRegressionBehavior(t *testing.T) {
	s := State{Values: map[string]int{"a": 1, "b": 0}}
	err := s.Apply([]Event{{"a", 2}, {"b", -1}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if s.Values["a"] != 1 || s.Values["b"] != 0 {
		t.Fatalf("partial state committed: %v", s.Values)
	}
}
