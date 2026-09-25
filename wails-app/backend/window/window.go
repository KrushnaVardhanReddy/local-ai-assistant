package window

import "context"

// WindowModifier interface defines the methods that can be performed on the application window.
type WindowModifier interface {
	SetIgnoreMouseEvents(ctx context.Context, ignore bool) error
	HideFromTaskbar(ctx context.Context) error
	SetCaptureExcluded(ctx context.Context, excluded bool) error
}

// defaultModifier is a global variable that holds the current WindowModifier implementation.
// It can be overridden by platform-specific implementations via init() functions,
// or by the mock implementation during tests.
var defaultModifier WindowModifier

// SetIgnoreMouseEvents makes the window click-through (ignoring mouse events) if ignore is true,
// or restores normal mouse interaction if ignore is false.
func SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	if defaultModifier == nil {
		// If no platform specific modifier is set, it's a no-op
		return nil
	}
	return defaultModifier.SetIgnoreMouseEvents(ctx, ignore)
}

// SetCaptureExcluded excludes the window from being captured by screen sharing/recording software.
func SetCaptureExcluded(ctx context.Context, excluded bool) error {
	if defaultModifier == nil {
		return nil
	}
	return defaultModifier.SetCaptureExcluded(ctx, excluded)
}

// HideFromTaskbar hides the application window from the taskbar.
func HideFromTaskbar(ctx context.Context) error {
	if defaultModifier == nil {
		return nil
	}
	return defaultModifier.HideFromTaskbar(ctx)
}
