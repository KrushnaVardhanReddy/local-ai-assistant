//go:build darwin
// +build darwin

package clickthrough

import (
	"context"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#include <stdbool.h>

void SetIgnoresMouseEvents(bool ignore) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSApplication *app = [NSApplication sharedApplication];
        NSArray *windows = [app windows];
        if ([windows count] > 0) {
            NSWindow *window = [windows objectAtIndex:0];
            [window setIgnoresMouseEvents:ignore ? YES : NO];
        }
    });
}
*/
import "C"

func SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	C.SetIgnoresMouseEvents(C.bool(ignore))
	return nil
}
