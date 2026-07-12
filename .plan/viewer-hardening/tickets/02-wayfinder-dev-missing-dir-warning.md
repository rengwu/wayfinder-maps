---
type: task
blocked_by: []
---

# Warn when WAYFINDER_DEV points at nothing

## Question

`WAYFINDER_DEV=<dir>` serves the frontend from disk instead of the embedded
copy. Point it at a directory that doesn't exist (or lacks `index.html`) and
every request 404s with no hint in the server log — verification found this
costs real confusion for a flag whose whole audience is developers.

At startup, when `WAYFINDER_DEV` is set, check the directory exists and holds
`index.html`; print a clear warning naming the resolved path when it doesn't.
Serving should still proceed (the check is advisory), and the embedded path
must stay silent as today.

## Answer

`webContent()` in `cmd/wayfinder-maps/assets.go` now stats
`<dir>/index.html` when `WAYFINDER_DEV` is set and prints

    wayfinder-maps: warning: WAYFINDER_DEV="/nonexistent" has no index.html (looked in /nonexistent) — the app will 404

to stderr, naming both the raw value and the resolved absolute path. Serving
proceeds regardless. Verified at the CLI across all three cases: bogus dir
warns and stays up, a valid dev dir is silent and serves 200, and the
embedded path (unset) is silent and unchanged.
