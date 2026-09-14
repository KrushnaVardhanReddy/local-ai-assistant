package window

import (
	"context"
	"testing"
)

func TestSetIgnoreMouseEvents(t *testing.T) {
	// First test with nil modifier to ensure it returns nil without error
	defaultModifier = nil
	err := SetIgnoreMouseEvents(context.Background(), true)
	if err != nil {
		t.Errorf("Expected nil error when defaultModifier is nil, got %v", err)
	}

	// Inject mock
	defaultModifier = &mockModifier{}

	err = SetIgnoreMouseEvents(context.Background(), true)
	if err != nil {
		t.Errorf("Expected nil error from mock modifier, got %v", err)
	}

	err = SetIgnoreMouseEvents(context.Background(), false)
	if err != nil {
		t.Errorf("Expected nil error from mock modifier, got %v", err)
	}
}
