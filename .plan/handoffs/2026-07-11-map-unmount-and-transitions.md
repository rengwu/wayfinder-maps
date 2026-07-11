# 2026-07-11 — Clean map unmount + load/unload fade-zoom transition

Short session, one focused change in the **wayfinder** repo
(`github.com:rengwu/wayfinder-map`, branch **`main`**). All work is in the client,
which is one Go raw string in `cmd/wayfinder/shell.go`.

**Predecessor handoff:** `.plan/handoffs/2026-07-11-starmap-verify-and-hud-redesign.md` —
read it for the broader state (GUI verified by eye, HUD-as-bottom-bar redesign, the
open taste-tuning thread). This session's user also **confirmed v1 sign-off** on two
items that handoff left owed: **pan/zoom easing works well** and **live-fold reload
works flawlessly** ("good enough for a v1 effort"). That handoff's "Not done" menu was
updated in place to strike them.

## What this session did — UNCOMMITTED, review before committing

Working tree has **`cmd/wayfinder/shell.go` modified** (+43/-8), not yet committed. Two
user asks, both implemented:

### 1. The map wasn't being unmounted when leaving to splash / map-list
`#splash` and `#maplist` are **semi-transparent** overlays (radial-gradient bg, see the
CSS ~line 87). So the live `nodes`/`edges`/`fogPts` kept rendering *behind* them — the
previous project's constellation bled through while picking another. Old `showSplash`
and `backbtn` handlers only set `currentEffort=null`; they never cleared the scene.

Fix: new `unmountMap()` clears `graph/nodes/edges/byNum/fogPts`, drops selection, idles
the poller. Both **"Open another"** and the **"← maps"** back button now route through
one `leaveMap(next)` path (the back button had the identical latent bug). Read the diff;
not repeated here.

### 2. A ~2s zoom + fade on map load / unload / change
Global `mapAlpha` (0..1), smoothstep-driven by a small `fade` controller
(`startFade(from,to,dur,cb)`), plus a **coupled gentle zoom** `zt = 0.95 + 0.05*mapAlpha`
taken **about the viewport centre**. Applied via an **effective camera `ec`** computed
each frame in `render()`; `w2s`, the map-layer draw transform, and `drawFogLabels` all
read `ec` so nodes/labels/hit-testing stay aligned mid-transition. `ctx.globalAlpha =
mapAlpha` wraps the map layer and the labels.
- **Load** (`loadMap`): `startFade(0,1,2.2)` — dissolves in + eases to full scale.
- **Leave** (`leaveMap`): `startFade(mapAlpha,0,1.8, →unmountMap+navigate)` — dissolves
  out + shrinks, THEN unmounts and shows the next screen.
- **Starfield backdrop deliberately never fades** (it's the constant across all screens).
- **Live-reload folds (`updateGraph`) deliberately untouched** — kept as the smooth
  in-place tween the user just praised; a full fade there would undo it.

## Validated
- Backtick guard `grep -o` backtick `| wc -l` = **2** ✓ (I broke it mid-session by putting
  backticks in JS comments around `ec`/`next` — Go read them as raw-string delimiters and
  `go vet` errored; removed them. **Watch for this**: no backticks anywhere inside the raw
  string, comments included.)
- `go vet ./...` + `go build ./...` clean ✓
- Client JS extracted and `node --check`'d ✓ (extract regex: `<script>\n(.*?)</script>`)

## NOT done — visual eyeball still owed
Could not launch the app to screenshot the dissolve: the **Bash safety classifier was
temporarily unavailable** (`claude-opus-4-8 temporarily unavailable ... cannot determine
safety`) for the whole back half of the session. Retry the launch when it recovers.

Two things to judge by eye, then likely tune:
- Do **2.2s in / 1.8s out** read as tasteful, or too slow/fast? Numbers are trivially
  editable in `loadMap` / `leaveMap`. Zoom depth is the `0.05` in `zt`.
- **Known rough edge:** on load, the DOM chrome (HUD / back button / hint) pops in at
  full opacity while the canvas fades — a slight mismatch. Left intentionally (user asked
  for the *map* to transition). Fade the chrome in step if it bothers the eye. `leaveMap`
  already hides that chrome up front, so only the load side mismatches.

Verify per the saved recipe (memory `verify-starmap-gui-visually`): build to scratchpad,
`wf app ../expensif/.plan/daily-timeline`, read live window bounds fresh before each
`screencapture -R`, frame-diff to prove motion. **Synthetic clicks don't reach the
WKWebView** — to see a leave/enter transition without a real click, temporarily fire
`leaveMap`/`loadMap` from a `setTimeout` in `applyGraph`, screenshot, revert.

## Still-open menu (carried over, unchanged)
- **Taste tuning** (fog strength / edge intensity / layout spread) — grill-me, by eye.
- `docs/starmap-design.md` decision 7 still calls the HUD an "RTS resource bar" — stale.
- Cross-platform folder picker (`/api/pick` is macOS `osascript` only).
- `.pocock-skills/` still untracked — commit or `.gitignore` (also `.plan/` is untracked
  here but is meant to be committed).

## Suggested skills (apply only if available; `.pocock-skills/<name>/SKILL.md`, read & follow)
- `grill-me` — for the transition-timing and taste-tuning forks, one decision at a time.
- `handoff` — when you finish.
