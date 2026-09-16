//go:build !test

package eventsadapter

import (
	"context"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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
	wailsruntime.EventsEmit(w.ctx, event, payload)
}

var _ driven.EventPort = (*WailsEventAdapter)(nil)
