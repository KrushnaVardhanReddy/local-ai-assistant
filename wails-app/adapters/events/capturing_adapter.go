package eventsadapter

import (
	"wails-app/core/ports/driven"
)

// EmittedEvent records a single event emission for test assertions.
type EmittedEvent struct {
	Name    string
	Payload any
}

// CapturingEventAdapter records all emitted events in order.
// Use in tests where you need to assert what events were fired and in what order.
type CapturingEventAdapter struct {
	Events []EmittedEvent
}

func (c *CapturingEventAdapter) Emit(event string, payload any) {
	c.Events = append(c.Events, EmittedEvent{Name: event, Payload: payload})
}

// HasEvent returns true if an event with the given name was emitted at least once.
func (c *CapturingEventAdapter) HasEvent(name string) bool {
	for _, e := range c.Events {
		if e.Name == name {
			return true
		}
	}
	return false
}

// Clear resets the captured event list.
func (c *CapturingEventAdapter) Clear() {
	c.Events = nil
}

var _ driven.EventPort = (*CapturingEventAdapter)(nil)
