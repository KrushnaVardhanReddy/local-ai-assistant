//go:build linux && !test

package window

import (
	"context"
	"fmt"
	"os"
	"time"
)

/*
#cgo pkg-config: gtk+-3.0 x11 xfixes
#include <gtk/gtk.h>
#include <gdk/gdkx.h>
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/extensions/shape.h>
#include <X11/extensions/Xfixes.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

struct ClickthroughData {
    Window wid;
    int enable;
};

gboolean do_set_clickthrough(gpointer data) {
    struct ClickthroughData *cdata = (struct ClickthroughData *)data;
    Window wid = cdata->wid;
    int enable = cdata->enable;
    free(cdata);

    // 1. GTK level clickthrough setting (works natively for GTK on both Wayland and X11)
    GList *toplevels = gtk_window_list_toplevels();
    for (GList *iter = toplevels; iter != NULL; iter = iter->next) {
        GtkWindow *win = GTK_WINDOW(iter->data);
        if (win && GTK_IS_WINDOW(win)) {
            if (enable) {
                gtk_widget_set_opacity(GTK_WIDGET(win), 0.65);
            } else {
                gtk_widget_set_opacity(GTK_WIDGET(win), 1.0);
            }
            gtk_widget_queue_draw(GTK_WIDGET(win));

            GdkWindow *gdk_win = gtk_widget_get_window(GTK_WIDGET(win));
            if (gdk_win) {
                if (enable) {
                    cairo_region_t *empty_region = cairo_region_create();
                    gdk_window_input_shape_combine_region(gdk_win, empty_region, 0, 0);
                    cairo_region_destroy(empty_region);
                } else {
                    gdk_window_input_shape_combine_region(gdk_win, NULL, 0, 0);
                }
            }
        }
    }
    if (toplevels) g_list_free(toplevels);

    // 2. X11 fallback
    if (wid != 0) {
        Display *display = XOpenDisplay(NULL);
        if (display != NULL) {
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
    }

    return G_SOURCE_REMOVE;
}

void set_window_clickthrough(Window wid, int enable) {
    struct ClickthroughData *cdata = (struct ClickthroughData *)malloc(sizeof(struct ClickthroughData));
    cdata->wid = wid;
    cdata->enable = enable;
    g_idle_add(do_set_clickthrough, cdata);
}

gboolean do_set_skip_taskbar(gpointer data) {
    Window wid = (Window)(uintptr_t)data;

    // 1. Direct GTK toplevel enumeration (bypasses PID lookup issues in dev mode)
    GList *toplevels = gtk_window_list_toplevels();
    for (GList *iter = toplevels; iter != NULL; iter = iter->next) {
        GtkWindow *win = GTK_WINDOW(iter->data);
        if (win && GTK_IS_WINDOW(win)) {
            gtk_window_set_skip_taskbar_hint(win, TRUE);
            gtk_window_set_skip_pager_hint(win, TRUE);
            gtk_window_set_type_hint(win, GDK_WINDOW_TYPE_HINT_UTILITY);
        }
    }
    if (toplevels) g_list_free(toplevels);

    // 2. Direct X11 window lookup fallback
    GdkDisplay *gdk_display = gdk_display_get_default();
    if (gdk_display && wid != 0) {
        GdkWindow *gdk_window = gdk_x11_window_lookup_for_display(gdk_display, wid);
        if (gdk_window) {
            gpointer widget = NULL;
            gdk_window_get_user_data(gdk_window, &widget);
            if (widget && GTK_IS_WINDOW(widget)) {
                GtkWindow *win = GTK_WINDOW(widget);
                gtk_window_set_skip_taskbar_hint(win, TRUE);
                gtk_window_set_skip_pager_hint(win, TRUE);
                gtk_window_set_type_hint(win, GDK_WINDOW_TYPE_HINT_UTILITY);
            }
        }
    }

    if (wid != 0) {
        Display *display = XOpenDisplay(NULL);
        if (display != NULL) {
            Atom net_wm_state = XInternAtom(display, "_NET_WM_STATE", False);
            Atom skip_taskbar = XInternAtom(display, "_NET_WM_STATE_SKIP_TASKBAR", False);
            Atom skip_pager = XInternAtom(display, "_NET_WM_STATE_SKIP_PAGER", False);
            Atom net_wm_type = XInternAtom(display, "_NET_WM_WINDOW_TYPE", False);
            Atom type_utility = XInternAtom(display, "_NET_WM_WINDOW_TYPE_UTILITY", False);

            XEvent event;
            memset(&event, 0, sizeof(event));
            event.type = ClientMessage;
            event.xclient.window = wid;
            event.xclient.message_type = net_wm_state;
            event.xclient.format = 32;
            event.xclient.data.l[0] = 1;
            event.xclient.data.l[1] = skip_taskbar;
            event.xclient.data.l[2] = skip_pager;
            event.xclient.data.l[3] = 1;

            Window root = DefaultRootWindow(display);
            XSendEvent(display, root, False, SubstructureRedirectMask | SubstructureNotifyMask, &event);
            XChangeProperty(display, wid, net_wm_state, XA_ATOM, 32, PropModeReplace, (unsigned char *)&skip_taskbar, 1);
            XChangeProperty(display, wid, net_wm_type, XA_ATOM, 32, PropModeReplace, (unsigned char *)&type_utility, 1);

            XFlush(display);
            XCloseDisplay(display);
        }
    }

    return G_SOURCE_REMOVE;
}

void set_window_skip_taskbar(Window wid) {
    g_idle_add(do_set_skip_taskbar, (gpointer)(uintptr_t)wid);
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
	for i := 0; i < 3; i++ {
		wid = C.get_window_by_pid(C.pid_t(os.Getpid()))
		if wid != 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
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
