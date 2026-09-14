//go:build windows

package clickthrough

import (
	"context"
	"testing"
)

func TestSetIgnoreMouseEventsWindows(t *testing.T) {
	ctx := context.Background()
    // FindWindowW will return 0 (window not found) and error safely
    err := SetIgnoreMouseEvents(ctx, true)
	if err != nil && err.Error() != "window not found" {
        t.Errorf("Unexpected error: %v", err)
    }
}
