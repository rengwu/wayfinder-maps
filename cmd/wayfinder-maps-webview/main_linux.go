//go:build linux && cgo

// wayfinder-maps-webview is the Linux native-window helper. The main binary
// stays pure Go so status/lint/serve run on headless machines; everything that
// needs webkit2gtk lives here, exec'd from the same directory. It answers
// three calls:
//
//	wayfinder-maps-webview --check                exit 0 (loading it is the check)
//	wayfinder-maps-webview --pick-folder          print a chosen folder, or nothing
//	wayfinder-maps-webview --window <url> <title> open the native window
package main

/*
#cgo pkg-config: gtk+-3.0 webkit2gtk-4.1
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>
#include <stdlib.h>

static char* wf_pick_folder(void) {
	GtkWidget *dlg = gtk_file_chooser_dialog_new("Open a project folder", NULL,
		GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER,
		"_Cancel", GTK_RESPONSE_CANCEL, "_Open", GTK_RESPONSE_ACCEPT, NULL);
	char *out = NULL;
	if (gtk_dialog_run(GTK_DIALOG(dlg)) == GTK_RESPONSE_ACCEPT)
		out = gtk_file_chooser_get_filename(GTK_FILE_CHOOSER(dlg));
	gtk_widget_destroy(dlg);
	while (g_main_context_iteration(NULL, FALSE));
	return out;
}

static void wf_window(const char *url, const char *title) {
	GtkWidget *win = gtk_window_new(GTK_WINDOW_TOPLEVEL);
	gtk_window_set_title(GTK_WINDOW(win), title);
	gtk_window_set_default_size(GTK_WINDOW(win), 1120, 820);
	GtkWidget *wv = webkit_web_view_new();
	gtk_container_add(GTK_CONTAINER(win), wv);
	webkit_web_view_load_uri(WEBKIT_WEB_VIEW(wv), url);
	g_signal_connect(win, "destroy", G_CALLBACK(gtk_main_quit), NULL);
	gtk_widget_show_all(win);
	gtk_main();
}
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: wayfinder-maps-webview --check | --pick-folder | --window <url> [title]")
		os.Exit(2)
	}
	switch args[0] {
	case "--check":
		// Reaching main means the dynamic linker found webkit2gtk.
		os.Exit(0)
	case "--pick-folder":
		if C.gtk_init_check(nil, nil) == 0 {
			os.Exit(1) // no display
		}
		p := C.wf_pick_folder()
		if p != nil {
			fmt.Println(C.GoString(p))
			C.free(unsafe.Pointer(p))
		}
	case "--window":
		if len(args) < 2 {
			os.Exit(2)
		}
		title := "wayfinder-maps"
		if len(args) > 2 {
			title = args[2]
		}
		if C.gtk_init_check(nil, nil) == 0 {
			fmt.Fprintln(os.Stderr, "wayfinder-maps-webview: no display")
			os.Exit(1)
		}
		curl, ctitle := C.CString(args[1]), C.CString(title)
		defer C.free(unsafe.Pointer(curl))
		defer C.free(unsafe.Pointer(ctitle))
		C.wf_window(curl, ctitle)
	default:
		os.Exit(2)
	}
}
