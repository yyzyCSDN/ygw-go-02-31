package timeline

import (
	"testing"
	"time"
)

func TestJournalRetentionAndQueries(t *testing.T) {
	journal := New(2)
	base := time.Unix(100, 0).UTC()
	for index, subject := range []string{"one", "two", "one"} {
		_, err := journal.Append(Change{Subject: subject, Kind: "save", Actor: "worker", At: base.Add(time.Duration(index) * time.Second), Fields: map[string]string{"name": subject}})
		if err != nil {
			t.Fatal(err)
		}
	}
	if values := journal.List(); len(values) != 2 || values[0].Sequence != 2 {
		t.Fatalf("unexpected retained values: %#v", values)
	}
	if values := journal.Subject("one"); len(values) != 1 || values[0].Sequence != 3 {
		t.Fatalf("unexpected subject values: %#v", values)
	}
	if values := journal.Since(2); len(values) != 1 || values[0].Sequence != 3 {
		t.Fatalf("unexpected cursor values: %#v", values)
	}
}

func TestJournalClonesFields(t *testing.T) {
	journal := New(5)
	fields := map[string]string{"state": "new"}
	stored, err := journal.Append(Change{Subject: "one", Kind: "save", At: time.Now(), Fields: fields})
	if err != nil {
		t.Fatal(err)
	}
	fields["state"] = "changed"
	stored.Fields["state"] = "also changed"
	if got := journal.List()[0].Fields["state"]; got != "new" {
		t.Fatalf("stored fields changed: %q", got)
	}
}

func TestSummarizeAndMerge(t *testing.T) {
	base := time.Unix(100, 0).UTC()
	left := []Change{{Sequence: 2, Subject: "b", Kind: "update", Actor: "sam", At: base.Add(time.Second)}}
	right := []Change{{Sequence: 1, Subject: "a", Kind: "create", Actor: "lee", At: base}}
	merged := Merge(left, right)
	if len(merged) != 2 || merged[0].Subject != "a" {
		t.Fatalf("unexpected merge: %#v", merged)
	}
	digest := Summarize(merged)
	if digest.Total != 2 || len(digest.Subjects) != 2 || digest.ByKind["create"] != 1 {
		t.Fatalf("unexpected digest: %#v", digest)
	}
}

func TestChangeValidation(t *testing.T) {
	if _, err := New(1).Append(Change{}); err == nil {
		t.Fatal("expected validation error")
	}
}
