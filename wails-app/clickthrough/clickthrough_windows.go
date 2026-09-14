//go:build windows
// +build windows

package clickthrough

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	findWindow = user32.NewProc("FindWindowW")
	setWindowLong = user32.NewProc("SetWindowLongW")
	getWindowLong = user32.NewProc("GetWindowLongW")
)

const (
	GWL_EXSTYLE = ^uint32(20 - 1)
	WS_EX_LAYERED = 0x00080000
	WS_EX_TRANSPARENT = 0x00000020
)

func SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	// Find window by title. Wails usually sets title.
    // In our app, it is "Local AI Assistant"
	titlePtr, _ := syscall.UTF16PtrFromString("Local AI Assistant")
	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))

	if hwnd == 0 {
		return fmt.Errorf("window not found")
	}

	exStyle, _, _ := getWindowLong.Call(hwnd, uintptr(GWL_EXSTYLE))

	if ignore {
		exStyle |= WS_EX_LAYERED | WS_EX_TRANSPARENT
	} else {
		exStyle &^= WS_EX_TRANSPARENT
	}

	setWindowLong.Call(hwnd, uintptr(GWL_EXSTYLE), exStyle)
	return nil
}
