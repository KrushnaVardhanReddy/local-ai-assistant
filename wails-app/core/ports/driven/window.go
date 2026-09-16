package driven

import "context"

// WindowPort is the driven port for OS-level window management.
type WindowPort interface {
	// SetCaptureExcluded sets whether this window is excluded from
	// OS screen capture (WDA_EXCLUDEFROMCAPTURE on Windows, XFixes on Linux).
	SetCaptureExcluded(ctx context.Context, excluded bool) error

	// HideFromTaskbar removes this window from the OS taskbar / dock.
	HideFromTaskbar(ctx context.Context) error
}
