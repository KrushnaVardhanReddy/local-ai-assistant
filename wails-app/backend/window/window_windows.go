//go:build windows && !test

package window

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	findWindowW   = user32.NewProc("FindWindowW")
	setWindowLong = user32.NewProc("SetWindowLongW")
	getWindowLong = user32.NewProc("GetWindowLongW")
)

const (
	GWL_EXSTYLE       = -20
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_LAYERED     = 0x00080000
	WS_EX_TOOLWINDOW  = 0x00000080
)

type windowsModifier struct{}

func init() {
	defaultModifier = &windowsModifier{}
}

func (w *windowsModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	// Wails currently does not expose the HWND directly in a cross-platform way.
	// Since we know the window title from wails options, we can find it.
	// We'll hardcode "BarnOwl AI" as a fallback, but ideally it should be dynamic.
	titlePtr, _ := syscall.UTF16PtrFromString("BarnOwl AI")

	hwnd, _, _ := findWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return fmt.Errorf("could not find window by title")
	}

	exStyle, _, _ := getWindowLong.Call(hwnd, uintptr(GWL_EXSTYLE&0xFFFFFFFF))

	if ignore {
		exStyle |= WS_EX_TRANSPARENT | WS_EX_LAYERED
	} else {
		exStyle &^= WS_EX_TRANSPARENT
	}

	setWindowLong.Call(hwnd, uintptr(GWL_EXSTYLE&0xFFFFFFFF), exStyle)

	return nil
}

func (w *windowsModifier) HideFromTaskbar(ctx context.Context) error {
	titlePtr, _ := syscall.UTF16PtrFromString("BarnOwl AI")

	hwnd, _, _ := findWindowW.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd == 0 {
		return fmt.Errorf("could not find window by title")
	}

	exStyle, _, _ := getWindowLong.Call(hwnd, uintptr(GWL_EXSTYLE&0xFFFFFFFF))

	exStyle |= WS_EX_TOOLWINDOW

	setWindowLong.Call(hwnd, uintptr(GWL_EXSTYLE&0xFFFFFFFF), exStyle)

	return nil
}
