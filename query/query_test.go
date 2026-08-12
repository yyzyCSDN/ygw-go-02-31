package query

import (
	"testing"

	"example.com/eventledger/model"
)

func fixtures() []model.EventRecord {
	return []model.EventRecord{
		{ID: "c", Name: "Gamma", Status: "ready", Priority: 2, Amount: 30, Active: true, Tags: []string{"blue"}},
		{ID: "a", Name: "Alpha", Status: "ready", Priority: 3, Amount: 10, Active: false, Tags: []string{"green"}},
		{ID: "b", Name: "Beta", Status: "new", Priority: 1, Amount: 20, Active: true, Tags: []string{"blue"}},
	}
}

func TestSelectSortPage(t *testing.T) {
	active := true
	selected := Select(fixtures(), Filter{Tag: "blue", Active: &active})
	if len(selected) != 2 {
		t.Fatalf("Select returned %d", len(selected))
	}
	sorted, err := Sort(fixtures(), SortByPriority, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := IDs(sorted); got[0] != "a" || got[2] != "b" {
		t.Fatalf("unexpected order: %v", got)
	}
	page, err := Page(sorted, 1, 5)
	if err != nil || len(page) != 2 {
		t.Fatalf("Page = %#v, %v", page, err)
	}
	page[0].Name = "changed"
	if sorted[1].Name == "changed" {
		t.Fatal("Page returned alias")
	}
}

func TestQueryValidationAndGrouping(t *testing.T) {
	if _, err := Sort(fixtures(), "missing", false); err == nil {
		t.Fatal("expected sort error")
	}
	if _, err := Page(fixtures(), -1, 2); err == nil {
		t.Fatal("expected page error")
	}
	groups := GroupByStatus(fixtures())
	if len(groups["ready"]) != 2 || len(groups["new"]) != 1 {
		t.Fatalf("unexpected groups: %#v", groups)
	}
}
