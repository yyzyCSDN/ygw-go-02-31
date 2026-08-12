package audit

import "testing"

func TestLogQueriesAndLimit(t *testing.T) {
	log := New(2)
	log.Append("create", "one", "alice", "")
	log.Append("update", "one", "bob", "")
	log.Append("update", "two", "alice", "")
	entries := log.Entries()
	if len(entries) != 2 || entries[0].Sequence != 2 {
		t.Fatalf("unexpected entries: %#v", entries)
	}
	if len(log.ByAction("update")) != 2 || len(log.Since(2)) != 1 {
		t.Fatal("query mismatch")
	}
	counts := CountByAction(entries)
	actors := Actors(entries)
	if counts["update"] != 2 || len(actors) != 2 || actors[0] != "alice" {
		t.Fatalf("summary mismatch: %#v %v", counts, actors)
	}
}
