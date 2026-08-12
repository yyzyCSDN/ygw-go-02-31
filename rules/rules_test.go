package rules

import (
	"testing"

	"example.com/eventledger/model"
)

func ruleValues() []model.EventRecord {
	return []model.EventRecord{
		{ID: "one", Name: "One", Status: "ready", Active: true, Amount: 20, Priority: 2, Tags: []string{"blue"}},
		{ID: "two", Name: "Two", Status: "new", Active: false, Amount: 5, Priority: 8, Tags: []string{"green"}},
	}
}

func TestPredicateCombinators(t *testing.T) {
	value := ruleValues()[0]
	if !(All{Active(), Status("ready"), Tagged("blue")}).Match(value) {
		t.Fatal("all predicate did not match")
	}
	if (Any{Status("new"), MinimumAmount(10)}).Match(value) != true {
		t.Fatal("any predicate did not match")
	}
	if (Not{Predicate: Active()}).Match(value) {
		t.Fatal("not predicate matched")
	}
	if !PriorityBetween(1, 3).Match(value) {
		t.Fatal("priority predicate did not match")
	}
}

func TestRuleSetEvaluationAndSelection(t *testing.T) {
	set, err := New([]Rule{{Name: "enabled", Predicate: Active()}, {Name: "large", Predicate: MinimumAmount(10)}})
	if err != nil {
		t.Fatal(err)
	}
	if names := set.Names(); len(names) != 2 || names[0] != "enabled" {
		t.Fatalf("unexpected names: %v", names)
	}
	evaluation := set.Evaluate(ruleValues()[0])
	if len(evaluation.Matched) != 2 || len(evaluation.Missed) != 0 {
		t.Fatalf("unexpected evaluation: %#v", evaluation)
	}
	selected, err := set.Select(ruleValues(), "enabled")
	if err != nil || len(selected) != 1 || selected[0].ID != "one" {
		t.Fatalf("unexpected selection: %#v, %v", selected, err)
	}
	selected[0].Name = "changed"
	if ruleValues()[0].Name != "One" {
		t.Fatal("selection did not clone values")
	}
}

func TestRuleSetValidation(t *testing.T) {
	if _, err := New([]Rule{{Name: "", Predicate: Active()}}); err == nil {
		t.Fatal("expected blank name error")
	}
	if _, err := New([]Rule{{Name: "same", Predicate: Active()}, {Name: "SAME", Predicate: Active()}}); err == nil {
		t.Fatal("expected duplicate name error")
	}
	set, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := set.Select(nil, "missing"); err == nil {
		t.Fatal("expected missing rule error")
	}
}
