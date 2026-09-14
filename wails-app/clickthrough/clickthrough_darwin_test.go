//go:build darwin

package clickthrough

import (
	"context"
	"testing"
    "os"
)

func TestSetIgnoreMouseEventsDarwin(t *testing.T) {
	ctx := context.Background()
    // Skip if CI environment variable is set
    if os.Getenv("CI") == "" {
        _ = SetIgnoreMouseEvents(ctx, true)
        _ = SetIgnoreMouseEvents(ctx, false)
    } else {
        t.Skip("Skipping OS-specific CGO invocation to avoid CI crashes")
    }
}
