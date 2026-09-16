//go:build test

package eventsadapter

import (
	"context"
	"wails-app/core/ports/driven"
)

// WailsEventAdapter implements driven.EventPort using the Wails runtime.
type WailsEventAdapter struct {
	ctx context.Context
}

func NewWailsEventAdapter(ctx context.Context) *WailsEventAdapter {
	return &WailsEventAdapter{ctx: ctx}
}

// Emit calls wailsruntime.EventsEmit with the event name and payload.
func (w *WailsEventAdapter) Emit(event string, payload any) {
}

var _ driven.EventPort = (*WailsEventAdapter)(nil)
