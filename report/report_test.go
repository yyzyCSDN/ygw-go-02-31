package report

import (
	"testing"

	"example.com/eventledger/model"
)

func TestBuildSummary(t *testing.T) {
	summary := Build([]model.EventRecord{
		{ID: "a", Name: "A", Status: "ready", Active: true, Amount: 10, Tags: []string{"blue", "green"}},
		{ID: "b", Name: "B", Status: "ready", Active: false, Amount: 30, Tags: []string{"blue"}},
	})
	if summary.Total != 2 || summary.Active != 1 || summary.Inactive != 1 || summary.TotalAmount != 40 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if AverageAmount(summary) != 20 || AverageAmount(Build(nil)) != 0 {
		t.Fatal("average mismatch")
	}
	if tags := TopTags(summary, 2); len(tags) != 2 || tags[0] != "blue" {
		t.Fatalf("unexpected tags: %v", tags)
	}
	if statuses := Statuses(summary); len(statuses) != 1 || statuses[0] != "ready" {
		t.Fatalf("unexpected statuses: %v", statuses)
	}
}
