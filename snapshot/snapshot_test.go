package snapshot

import (
	"testing"

	"example.com/eventledger/model"
)

func TestSnapshotsAreBoundedAndIndependent(t *testing.T) {
	manager := New(2)
	first := manager.Capture("one", []model.EventRecord{{ID: "one", Name: "First"}})
	manager.Capture("two", nil)
	third := manager.Capture("three", []model.EventRecord{{ID: "three", Name: "Third"}})
	if len(manager.List()) != 2 {
		t.Fatal("snapshot limit not applied")
	}
	if _, ok := manager.Get(first.ID); ok {
		t.Fatal("old snapshot was retained")
	}
	restored, err := manager.Restore(third.ID)
	if err != nil || len(restored) != 1 {
		t.Fatalf("Restore = %#v, %v", restored, err)
	}
	restored[0].Name = "changed"
	latest, _ := manager.Latest()
	if latest.Values[0].Name != "Third" {
		t.Fatal("restore returned aliased values")
	}
	if _, err := manager.Restore(999); err == nil {
		t.Fatal("expected missing error")
	}
}
