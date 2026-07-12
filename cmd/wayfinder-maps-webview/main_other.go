//go:build !(linux && cgo)

// The helper only has a job on Linux, where the main binary stays pure Go and
// the webkit linkage lives here. Other platforms embed their webview directly.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "wayfinder-maps-webview: Linux native-window helper; nothing to do on this platform")
	os.Exit(2)
}
