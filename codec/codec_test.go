package codec

import (
	"bytes"
	"testing"
	"time"

	"example.com/eventledger/model"
)

func codecValues() []model.EventRecord {
	return []model.EventRecord{{ID: "one", Name: "First", Status: "ready", Priority: 2, Amount: 12, Active: true, Version: 3, Tags: []string{"blue", "green"}, UpdatedAt: time.Unix(10, 0).UTC()}}
}

func TestJSONRoundTrip(t *testing.T) {
	data, err := EncodeJSON(codecValues())
	if err != nil {
		t.Fatal(err)
	}
	values, err := DecodeJSON(bytes.NewReader(data))
	if err != nil || len(values) != 1 || values[0].ID != "one" {
		t.Fatalf("DecodeJSON = %#v, %v", values, err)
	}
	if _, err := DecodeJSON(bytes.NewBufferString("[{\"id\":\"x\",\"name\":\"X\",\"unknown\":1}]")); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestCSVRoundTrip(t *testing.T) {
	data, err := EncodeCSV(codecValues())
	if err != nil {
		t.Fatal(err)
	}
	values, err := DecodeCSV(bytes.NewReader(data))
	if err != nil || len(values) != 1 {
		t.Fatalf("DecodeCSV = %#v, %v", values, err)
	}
	if values[0].Amount != 12 || !values[0].Active || len(values[0].Tags) != 2 {
		t.Fatalf("round trip mismatch: %#v", values[0])
	}
	if _, err := DecodeCSV(bytes.NewBufferString("bad,header\n")); err == nil {
		t.Fatal("expected header error")
	}
}
