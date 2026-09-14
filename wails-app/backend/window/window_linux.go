//go:build linux && !test

package window

import (
	"context"
	"fmt"
	"os"
	"time"
)

/*
#cgo LDFLAGS: -lX11 -lXfixes
#include <X11/Xlib.h>
#include <X11/Xatom.h>
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

void set_window_skip_taskbar(Window wid) {
    Display *display = XOpenDisplay(NULL);
    if (display == NULL) return;

    Atom wmState = XInternAtom(display, "_NET_WM_STATE", False);
    Atom skipTaskbar = XInternAtom(display, "_NET_WM_STATE_SKIP_TASKBAR", False);
    Atom skipPager = XInternAtom(display, "_NET_WM_STATE_SKIP_PAGER", False);

    Atom states[2] = {skipTaskbar, skipPager};
    XChangeProperty(display, wid, wmState, XA_ATOM, 32, PropModeAppend, (unsigned char *)states, 2);

    XEvent e;
    e.xclient.type = ClientMessage;
    e.xclient.message_type = wmState;
    e.xclient.display = display;
    e.xclient.window = wid;
    e.xclient.format = 32;
    e.xclient.data.l[0] = 1; // 1 = _NET_WM_STATE_ADD
    e.xclient.data.l[1] = skipTaskbar;
    e.xclient.data.l[2] = skipPager;
    e.xclient.data.l[3] = 0;
    e.xclient.data.l[4] = 0;

    XSendEvent(display, DefaultRootWindow(display), False,
               SubstructureRedirectMask | SubstructureNotifyMask, &e);

    XFlush(display);
    XCloseDisplay(display);
}
Window search_window_tree(Display *display, Window root, pid_t target_pid, Atom pid_atom) {
    Window parent, *children;
    unsigned int num_children;
    Window result = 0;

    if (XQueryTree(display, root, &root, &parent, &children, &num_children)) {
        for (unsigned int i = 0; i < num_children; i++) {
            Atom type;
            int format;
            unsigned long nitems, bytes_after;
            unsigned char *prop;
            if (XGetWindowProperty(display, children[i], pid_atom, 0, 1, False, AnyPropertyType,
                                   &type, &format, &nitems, &bytes_after, &prop) == Success) {
                if (prop != NULL) {
                    pid_t window_pid = *((pid_t *)prop);
                    XFree(prop);
                    if (window_pid == target_pid) {
                        result = children[i];
                        break;
                    }
                }
            }
            result = search_window_tree(display, children[i], target_pid, pid_atom);
            if (result != 0) break;
        }
        if (children) XFree(children);
    }
    return result;
}

Window get_window_by_pid(pid_t pid) {
    Display *display = XOpenDisplay(NULL);
    if (!display) return 0;
    Atom pid_atom = XInternAtom(display, "_NET_WM_PID", True);
    if (pid_atom == None) {
        XCloseDisplay(display);
        return 0;
    }
    Window root = DefaultRootWindow(display);
    Window result = search_window_tree(display, root, pid, pid_atom);
    XCloseDisplay(display);
    return result;
}
*/
import "C"

type linuxModifier struct{}

func init() {
	defaultModifier = &linuxModifier{}
}

func (l *linuxModifier) SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
    var wid C.Window
    for i := 0; i < 5; i++ {
        wid = C.get_window_by_pid(C.pid_t(os.Getpid()))
        if wid != 0 {
            break
        }
        time.Sleep(200 * time.Millisecond)
    }
    if wid == 0 {
        return fmt.Errorf("could not find window ID via X11")
    }

    enable := 0
    if ignore {
        enable = 1
    }

    C.set_window_clickthrough(wid, C.int(enable))
    return nil
}

func (l *linuxModifier) HideFromTaskbar(ctx context.Context) error {
    go func() {
        // Wait for GTK to map the window and set its initial states
        time.Sleep(1 * time.Second)

        // Try to get the window ID, retrying if necessary
        var wid C.Window
        for i := 0; i < 10; i++ {
            wid = C.get_window_by_pid(C.pid_t(os.Getpid()))
            if wid != 0 {
                break
            }
            time.Sleep(200 * time.Millisecond)
        }
        
        if wid != 0 {
            C.set_window_skip_taskbar(wid)
        } else {
            fmt.Println("HideFromTaskbar: could not find window ID via X11")
        }
    }()
    
    return nil
}
