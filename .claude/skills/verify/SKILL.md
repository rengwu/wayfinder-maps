---
name: verify
description: Build, launch and drive wayfinder-maps to verify a change end-to-end (server + headless browser).
---

# Verifying wayfinder-maps

## Build & fixture

```bash
go build -o /tmp/wayfinder-maps ./cmd/wayfinder-maps
```

There is no sample map in the repo. Create an effort dir with `map.md` +
`tickets/NN-slug.md`. Ticket format: YAML frontmatter (`type:`, `blocked_by: [01]`,
`claimed_by:`, `undermined_by: [NN]`) then `# Title`, `## Question`; a `## Answer`
section with prose = resolved, `## Ruled out` = out_of_scope. Map format: `# Name`,
`## Destination`, and fog under `## Not yet specified` as
`- **Title.** clears-with: NN`. Sanity-check the fixture with
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
- App state is inside ES modules (not on window). To click a specific star, find
  it by pixel colour on the canvas (frontier glow is gold) — layout is
  deterministic (fixed PRNG seed), so positions repeat across runs.
- Live-reload probe: append `## Answer\ntext` to a ticket while the map is open;
  the 1.5s poller updates the HUD counts in ~2-3s.
- Recents on the splash come from the user's real config — don't dismiss entries.

Flows worth driving: map render (all statuses + fog + edges), click star → panel
markdown, `data-goto` cross-ticket link, Escape closes, wheel zoom (labels thin
out), back button → splash/maplist.
