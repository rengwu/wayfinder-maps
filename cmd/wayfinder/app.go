//go:build cgo

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"

	webview "github.com/webview/webview_go"
)

// app wraps the same web server in a native OS window (WKWebView on macOS). The
// server runs on an ephemeral localhost port in a goroutine; the webview owns
// the main thread, which Cocoa requires, so LockOSThread pins us there.
func app(dir string) int {
	runtime.LockOSThread()

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder: %v\n", err)
		return 2
	}
	go http.Serve(ln, newServer(dir))
	url := "http://" + ln.Addr().String() + "/"

	title := "wayfinder"
	if dir != "" {
		title = "wayfinder — " + dir
	}
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle(title)
	w.SetSize(1120, 820, webview.HintNone)
	w.Navigate(url)
	w.Run()
	return 0
}
