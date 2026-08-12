package domain

import (
	"encoding/json"
	"testing"
)

func TestEventBookLifecycle(t *testing.T) {
	engine := NewEventBook()
	if err := engine.OpenStream(); err != nil {
		t.Fatal(err)
	}
	if err := engine.RegisterEventRecord(EventBookRecord{ID: "primary", Quantity: 4, Labels: map[string]string{"zone": "a"}}); err != nil {
		t.Fatal(err)
	}
	if err := engine.ApplyEventRecord("primary", 3); err != nil {
		t.Fatal(err)
	}
	if err := engine.RollbackEventRecord("primary", 2); err != nil {
		t.Fatal(err)
	}
	if got := engine.CountCommitted(); got != 5 {
		t.Fatalf("count = %d; want 5", got)
	}
	if err := engine.CloseStream(); err != nil {
		t.Fatal(err)
	}
}

func TestEventBookPrioritiesAndExport(t *testing.T) {
	engine := NewEventBook()
	_ = engine.RegisterEventRecord(EventBookRecord{ID: "low", Quantity: 1})
	_ = engine.RegisterEventRecord(EventBookRecord{ID: "high", Quantity: 2})
	if err := engine.PrioritizeEventRecord("high", 9); err != nil {
		t.Fatal(err)
	}
	values := engine.List()
	if len(values) != 2 || values[0].ID != "high" {
		t.Fatalf("unexpected order: %#v", values)
	}
	values[0].Labels = map[string]string{"changed": "yes"}
	data, err := engine.ExportEvents()
	if err != nil || !json.Valid(data) {
		t.Fatalf("invalid export: %s, %v", data, err)
	}
}

func TestEventBookRejectsInvalidOperations(t *testing.T) {
	engine := NewEventBook()
	if err := engine.RegisterEventRecord(EventBookRecord{}); err == nil {
		t.Fatal("expected blank id error")
	}
	if err := engine.RegisterEventRecord(EventBookRecord{ID: "one", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	if err := engine.RegisterEventRecord(EventBookRecord{ID: "one"}); err == nil {
		t.Fatal("expected duplicate error")
	}
	if err := engine.RollbackEventRecord("one", 2); err == nil {
		t.Fatal("expected insufficient quantity error")
	}
	if err := engine.PrioritizeEventRecord("missing", 1); err == nil {
		t.Fatal("expected missing record error")
	}
}
