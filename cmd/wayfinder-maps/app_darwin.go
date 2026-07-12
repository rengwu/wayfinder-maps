//go:build darwin && cgo

package main

import (
	"fmt"
	"os"
	"runtime"

	webview "github.com/webview/webview_go"
)

// app wraps the web server in a native window — WKWebView, part of macOS, so
// there is nothing to install. The webview owns the main thread, which Cocoa
// requires, so LockOSThread pins us there.
func app(dir string) int {
	runtime.LockOSThread()

	url, title, err := startAppServer(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: %v\n", err)
		return 2
	}

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle(title)
	w.SetSize(1120, 820, webview.HintNone)
	w.Navigate(url)
	w.Run()
	return 0
}
