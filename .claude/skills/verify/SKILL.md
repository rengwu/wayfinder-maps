---
name: verify
description: Build, launch and drive wayfinder-maps to verify a change end-to-end (server + headless browser).
---

# Verifying wayfinder-maps

## Build & fixture

```bash
go build -o /tmp/wayfinder-maps ./cmd/wayfinder-maps
```

The sample map is `internal/wayfinder/testdata/implement-demo/` (20 tickets: every
type and status, undermined decision, anchored and unanchored fog, implement tickets
with a merged PR, a claim, a frontier slot and an implement ticket blocking a
decision). Serve and `status` it directly. `lint` warns about its two claims once
the real clock passes `claimed_at` plus 72h; tests pin `Now` instead.

Expected HUD: decisions 6/15 resolved, 1 claimed, 3 blocked, 1 out of scope;
implement 1/5 resolved, 1 claimed, 1 blocked, 1 out of scope; frontier 3 agents can
take, 2 need you. Ticket 07 carries the PR chip, 09 is the frontier station.

To write a new fixture: an effort dir with `map.md` + `tickets/NN-slug.md`.
Ticket format: YAML frontmatter (`type:` research|prototype|grilling|task|implement,
`blocked_by: [01]`, `claimed_by:` + `claimed_at:`, `undermined_by: [NN]`, `pr:` on
implement only, `afk: true` on task only) then `# Title`, `## Question` (implement:
`## What to build` with `- [ ]` criteria); a `## Answer` section with prose =
resolved, `## Ruled out` = out_of_scope. Map format: `# Name`, `## Destination`,
`## Decisions so far`, `## Implemented so far`, `## Out of scope` (links to
`./tickets/NN-slug.md`), and fog under `## Not yet specified` as
`- **Title.** prose <clears-with: NN>`. Sanity-check with
`wayfinder-maps status <effort-dir>`.

## Launch

```bash
PORT=78xx /tmp/wayfinder-maps serve <effort-dir>   # pick a fresh port EVERY time
```

Gotcha: the user often has an instance (binary `wm`) already listening; a bind
failure only shows in the log while curl happily talks to the OLD server — check
the serve log says "serving", don't trust the port being answerable.

Dev mode: `WAYFINDER_DEV=cmd/wayfinder-maps/web` (path to web/ from cwd) serves
the frontend from disk with `Cache-Control: no-store` instead of the go:embed copy.

## Drive (GUI surface)

Playwright's cached headless shell works without a full playwright install:

```bash
cd <scratch> && npm i playwright-core
# executablePath: ~/Library/Caches/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-mac-arm64/chrome-headless-shell
```

- Wait ~4s after goto: the constellation fades in over 2.2s (a plain
  `--screenshot --virtual-time-budget` capture shows HUD but NO stars — artifact,
  not a bug).
- App state is inside ES modules, but the page can import them: in `page.evaluate`,
  `await import("/js/state.js")` gives `S.byNum[n]` and `await import("/js/draw.js")`
  gives `w2s(node)` for screen coordinates to `page.mouse.click`. Layout is
  deterministic (fixed PRNG seed), so positions repeat across runs.
- Live-reload probe: append `## Answer\ntext` to a ticket while the map is open;
  the 1.5s poller updates the HUD counts in ~2-3s.
- Recents on the splash come from the user's real config — don't dismiss entries.

Flows worth driving: map render (all statuses + fog + edges), click star → panel
markdown, `data-goto` cross-ticket link, Escape closes, wheel zoom (labels thin
out), back button → splash/maplist.

## WebKit

`npx playwright-core install webkit` (~77MB, cached in ms-playwright) then
`require("playwright-core").webkit.launch()` — no executablePath needed.
Covered headlessly: full map render, the label-alpha fade workaround (capture a
frame ~600ms into loadMap: labels must be dim WITH the stars, not opaque over
them), touch tap via a hasTouch context, wheel zoom, panel markdown.

NOT reachable headlessly — manual Safari checklist:
- Trackpad pinch zooms the map, not the page (gesture* events; Playwright's
  WebKit cannot construct GestureEvent, and Chromium never fires them).
- Touch pan/pinch on a real iOS device or simulator.
