//go:build linux

package clickthrough

import (
	"context"
	"testing"
    "os"
)

func TestSetIgnoreMouseEventsLinux(t *testing.T) {
	ctx := context.Background()
	// Safely bypassing actual CGO execution in tests to prevent CI crashes.
    // If DISPLAY is not set, we skip execution, which protects headless CI.
    if os.Getenv("DISPLAY") != "" {
        _ = SetIgnoreMouseEvents(ctx, true)
        _ = SetIgnoreMouseEvents(ctx, false)
    } else {
        t.Skip("Skipping OS-specific CGO invocation to avoid CI crashes")
    }
}
