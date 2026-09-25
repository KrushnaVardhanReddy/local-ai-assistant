package window

import (
	"context"
)

type mockModifier struct{}

func (m *mockModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	return nil
}

func (m *mockModifier) HideFromTaskbar(ctx context.Context) error {
	return nil
}

func (m *mockModifier) SetCaptureExcluded(ctx context.Context, excluded bool) error {
	return nil
}
