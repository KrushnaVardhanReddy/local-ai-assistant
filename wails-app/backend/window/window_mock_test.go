package window

import (
	"context"
)

type mockModifier struct{}


func (m *mockModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	return nil
}
