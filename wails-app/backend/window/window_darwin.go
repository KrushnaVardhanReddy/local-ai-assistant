//go:build darwin && !test

package window

import (
	"context"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

void set_window_ignores_mouse_events(int ignore) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSApplication *app = [NSApplication sharedApplication];
        for (NSWindow *window in [app windows]) {
            [window setIgnoresMouseEvents:(BOOL)ignore];
        }
    });
}

void mac_hide_from_dock() {
    dispatch_async(dispatch_get_main_queue(), ^{
        // NSApplicationActivationPolicyAccessory (1) hides the app from the macOS Dock and CMD+Tab app switcher
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    });
}

void mac_set_capture_excluded(int excluded) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSApplication *app = [NSApplication sharedApplication];
        for (NSWindow *window in [app windows]) {
            if (excluded) {
                [window setSharingType:NSWindowSharingNone];
            } else {
                [window setSharingType:NSWindowSharingReadWrite];
            }
        }
    });
}
*/
import "C"

type darwinModifier struct{}

func init() {
	defaultModifier = &darwinModifier{}
}

func (d *darwinModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	enable := 0
	if ignore {
		enable = 1
	}
	C.set_window_ignores_mouse_events(C.int(enable))
	return nil
}

func (d *darwinModifier) HideFromTaskbar(ctx context.Context) error {
	C.mac_hide_from_dock()
	return nil
}

func (d *darwinModifier) SetCaptureExcluded(ctx context.Context, excluded bool) error {
	enable := 0
	if excluded {
		enable = 1
	}
	C.mac_set_capture_excluded(C.int(enable))
	return nil
}
