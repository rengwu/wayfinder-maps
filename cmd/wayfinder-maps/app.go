package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

// startAppServer starts the web server on an ephemeral localhost port for the
// native window (or its browser fallback) to point at.
func startAppServer(dir string) (url, title string, err error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", "", err
	}
	go http.Serve(ln, newServer(dir))
	url = "http://" + ln.Addr().String() + "/"
	title = "wayfinder-maps"
	if dir != "" {
		title = "wayfinder-maps — " + dir
	}
	return url, title, nil
}

// browserFallback is what app does when the platform's native window cannot
// open: say why, open the default browser instead, and keep serving — the
// window was only ever a shell around the same server.
func browserFallback(url, reason string) int {
	fmt.Fprintf(os.Stderr, "wayfinder-maps: %s; opening your browser instead\n", reason)
	if err := openBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: open %s manually\n", url)
	}
	select {} // serve until interrupted, like `serve`
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
