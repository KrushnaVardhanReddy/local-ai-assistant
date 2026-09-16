package eventsadapter

import (
	"testing"
)

func TestNoopEventAdapter(t *testing.T) {
	// Setup
	adapter := &NoopEventAdapter{}

	// Action & Assertion
	// Emit should not panic and silently discard the event.
	adapter.Emit("test_event", "payload")
}

func TestCapturingEventAdapter(t *testing.T) {
	t.Run("records events in order", func(t *testing.T) {
		adapter := &CapturingEventAdapter{}

		adapter.Emit("event1", "payload1")
		adapter.Emit("event2", "payload2")

		if len(adapter.Events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(adapter.Events))
		}

		if adapter.Events[0].Name != "event1" || adapter.Events[0].Payload != "payload1" {
			t.Errorf("expected first event to be 'event1' with 'payload1', got '%s' with '%v'", adapter.Events[0].Name, adapter.Events[0].Payload)
		}
		if adapter.Events[1].Name != "event2" || adapter.Events[1].Payload != "payload2" {
			t.Errorf("expected second event to be 'event2' with 'payload2', got '%s' with '%v'", adapter.Events[1].Name, adapter.Events[1].Payload)
		}
	})

	t.Run("has event", func(t *testing.T) {
		adapter := &CapturingEventAdapter{}

		adapter.Emit("event_exist", nil)

		if !adapter.HasEvent("event_exist") {
			t.Errorf("expected HasEvent('event_exist') to be true")
		}

		if adapter.HasEvent("event_missing") {
			t.Errorf("expected HasEvent('event_missing') to be false")
		}
	})

	t.Run("clear resets events", func(t *testing.T) {
		adapter := &CapturingEventAdapter{}

		adapter.Emit("event1", nil)
		adapter.Emit("event2", nil)

		if len(adapter.Events) != 2 {
			t.Fatalf("expected 2 events before clear, got %d", len(adapter.Events))
		}

		adapter.Clear()

		if len(adapter.Events) != 0 {
			t.Errorf("expected 0 events after clear, got %d", len(adapter.Events))
		}
		if adapter.Events != nil {
			t.Errorf("expected events to be nil after clear, got %v", adapter.Events)
		}
	})
}
