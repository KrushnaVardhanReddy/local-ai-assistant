//go:build linux
// +build linux

package clickthrough

import (
	"context"
)

/*
#cgo pkg-config: gtk+-3.0 x11 xfixes xext
#include <gtk/gtk.h>
#include <gdk/gdkx.h>
#include <X11/Xlib.h>
#include <X11/extensions/shape.h>
#include <X11/extensions/Xfixes.h>
#include <stdbool.h>
#include <stdint.h>

static gboolean set_ignores_mouse_events_main_thread(gpointer user_data) {
    bool ignore = (bool)(intptr_t)user_data;

    GList *windows = gtk_window_list_toplevels();
    if (windows == NULL) {
        return G_SOURCE_REMOVE;
    }

    // Assuming the first window is the main window
    GtkWindow *window = GTK_WINDOW(windows->data);
    GdkWindow *gdk_win = gtk_widget_get_window(GTK_WIDGET(window));
    if (gdk_win == NULL) {
        g_list_free(windows);
        return G_SOURCE_REMOVE;
    }

    Display *display = gdk_x11_get_default_xdisplay();
    Window xid = gdk_x11_window_get_xid(gdk_win);

    if (ignore) {
        XRectangle rect = {0, 0, 0, 0};
        // Create an empty region (no clickable area)
        XserverRegion region = XFixesCreateRegion(display, &rect, 1);
        XFixesSetWindowShapeRegion(display, xid, ShapeInput, 0, 0, region);
        XFixesDestroyRegion(display, region);
    } else {
        // Reset to default shape (clickable)
        XFixesSetWindowShapeRegion(display, xid, ShapeInput, 0, 0, 0);
    }

    g_list_free(windows);
    return G_SOURCE_REMOVE;
}

void SetIgnoresMouseEvents(bool ignore) {
    g_idle_add(set_ignores_mouse_events_main_thread, (gpointer)(intptr_t)ignore);
}
*/
import "C"

func SetIgnoreMouseEvents(ctx context.Context, ignore bool) error {
	C.SetIgnoresMouseEvents(C.bool(ignore))
	return nil
}
