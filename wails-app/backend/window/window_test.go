package window

import (
	"context"
	"testing"
)

func TestSetIgnoreMouseEvents(t *testing.T) {
	// First test with nil modifier to ensure it returns nil without error
	originalModifier := defaultModifier
	defaultModifier = nil
	err := SetIgnoreMouseEvents(context.Background(), true)
	if err != nil {
		t.Errorf("Expected nil error when defaultModifier is nil, got %v", err)
	}

	// Restore mock and test again
	defaultModifier = originalModifier
	if defaultModifier == nil {
		t.Skip("defaultModifier is nil, skipping test (ensure mock is registered via build tags)")
	}

	err = SetIgnoreMouseEvents(context.Background(), true)
	if err != nil {
		t.Errorf("Expected nil error from mock modifier, got %v", err)
	}

	err = SetIgnoreMouseEvents(context.Background(), false)
	if err != nil {
		t.Errorf("Expected nil error from mock modifier, got %v", err)
	}
}
