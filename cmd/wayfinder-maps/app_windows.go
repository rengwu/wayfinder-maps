//go:build windows

package main

import (
	"fmt"
	"os"
	"runtime"

	webview2 "github.com/jchv/go-webview2"
)

// app wraps the web server in a native window via WebView2 — pure Go, and the
// runtime ships with Windows 10/11. NewWithOptions returns nil when the
// runtime is missing or the window cannot be created, which is catchable, so
// a failure degrades to the browser rather than a dead binary.
func app(dir string) int {
	runtime.LockOSThread()

	url, title, err := startAppServer(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: %v\n", err)
		return 2
	}

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  title,
			Width:  1120,
			Height: 820,
			Center: true,
		},
	})
	if w == nil {
		return browserFallback(url, "no WebView2 runtime (install it from Microsoft's site for the native window)")
	}
	defer w.Destroy()
	w.Navigate(url)
	w.Run()
	return 0
}
