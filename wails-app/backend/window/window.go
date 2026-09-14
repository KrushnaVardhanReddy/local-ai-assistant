package window

import "context"

// WindowModifier interface defines the methods that can be performed on the application window.
type WindowModifier interface {
	SetIgnoreMouseEvents(ctx context.Context, ignore bool) error
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
