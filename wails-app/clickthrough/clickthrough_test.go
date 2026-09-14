package clickthrough

import (
	"context"
	"testing"
)

// The actual logic is tested in platform-specific test files
// (clickthrough_linux_test.go, clickthrough_windows_test.go, clickthrough_darwin_test.go).
// This test ensures `clickthrough_test.go` always exists and runs at least a dummy.
func TestSetIgnoreMouseEventsDummy(t *testing.T) {
	ctx := context.Background()
    // Test the public wrapper using the assigned implementation from init()
    _ = SetIgnoreMouseEvents(ctx, true)
}
