//go:build darwin && !cgo

package main

import (
	"fmt"
	"os"
)

// A source build without cgo has no WKWebView binding. Release binaries for
// macOS are always cgo builds, so only from-source users see this.
func app(string) int {
	fmt.Fprintln(os.Stderr, "wayfinder-maps: 'app' needs a cgo build on macOS; use 'wayfinder-maps serve' instead")
	return 2
}
