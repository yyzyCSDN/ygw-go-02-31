package model

import (
	"testing"
	"time"
)

func TestEventRecordCloneAndTags(t *testing.T) {
	original := EventRecord{ID: "one", Name: "First", Tags: []string{"Blue"}, Metadata: map[string]string{"zone": "a"}}
	copy := original.WithTag("green").WithMetadata("zone", "b")
	copy.Tags[0] = "changed"
	if original.Tags[0] != "Blue" || original.Metadata["zone"] != "a" {
		t.Fatal("copy operation mutated original")
	}
	if copy.Metadata["zone"] != "b" || !copy.HasTag("green") {
		t.Fatalf("copy missing changes: %#v", copy)
	}
}

func TestEventRecordValidationAndTouch(t *testing.T) {
	if err := (EventRecord{}).Validate(); err == nil {
		t.Fatal("expected validation error")
	}
	now := time.Unix(100, 0).UTC()
	value := (EventRecord{ID: "one", Name: "First"}).Touch(now)
	if value.Version != 1 || !value.CreatedAt.Equal(now) || !value.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected touched value: %#v", value)
	}
}

func TestNormalizeTags(t *testing.T) {
	got := NormalizeTags([]string{" Blue ", "red", "blue", "", "Green"})
	want := []string{"blue", "green", "red"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
