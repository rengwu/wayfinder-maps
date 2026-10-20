---
name: wayfinder-orca
description: Show a wayfinder map in a tab of Orca's built-in browser. Use when the user asks to show, open, view, render or refresh a wayfinder map (or "the map", "the plan") inside Orca, whether the map lives in a local .plan/ folder or in GitHub issues.
---

# Wayfinder map in Orca

Run `wayfinder-orca`; it is on PATH. If the shell can't find it, run it from the wayfinder-maps clone this skill is linked from, at `../../scripts/wayfinder-orca` relative to this skill's directory. It starts `wayfinder-maps serve` in the background if nothing answers on port 7777, then opens a tab in Orca's built-in browser, in the current worktree, at `http://localhost:7777/?path=<dir>`. If a tab already shows that URL, the script switches to it. It prints the URL.

After it runs, tell the user the map is open in the Orca browser tab titled "wayfinder-maps" in this worktree.

The page polls the files on disk, so the open tab redraws after you edit a ticket. Do not reopen it.

## Map in a local `.plan/`

```sh
wayfinder-orca            # enclosing git repo: list of its maps
wayfinder-orca .plan/foo  # one effort: straight to its map
```

## Map kept as GitHub issues

Find the map issue with `gh issue list -l wayfinder:map --state all`, then pass it as `owner/repo#N`:

```sh
wayfinder-orca acme/app#29
wayfinder-orca --every 120 acme/app#29
```

The script mirrors the map with `wayfinder-pull` into `~/.cache/wayfinder-maps/<owner>-<repo>/.plan/<N>`, outside the working tree, and opens that copy. `--every SECS` keeps re-pulling in the background so the tab follows the tracker. Rerunning the command for the same map replaces the old loop, and running it without `--every` stops the loop. Each pull is swapped in whole, so the tab never shows half a map.

## Outside the current worktree

The tab opens in the Orca worktree that contains the shell's working directory. To put it in a different one, set `ORCA_WORKTREE` to any `orca` worktree selector, such as `ORCA_WORKTREE=path:/abs/repo`.

## Failures

- `wayfinder-maps not found`: run `go install ./cmd/wayfinder-maps` from the repo root.
- `the server on :7777 ignores ?path=`: an older build holds the port. Stop it with `pkill -f 'wayfinder-maps serve'` and rerun.
- The server log is at `$TMPDIR/wayfinder-maps-7777.log`.
