---
type: task
blocked_by: []
claimed_by: claude-code/fable-5
claimed_at: 2026-07-12T18:11:40Z
---

# WebKit verification pass

## Question

The viewer has only ever been verified in Chromium, and it now carries two
WebKit-specific code paths taken on trust: the label-alpha workaround in
`web/js/draw.js` (WebKit ignores canvas globalAlpha on shadowed fillText) and
the Safari `gesture*` pinch handling from
[touch and trackpad input](01-touch-and-trackpad-input.md), which Chromium
cannot even fire.

Run the app in WebKit and confirm: the map renders (stars, edges, fog,
labels), labels actually fade during map transitions, trackpad pinch zooms
the map rather than the page, and touch works on an iOS device or simulator.
Decide the tooling as part of the task — Playwright's WebKit build may cover
most of it headlessly (and could then join CI); whatever it can't reach
becomes a short manual Safari checklist documented in the verify skill.
