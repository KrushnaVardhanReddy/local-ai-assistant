//go:build test

package window

import (
	"context"
)

type mockModifier struct{}

func init() {
	defaultModifier = &mockModifier{}
}

func (m *mockModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	return nil
}
