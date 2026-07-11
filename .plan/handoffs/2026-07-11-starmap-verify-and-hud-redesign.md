# 2026-07-11 — Star-map seen for the first time; HUD becomes a bottom bar

This session did the **visual verification pass** the previous handoff demanded, then
acted on what that pass revealed. All work is in the **wayfinder repo** (now its own
standalone repo: `github.com:rengwu/wayfinder-map`, branch/remote **`main`**).

**Predecessor handoff — read it first:** `../expensif/.plan/handoffs/2026-07-11-wayfinder-starmap-app.md`.
It built the whole GUI blind (no display) and its explicit first instruction was "run
`wayfinder app` and look." That is what happened here.

## The headline: it renders, and the motion is real

The prior handoff's "NOT verified — no human or tool has seen it render" is now retired.
Verified by eye via `wayfinder app ../expensif/.plan/daily-timeline` (screencapture on
this Mac, see the memory recipe below):

- Renders correctly; status hierarchy reads at a glance (frontier gold+large, resolved
  dim blue embers, undermined ringed).
- **Motion layer is live** — proven by frame-diffing: frontier pulse, the undermined
  cracked halo with travelling gaps (ticket 02), flow particles on satisfied edges,
  idle bob. Spatial stability holds (no drift; an early "drift" reading was just the
  native window relocating under a fixed capture rect).
- **Detail panel works end to end** — click→panel, markdown render (bold, inline code,
  cross-ticket links, lists), camera ease-on-select, and the new HUD-hide.

## What changed — commit `9a88b48`

One commit, `9a88b48` (`feat: HUD as a flat bottom bar; label declutter and visual
tuning`). Read it and the diff; not repeated here. In short:

- **HUD redesign (user-driven):** floating top-left card → **flat full-width bottom bar**
  (title+destination left, counts/progress/legend right). Slides down + fades when a
  detail panel opens. This dissolved the occlusion bug (the "← maps" button had sat on
  the HUD title).
- **Label de-collision:** greedy vertical declutter in `drawLabels` so overlapping node
  labels (04/05) push apart.
- **First taste pass:** fog nebulae larger/brighter; edges slightly stronger; `fitCamera`
  reserves top (hint) and bottom (bar) bands and centres into what's left; hint moved to
  top-centre; detail-panel scrollbar pulled to the far viewport edge via `margin-right`.

## Git history note — a rewrite happened, read this

- The branch `derive-status-from-body` was **renamed to `main`** and pushed to origin.
  `main` on `github.com:rengwu/wayfinder-map` is now the full GUI (the old handoff's
  "main still has the CLI-only tool" is stale).
- The commit was amended to **remove the `Co-Authored-By: Claude` trailer** (was
  `497eca3`, now `9a88b48`) and **force-pushed** (`--force-with-lease`). Local, remote
  tracking, and remote `refs/heads/main` all agree on `9a88b48`. Old SHA lingers only in
  the local reflog.

## State of the tree

Clean except **`.pocock-skills/` is untracked** — the skills library the user vendored in.
Decide: commit it, or add to `.gitignore`. Left untouched this session. `go build/vet/test`
all pass; JS syntax + backtick-guard (==2) pass.

## Not done — the next session's menu

- **Taste tuning is still open (grill-me, one at a time):** the user explicitly wants to
  drive fog strength (may now be *too* strong?), edge intensity, and layout spread by eye.
  This was the agreed next step before "commit first" interrupted.
- ~~**Human eyeball still owed on:** pan/zoom by hand, and **live-reload** folding a live
  edit in.~~ **CONFIRMED by user (2026-07-11):** pan/zoom easing works well; live-fold
  reload of a live edit works flawlessly. Good enough for a v1 effort. (Note for future
  work: live-reload rewrites ticket files, so test against a COPY of an effort, never the
  real `.plan`.)
- **`docs/starmap-design.md` decision 7** still describes the HUD as an "RTS resource
  bar" in the corner — update it to match the bottom-bar reality.
- **Cross-platform folder picker** unbuilt (`/api/pick` is macOS osascript only).
- **`type-as-celestial-body` v2** (design doc) still deferred.
- **Nothing in expensif points at this tool** — no `make wayfinder`, no README line.
  Deferred across six sessions now.

## Environment & gotchas

- `wayfinder {app,serve,status,lint} [path]`. `app` needs cgo (system WebKit, macOS);
  `serve` is pure stdlib (`PORT`, default 7777). Client is one Go raw string in
  `cmd/wayfinder/shell.go` — **no backticks anywhere in it** (guard: `grep -o` backtick
  `| wc -l` must be 2).
- **Visual verification recipe is saved as a memory:** `verify-starmap-gui-visually.md`
  (in this repo's `memory/`). Key points: the native window **relocates on focus** so
  read its bounds fresh right before each `screencapture -R`; **synthetic clicks do NOT
  reach the WKWebView** (tried CGEventPost/JXA on both taps) — to verify a click-driven
  interaction, temporarily add a `setTimeout(...openPanel(...))` in `applyGraph`,
  screenshot, then revert.

## Suggested skills

Apply only if available; `.pocock-skills/<name>/SKILL.md`, not auto-discovered — read and follow.

- `grill-me` — the natural mode for the remaining visual taste forks (fog/edge/layout);
  the user drives one decision at a time, and drove the whole redesign this way.
- `handoff` — when you finish.
