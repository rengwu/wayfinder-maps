<img src="https://i.imgur.com/G2gleBR.png" alt="demo app: one wayfinder map at a moment">

# wayfinder-maps

A read-only CLI and viewer for
[wayfinder](https://github.com/mattpocock/skills/tree/main/skills/engineering/wayfinder)
maps — the markdown planning memory an agent leaves under `.plan/<effort>/` as it charts
a large effort as a graph of investigation tickets. The wayfinder method and its skills
were created by [Matt Pocock](https://github.com/mattpocock); this repo adapts them and
adds the tooling.

- **`wayfinder-maps status <dir>`** — what is resolved, what is in flight, and what is ready
  to claim (the _frontier_: open, unclaimed, every blocker resolved).
- **`wayfinder-maps lint <dir>`** — does the map still tell the truth?
- **`wayfinder-maps serve` / `app`** — the map as a star-map, in a browser or a native window
  ([design notes](docs/starmap-design.md)).

```
$ wayfinder-maps status ../expensif/.plan/daily-timeline
Daily Timeline — continuous days, empty days included
6 resolved · 0 claimed · 4 open · 0 out of scope

Frontier — ready to claim, first by number wins:
  05  Should HandleDaily's two branches converge      grilling
  06  Contain the day-card chrome drift               grilling
  07  The infinite-scroll island's contract           grilling
  10  Test strategy for the date-indexed timeline     grilling

Undermined — resolved on a premise that later changed:
  02  Window size and how older days load             broken by 08

Fog: 5 patches, 1 anchored to a ticket
```

## Install

### Skills, as a Claude Code plugin

```
/plugin marketplace add rengwu/wayfinder-maps
/plugin install wayfinder-maps@wayfinder-maps
```

`/wayfinder-maps` then works in every project; the map lands in that project's `.plan/`. The
plugin bundles `wayfinder-maps` plus the four skills it invokes by name — `grill-me`,
`research`, `prototype`, `domain-modeling`. All five were originally authored by Matt
Pocock in [mattpocock/skills](https://github.com/mattpocock/skills); the copies under
[`skills/`](skills/) are MIT-licensed adaptations.

### Skills, drop-in for any harness

Copy [`skills/`](skills/) into your project and tell your agent to read
`skills/wayfinder-maps/SKILL.md` and follow it; the skills it names live alongside. A pointer
in the project's `AGENTS.md` / `CLAUDE.md` saves retyping.

### The binary

- Download from [Releases](https://github.com/rengwu/wayfinder-maps/releases) — no Go
  toolchain needed, native window and folder dialog on all three platforms. macOS uses
  WKWebView (part of the OS); Windows uses WebView2 (ships with Windows 10/11); the
  Linux archive bundles a small helper that binds the system webkit —
  `libwebkit2gtk-4.1`, preinstalled on most desktops, one
  `sudo apt install libwebkit2gtk-4.1-0` otherwise. `status` and `lint` need nothing
  installed at all, even headless.
- `go install github.com/rengwu/wayfinder-maps/cmd/wayfinder-maps@latest`

Skill and binary are independent — each just reads the on-disk contract.

## Why

The map is shared memory: sessions orient to it before choosing work, and when it drifts
it lies to them silently. The skill makes most drift unrepresentable and verifies each
session's delta, no tool assumed. What a delta check cannot see is an edit made outside
the protocol, or two parallel sessions taking the same ticket number. `lint` is `fsck`
for those: it runs after the fact and re-establishes the base the next session rests on.

## The format contract

The contract is the skill's local-markdown adapter,
[`TRACKER-MARKDOWN.md`](skills/wayfinder-maps/TRACKER-MARKDOWN.md) — the skill itself is
tracker-agnostic method. This tool implements that adapter; any other reader should
target the same file. In summary: structure the facts, leave the prose alone. A ticket's
type and edges have exactly one correct value a machine can check. Its **Question** and
**Answer**, the map's **Destination** and **Notes**, and each one-line gist in
Decisions-so-far are lossy prose, never parsed.

A ticket is `.plan/<effort>/tickets/NN-<slug>.md`:

```markdown
---
type: research | prototype | grilling | task
blocked_by: [02, 03] # [] when none
claimed_by: <session id> # while a session holds it
claimed_at: 2026-07-10T09:00:00Z # RFC 3339, alongside claimed_by
undermined_by: [08] # optional — resolved on a premise 08 later broke
assets: [../assets/x.html.approved] # optional
---

# <Ticket title>

## Question

…

## Answer # its presence _is_ the resolution

…
```

### Status is derived

There is no `status:` field — every value it could hold is already written in the file,
and a second copy is free to go stale:

| Derived status | When                                                |
| -------------- | --------------------------------------------------- |
| `resolved`     | the body has an `## Answer` **with prose under it** |
| `out_of_scope` | the body has a `## Ruled out` with prose under it   |
| `claimed`      | neither, and `claimed_by` is set                    |
| `open`         | none of the above                                   |

Closure is read first, so a `claimed_by` left on a closed ticket is inert litter. And it
is the prose, not the heading, that closes a ticket: a session that types `## Answer` and
dies has resolved nothing — the ticket stays claimed, and its claim goes stale on
schedule. `lint` errors on the empty heading outright.

**`undermined_by`** marks a decision whose premise a later ticket broke; without it a
renderer paints the node green and launders a live problem into a checkmark.
**`claimed_at`** separates a dead session from live work. And `out_of_scope` stays
distinct from `resolved` — a scope boundary is not a step on the route, and a parser that
lumps them in over-counts.

### Fenced code blocks are not structure

Every scan — headings, titles, bullets, links — ignores what sits inside a ` ``` ` or
`~~~` fence. A ticket that quotes the ticket format contains the line `## Answer`, and
must not thereby resolve itself. The rule cuts one way only: an answer whose entire body
is a code fence is still an answer. Any other reader of these files owes the same rule.

### One session at a time

The adapter forbids concurrent sessions: git merges a duplicate ticket number cleanly and
neither session can see it. `lint` still reports duplicates, because a merge can happen
anyway.

Fog patches in the map's `## Not yet specified` are bullets with a bolded lead title and
an optional anchor; the rest stays loose prose:

```markdown
- **Today's row.** Whether today is visually distinguished. <clears-with: 04>
```

## Checks

Errors: dangling or self `blocked_by`; `blocked_by` cycles; duplicate ticket numbers; a
ticket carrying both `## Answer` and `## Ruled out`; a closing heading with nothing under
it; a resolved ticket missing from Decisions-so-far; an out-of-scope ticket listed as a
decision, or missing from Out-of-scope; Decisions-so-far pointing at an unresolved
ticket; a fog patch duplicating a live ticket's title, or anchored to a resolved ticket;
a missing Destination; an unknown type; a leftover `status:` field that disagrees with
the body.

Warnings: loose (pre-frontmatter) headers; a `status:` field that merely duplicates the
body; a claim with no owner or older than 72h; a claim left on a closed ticket; a ticket
blocked by something out of scope, which can never unblock.

Checks that need judgment — does a gist say what its answer says, does a fog title
duplicate a live ticket _in substance_ — stay with the skill; no parser can make them.

Exit `0` clean, `1` errors found, `2` no map at that path.

## Reading legacy maps

Two older shapes still parse, each lints as a warning, and neither is written any more:

- the loose header (`Type:` / `Status:` / `Blocked by:` after the H1), from before
  frontmatter;
- a stored `status:` field, from before status was derived. Where it agrees with the
  body, `lint` says to delete it; where it disagrees, that is an error and the body wins.

## Build from source

```
go build ./cmd/wayfinder-maps      # pure Go: status, lint, serve
go test ./...
```

The native window needs cgo on macOS (WKWebView); Windows is pure Go (WebView2). On
Linux the webkit linkage lives in a separate helper, `./cmd/wayfinder-maps-webview`,
built with `libwebkit2gtk-4.1-dev` and kept next to the main binary. Releases are cut
by tagging `v*` — CI builds the helper on Linux runners and GoReleaser packs everything.
