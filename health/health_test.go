package health

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMonitorRunsAndStoresChecks(t *testing.T) {
	monitor := New()
	if err := monitor.Register(CheckFunc{CheckName: "store", Function: func(context.Context) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	if err := monitor.Register(CheckFunc{CheckName: "queue", Function: func(context.Context) error { return fmt.Errorf("offline") }}); err != nil {
		t.Fatal(err)
	}
	results := monitor.RunAll(context.Background())
	if len(results) != 2 || results[0].Name != "queue" || results[1].Name != "store" {
		t.Fatalf("unexpected results: %#v", results)
	}
	summary := Summarize(results)
	if summary.Passing != 1 || summary.Failing != 1 || summary.Healthy() {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if last, ok := monitor.Last("queue"); !ok || last.Message != "offline" {
		t.Fatalf("unexpected last result: %#v, %v", last, ok)
	}
}

func TestMonitorValidation(t *testing.T) {
	monitor := New()
	if err := monitor.Register(nil); err == nil {
		t.Fatal("expected nil check error")
	}
	check := CheckFunc{CheckName: "store", Function: func(context.Context) error { return nil }}
	if err := monitor.Register(check); err != nil {
		t.Fatal(err)
	}
	if err := monitor.Register(check); err == nil {
		t.Fatal("expected duplicate error")
	}
	if _, err := monitor.Run(context.Background(), "missing"); err == nil {
		t.Fatal("expected missing check error")
	}
}

func TestMonitorDuration(t *testing.T) {
	monitor := New()
	times := []time.Time{time.Unix(10, 0), time.Unix(12, 0)}
	monitor.clock = func() time.Time {
		value := times[0]
		times = times[1:]
		return value
	}
	_ = monitor.Register(CheckFunc{CheckName: "clock", Function: func(context.Context) error { return nil }})
	result, err := monitor.Run(context.Background(), "clock")
	if err != nil || result.Duration != 2*time.Second {
		t.Fatalf("unexpected duration: %#v, %v", result, err)
	}
}

func TestFilterAndFailureMessages(t *testing.T) {
	results := []Result{
		{Name: "store", Status: Passing},
		{Name: "queue", Status: Failing, Message: "offline"},
		{Name: "cache", Status: Failing, Message: "stale"},
	}
	failures := Filter(results, Failing)
	if len(failures) != 2 || failures[0].Name != "cache" {
		t.Fatalf("unexpected failures: %#v", failures)
	}
	messages := FailureMessages(results)
	if len(messages) != 2 || messages[1] != "queue: offline" {
		t.Fatalf("unexpected messages: %v", messages)
	}
	if Count(results, Failing) != 2 || Count(results, Warning) != 0 {
		t.Fatal("status counts do not match")
	}
}
