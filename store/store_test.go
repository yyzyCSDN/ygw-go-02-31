package store

import (
	"testing"

	"example.com/eventledger/model"
)

func sample(id, name string) model.EventRecord {
	return model.EventRecord{ID: id, Name: name, Active: true, Tags: []string{"sample"}}
}

func TestStoreLifecycle(t *testing.T) {
	values := New(4)
	created, err := values.Create(sample("b", "Beta"))
	if err != nil || created.Version != 1 {
		t.Fatalf("create = %#v, %v", created, err)
	}
	if _, err := values.Create(sample("b", "Again")); err == nil {
		t.Fatal("expected duplicate error")
	}
	if _, err := values.Put(sample("a", "Alpha")); err != nil {
		t.Fatal(err)
	}
	listed := values.List()
	if len(listed) != 2 || listed[0].ID != "a" || listed[1].ID != "b" {
		t.Fatalf("unexpected list: %#v", listed)
	}
	listed[0].Name = "changed"
	stored, _ := values.Get("a")
	if stored.Name != "Alpha" {
		t.Fatal("List returned aliased values")
	}
	if _, err := values.Delete("missing"); err == nil {
		t.Fatal("expected missing delete error")
	}
	if _, err := values.Delete("a"); err != nil || values.Count() != 1 {
		t.Fatalf("delete failed: %v", err)
	}
}

func TestStoreReplaceAllIsAtomic(t *testing.T) {
	values := New(3)
	_, _ = values.Put(sample("old", "Old"))
	err := values.ReplaceAll([]model.EventRecord{sample("same", "One"), sample("same", "Two")})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if _, exists := values.Get("old"); !exists {
		t.Fatal("failed replacement changed store")
	}
	if err := values.ReplaceAll([]model.EventRecord{sample("new", "New")}); err != nil {
		t.Fatal(err)
	}
	if values.Count() != 1 || len(values.History()) != 2 {
		t.Fatalf("unexpected store state: %d %#v", values.Count(), values.History())
	}
}

func TestStoreFindAndHistoryLimit(t *testing.T) {
	values := New(2)
	_, _ = values.Put(sample("a", "North Alpha"))
	_, _ = values.Put(sample("b", "South Beta"))
	_, _ = values.Put(sample("c", "North Gamma"))
	if got := values.FindByName("north"); len(got) != 2 {
		t.Fatalf("FindByName returned %d", len(got))
	}
	history := values.History()
	if len(history) != 2 || history[0].Revision != 2 || history[1].Revision != 3 {
		t.Fatalf("unexpected history: %#v", history)
	}
}
