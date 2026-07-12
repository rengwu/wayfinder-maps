//go:build !darwin && !windows && !linux

package main

import (
	"fmt"
	"os"
)

// Platforms without a webview story get the server plus their browser.
func app(dir string) int {
	url, _, err := startAppServer(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: %v\n", err)
		return 2
	}
	return browserFallback(url, "no native window on this platform")
}
