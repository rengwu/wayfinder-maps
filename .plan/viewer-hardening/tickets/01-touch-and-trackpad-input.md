---
type: task
blocked_by: []
---

# Touch and trackpad input for the star-map

## Question

`web/js/input.js` wires mouse events only. On any touchscreen the map is
completely inert — no pan, no zoom, no way to open a ticket. On macOS
trackpads, pinch works in Chrome/Firefox by accident (it arrives as
ctrl+wheel) but in Safari pinch fires `gesture*` events instead, so it zooms
the page rather than the map.

Wire the missing input: one-finger drag pans, tap selects a star (with a
touch-sized hit radius), two-finger pinch zooms about the gesture midpoint
while tracking its pan, and Safari's gesture events drive the same zoom. The
browser's own touch panning/zooming must be suppressed on the canvas, and
mouse behavior — including wheel-to-zoom — must not change.
