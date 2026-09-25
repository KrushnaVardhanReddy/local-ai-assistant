package windowadapter

import (
	"context"

	"wails-app/backend/window"
	"wails-app/core/ports/driven"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsWindowAdapter implements driven.WindowPort using Wails v2 runtime.
type WailsWindowAdapter struct{}

// NewWailsWindowAdapter creates a new WailsWindowAdapter.
func NewWailsWindowAdapter() *WailsWindowAdapter {
	return &WailsWindowAdapter{}
}

func (w *WailsWindowAdapter) SetCaptureExcluded(ctx context.Context, excluded bool) error {
	return window.SetCaptureExcluded(ctx, excluded)
}

func (w *WailsWindowAdapter) HideFromTaskbar(ctx context.Context) error {
	runtime.WindowSetAlwaysOnTop(ctx, true)
	return window.HideFromTaskbar(ctx)
}

var _ driven.WindowPort = (*WailsWindowAdapter)(nil)
