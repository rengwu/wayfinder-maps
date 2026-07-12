package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// webFS holds the browser app — index.html, style.css and the ES modules under
// js/ — baked into the binary so it stays self-contained. Vanilla JS, no
// dependencies, no build step: the modules are served to the browser as-is.
//
//go:embed web
var webFS embed.FS

// webContent is the filesystem the app is served from: the embedded copy, or —
// when WAYFINDER_DEV names the web/ source directory (e.g.
// WAYFINDER_DEV=cmd/wayfinder-maps/web) — the live files on disk, so a JS/CSS
// edit shows on browser reload without recompiling.
func webContent() fs.FS {
	if dir := os.Getenv("WAYFINDER_DEV"); dir != "" {
		// Advisory only: serving proceeds either way, but a typo'd path would
		// otherwise surface as nothing but silent 404s on every request.
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err != nil {
			abs, aerr := filepath.Abs(dir)
			if aerr != nil {
				abs = dir
			}
			fmt.Fprintf(os.Stderr, "wayfinder-maps: warning: WAYFINDER_DEV=%q has no index.html (looked in %s) — the app will 404\n", dir, abs)
		}
		return os.DirFS(dir)
	}
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err) // unreachable: web/ is embedded at build time
	}
	return sub
}
