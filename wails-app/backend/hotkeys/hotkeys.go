package hotkeys

import (
	"context"
	"fmt"

	"golang.design/x/hotkey"
)

// Wails functions mockable for tests
var (
	WindowGetPosition = func(ctx context.Context) (int, int) { return 0, 0 }
	WindowSetPosition = func(ctx context.Context, x, y int) {}
	EventsEmit        = func(ctx context.Context, eventName string, optionalData ...interface{}) {}
)

// Mockable functions to abstract the third-party hardware listener
var (
	NewHotkey = func(mods []hotkey.Modifier, key hotkey.Key) HotkeyInterface {
		return &realHotkey{hk: hotkey.New(mods, key)}
	}
)

type HotkeyInterface interface {
	Register() error
	Unregister() error
	Keydown() <-chan hotkey.Event
}

type realHotkey struct {
	hk *hotkey.Hotkey
}

func (r *realHotkey) Register() error {
	return r.hk.Register()
}

func (r *realHotkey) Unregister() error {
	return r.hk.Unregister()
}

func (r *realHotkey) Keydown() <-chan hotkey.Event {
	return r.hk.Keydown()
}

// Start registers global hotkeys and listens for them.
// Note: hotkey package functions can fail if no display server is present (e.g., in CI).
// We return an error so tests/callers can handle it.
func Start(ctx context.Context) error {
	// Ctrl+Alt+Right
	hkRight := NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, ModAlt}, hotkey.KeyRight)
	if err := hkRight.Register(); err != nil {
		return fmt.Errorf("failed to register Ctrl+Alt+Right: %v", err)
	}

	// Ctrl+Alt+Left
	hkLeft := NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, ModAlt}, hotkey.KeyLeft)
	if err := hkLeft.Register(); err != nil {
		return fmt.Errorf("failed to register Ctrl+Alt+Left: %v", err)
	}

	// Ctrl+Alt+1
	hk1 := NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, ModAlt}, hotkey.Key1)
	if err := hk1.Register(); err != nil {
		return fmt.Errorf("failed to register Ctrl+Alt+1: %v", err)
	}

	// Ctrl+Alt+F8 (Toggle Stealth / Click-through Mode)
	hkStealth := NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, ModAlt}, hotkey.KeyF8)
	if err := hkStealth.Register(); err != nil {
		// Log error but non-fatal if hotkey binding conflicts
		fmt.Printf("Warning: failed to register Ctrl+Alt+F8: %v\n", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				hkRight.Unregister()
				hkLeft.Unregister()
				hk1.Unregister()
				if hkStealth != nil {
					_ = hkStealth.Unregister()
				}
				return
			case <-hkRight.Keydown():
				handleMoveWindow(ctx, 100)
			case <-hkLeft.Keydown():
				handleMoveWindow(ctx, -100)
			case <-hk1.Keydown():
				handleIPC(ctx)
			case <-hkStealth.Keydown():
				EventsEmit(ctx, "toggle-clickthrough")
			}
		}
	}()
	return nil
}

func handleMoveWindow(ctx context.Context, deltaX int) {
	x, y := WindowGetPosition(ctx)
	WindowSetPosition(ctx, x+deltaX, y)
}

func handleIPC(ctx context.Context) {
	EventsEmit(ctx, "hotkey_transcript_1")
}
