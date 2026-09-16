package eventsadapter

import (
	"wails-app/core/ports/driven"
)

// NoopEventAdapter silently discards all emitted events.
// Use in tests where you don't need to assert on events.
type NoopEventAdapter struct{}

func (n *NoopEventAdapter) Emit(_ string, _ any) {}

var _ driven.EventPort = (*NoopEventAdapter)(nil)
