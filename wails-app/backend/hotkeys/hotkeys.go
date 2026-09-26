package hotkeys

import (
	"context"
	"fmt"
	"log"

	"golang.design/x/hotkey"
)

// Wails functions mockable for tests
var (
	WindowGetPosition  = func(ctx context.Context) (int, int) { return 0, 0 }
	WindowSetPosition  = func(ctx context.Context, x, y int) {}
	EventsEmit         = func(ctx context.Context, eventName string, optionalData ...interface{}) {}
	ToggleClickthrough = func(ctx context.Context) {}
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

	// Ctrl+Alt+S (Send candidate to LLM)
	hkS := NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, ModAlt}, hotkey.KeyS)
	if err := hkS.Register(); err != nil {
		return fmt.Errorf("failed to register Ctrl+Alt+S: %v", err)
	}

	// Register Stealth Mode hotkey variants (Ctrl+Alt+M & Ctrl+Shift+M, including NumLock variants)
	stealthVariants := []HotkeyInterface{
		NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, ModAlt}, hotkey.KeyM),
		NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyM),
		NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, ModAlt, hotkey.Mod2}, hotkey.KeyM),
		NewHotkey([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift, hotkey.Mod2}, hotkey.KeyM),
	}

	var activeStealthKeys []HotkeyInterface
	for _, hk := range stealthVariants {
		if err := hk.Register(); err == nil {
			activeStealthKeys = append(activeStealthKeys, hk)
		}
	}
	log.Printf("✅ Registered %d stealth hotkey variant(s) for Ctrl+Alt+M / Ctrl+Shift+M\n", len(activeStealthKeys))

	stealthCh := mergedStealthChannel(activeStealthKeys)

	go func() {
		for {
			select {
			case <-ctx.Done():
				hkRight.Unregister()
				hkLeft.Unregister()
				hk1.Unregister()
				hkS.Unregister()
				for _, hk := range activeStealthKeys {
					_ = hk.Unregister()
				}
				return
			case <-hkRight.Keydown():
				handleMoveWindow(ctx, 100)
			case <-hkLeft.Keydown():
				handleMoveWindow(ctx, -100)
			case <-hk1.Keydown():
				handleIPC(ctx)
			case <-hkS.Keydown():
				EventsEmit(ctx, "hotkey_send_candidate_to_llm")
			case <-stealthCh:
				log.Println("🔥 [HOTKEY] Stealth toggle keydown detected! Invoking ToggleClickthrough...")
				ToggleClickthrough(ctx)
				EventsEmit(ctx, "toggle-clickthrough")
			}
		}
	}()
	return nil
}

func mergedStealthChannel(keys []HotkeyInterface) <-chan hotkey.Event {
	out := make(chan hotkey.Event, 5)
	for _, k := range keys {
		if k == nil {
			continue
		}
		go func(hk HotkeyInterface) {
			for ev := range hk.Keydown() {
				out <- ev
			}
		}(k)
	}
	return out
}

func handleMoveWindow(ctx context.Context, deltaX int) {
	x, y := WindowGetPosition(ctx)
	WindowSetPosition(ctx, x+deltaX, y)
}

func handleIPC(ctx context.Context) {
	EventsEmit(ctx, "hotkey_transcript_1")
}
