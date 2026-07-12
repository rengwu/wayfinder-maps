//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// helperPath is where the webkit helper lives: next to this binary, the way
// the release archive unpacks. The main binary stays pure Go so status, lint
// and serve run on machines with no GUI stack at all; the webkit linkage is
// quarantined in the helper.
func helperPath() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(exe), "wayfinder-maps-webview")
}

// app starts the server and hands the window to the helper. If the helper is
// missing or its webkit libraries are not installed, fall back to the browser
// with the fix spelled out.
func app(dir string) int {
	if err := dialogToolErr(); err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: %v\n", err)
		return 2
	}
	url, title, err := startAppServer(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: %v\n", err)
		return 2
	}

	h := helperPath()
	if h != "" {
		cmd := exec.Command(h, "--window", url, title)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			return 0
		}
	}
	return browserFallback(url, "native window unavailable — install libwebkit2gtk-4.1 and keep wayfinder-maps-webview next to this binary")
}
