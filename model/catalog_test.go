package model

import "testing"

func TestFieldCatalog(t *testing.T) {
	if len(StandardFields) < 80 {
		t.Fatalf("catalog too small: %d", len(StandardFields))
	}
	first := StandardFields[0]
	if got, ok := FieldByName(first.Name); !ok || got.Label == "" {
		t.Fatalf("FieldByName = %#v, %v", got, ok)
	}
	if _, ok := FieldByName("missing"); ok {
		t.Fatal("missing field found")
	}
	if len(ExportedFieldNames()) != len(StandardFields) {
		t.Fatal("exported names mismatch")
	}
}
