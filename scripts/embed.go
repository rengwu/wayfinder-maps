// Package scripts carries the shell helpers so the viewer binary can run them.
package scripts

import _ "embed"

// WayfinderPull is the source of wayfinder-pull.
//
//go:embed wayfinder-pull
var WayfinderPull string
