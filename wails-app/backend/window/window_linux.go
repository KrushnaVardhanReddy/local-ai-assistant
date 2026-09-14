//go:build linux && !test

package window

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

/*
#cgo LDFLAGS: -lX11 -lXfixes
#include <X11/Xlib.h>
#include <X11/extensions/shape.h>
#include <X11/extensions/Xfixes.h>
#include <stdlib.h>

void set_window_clickthrough(Window wid, int enable) {
    Display *display = XOpenDisplay(NULL);
    if (display == NULL) {
        return;
    }

    if (enable) {
        XserverRegion region = XFixesCreateRegion(display, NULL, 0);
        XFixesSetWindowShapeRegion(display, wid, ShapeInput, 0, 0, region);
        XFixesDestroyRegion(display, region);
    } else {
        XFixesSetWindowShapeRegion(display, wid, ShapeInput, 0, 0, 0);
    }

    XFlush(display);
    XCloseDisplay(display);
}
*/
import "C"

type linuxModifier struct{}

func init() {
	defaultModifier = &linuxModifier{}
}

func (l *linuxModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
    // In order to call X11 functions we need the Window ID. Wails currently does not expose this cleanly
    // so we will find it using xdotool based on the PID. Wait briefly for window mapping if needed.
    // Try up to 5 times.
    var wid string
    var err error
    for i := 0; i < 5; i++ {
        wid, err = getWindowID()
        if err == nil && wid != "" {
            break
        }
        time.Sleep(200 * time.Millisecond)
    }
    if err != nil || wid == "" {
        return fmt.Errorf("could not find window ID: %v", err)
    }

    var widDec uint64
    fmt.Sscanf(wid, "%d", &widDec)

    enable := 0
    if ignore {
        enable = 1
    }

    C.set_window_clickthrough(C.Window(widDec), C.int(enable))
    return nil
}

func getWindowID() (string, error) {
    // This is a naive approach assuming we can find our own window via xdotool or wmctrl.
    // A robust Linux app would use gdk/gtk APIs natively but we are abstracted behind Wails.
    // We can search by our executable PID.
    pid := os.Getpid()

    // Command to find windows for a specific PID using xdotool
    cmd := exec.Command("xdotool", "search", "--pid", fmt.Sprintf("%d", pid))
    out, err := cmd.Output()
    if err != nil {
        return "", err
    }

    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    if len(lines) == 0 || lines[0] == "" {
        return "", fmt.Errorf("no window found")
    }

    return lines[0], nil
}
