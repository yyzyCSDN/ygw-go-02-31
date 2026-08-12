package events

import (
	"testing"
)

func TestApplySuccess(t *testing.T) {
	s := State{Values: map[string]int{"a": 1}}
	if err := s.Apply([]Event{{"a", 2}}); err != nil || s.Values["a"] != 3 {
		t.Fatal("apply failed")
	}
}
