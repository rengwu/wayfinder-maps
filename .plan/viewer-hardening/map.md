# Star-map viewer hardening

## Destination

The maps viewer is first-class on every input device (mouse, trackpad, touch)
and the newly modular frontend under `cmd/wayfinder-maps/web/` cannot silently
regress — its sharp edges warned about, its pure logic tested, its syntax
checked before anything ships embedded.

## Notes

This effort carries execution: its tickets do the work, not just decide it.
The viewer's code lives in `cmd/wayfinder-maps/` (Go server + vanilla-JS ES
modules under `web/`, no build step — a constraint to preserve). Verify changes
end-to-end with the repo's `verify` skill (`.claude/skills/verify/SKILL.md`),
which documents how to build, serve a fixture map, and drive the app headless.

## Decisions so far

<!-- one line per resolved ticket: gist + link -->

## Not yet specified

- **A WebKit verification pass.** The renderer carries a WebKit-specific
  label-alpha workaround but has only ever been verified in Chromium; whether
  this wants Playwright WebKit in the loop or a manual Safari checklist is
  unclear until touch/gesture work shows how far headless WebKit gets. clears-with: 01
- **Whether the frontend's ES5-style JS modernizes.** The modules kept the old
  `var`-and-concatenation style verbatim through the refactor; module-by-module
  modernization is possible now, but it's not yet clear it's worth the churn.

## Out of scope

- **Editing from the viewer** — claiming or resolving tickets in the panel is
  ruled out for now: all writes go through files, which keeps agents and humans
  on equal footing and the viewer honest as a pure lens on the map.
