//go:build windows && !test

package window

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	findWindowW = user32.NewProc("FindWindowW")
	setWindowLong = user32.NewProc("SetWindowLongW")
	getWindowLong = user32.NewProc("GetWindowLongW")
)

const (
	GWL_EXSTYLE = -20
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_LAYERED = 0x00080000
)

type windowsModifier struct{}

func init() {
	defaultModifier = &windowsModifier{}
}

func (w *windowsModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	// Wails currently does not expose the HWND directly in a cross-platform way.
	// Since we know the window title from wails options, we can find it.
	// We'll hardcode "Local AI Assistant" as a fallback, but ideally it should be dynamic.
	titlePtr, _ := syscall.UTF16PtrFromString("Local AI Assistant")

	hwnd, _, _ := findWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return fmt.Errorf("could not find window by title")
	}

	exStyle, _, _ := getWindowLong.Call(hwnd, ^uintptr(0)-uintptr(19)) // -20 in 2's complement

	if ignore {
		exStyle |= WS_EX_TRANSPARENT | WS_EX_LAYERED
	} else {
		exStyle &^= WS_EX_TRANSPARENT
	}

	setWindowLong.Call(hwnd, ^uintptr(0)-uintptr(19), exStyle)

	return nil
}
