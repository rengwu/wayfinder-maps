# wayfinder

A read-only CLI for [wayfinder](https://agentskills.io) maps — the markdown planning
memory an agent leaves under `.plan/<effort>/` as it charts a large effort as a graph
of investigation tickets.

It answers two questions without an LLM:

- **`wayfinder status <dir>`** — what is resolved, what is in flight, and what is ready
  to claim (the *frontier*: open, unclaimed, every blocker resolved).
- **`wayfinder lint <dir>`** — does the map still tell the truth?

```
$ wayfinder status ../expensif/.plan/daily-timeline
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

(Illustrative. It is a snapshot of one map at one moment — do not trust it as that
map's current state, which is the whole point of the tool.)

## Why

The map is shared memory: concurrent and future agent sessions orient to it before
choosing what to work on. When it drifts, it lies to them silently. It has already
happened — a project's `AGENTS.md` read "4 of 9 tickets resolved" while the map said 5.

**This tool is `fsck`, not a gate.** The skill makes most drift unrepresentable by
deriving what it can and structuring only the rest, and it verifies the delta of a single
session against the invariants — no tool required, and none assumed. What a delta check
cannot see is an edit made outside that protocol, and two parallel sessions reaching for
the same ticket number. Those are what `lint` is for. It runs after the fact, it
re-establishes the base case the next session's induction rests on, and its absence
costs you recovery, not correctness.

## The format contract

The contract is the skill's **local-markdown adapter**, `TRACKER-MARKDOWN.md` — not the
skill itself, which is tracker-agnostic method. This tool implements that adapter; a
viewer or GUI should target the same file. What follows is a summary of it.

Structure the facts, leave the prose alone. A ticket's type and edges have exactly one
correct value and a machine can check them. Its **Question**, its **Answer**, the map's
**Destination** and **Notes**, and the one-line gist in Decisions-so-far are prose, are
lossy by design, and are never parsed. Structure those and you have built a Jira.

A ticket is `.plan/<effort>/tickets/NN-<slug>.md`:

```markdown
---
type: research | prototype | grilling | task
blocked_by: [02, 03]                  # [] when none
claimed_by: <session id>              # while a session holds it
claimed_at: 2026-07-10T09:00:00Z      # RFC 3339, alongside claimed_by
undermined_by: [08]                   # optional — resolved on a premise 08 later broke
assets: [../assets/x.html.approved]   # optional
---

# <Ticket title>

## Question
…
## Answer          # its presence *is* the resolution
…
```

### There is no `status` field

Status is **derived**, because every value it could hold is already written in the file
and a second copy is free to go stale:

| Derived status | When |
|---|---|
| `resolved` | the body has an `## Answer` **with prose under it** |
| `out_of_scope` | the body has a `## Ruled out` with prose under it |
| `claimed` | neither, and `claimed_by` is set |
| `open` | none of the above |

Closure is read first, so a `claimed_by` left behind on a closed ticket is inert litter,
not a broken invariant — it can never hold the frontier. Writing the answer is the act of
resolving, so "resolved with no answer" is unrepresentable rather than merely caught.

It is the prose, not the heading, that closes a ticket. A session that types `## Answer`
and dies has resolved nothing: the ticket stays claimed, and its claim goes stale on
schedule. Were the bare heading enough, that ticket would read as finished and the stale
claim — the one tell that a session died — would be suppressed as litter on a closed
ticket. `lint` also errors on the empty heading outright.

Two fields carry weight beyond bookkeeping. **`undermined_by`** marks a decision whose
premise a later ticket broke; without it a renderer paints the node green and launders a
live problem into a checkmark. **`claimed_at`** is what separates a dead session from live
work. And `out_of_scope` remains distinct from `resolved` — the skill keeps scope
boundaries out of Decisions-so-far because they are not steps on the route, so a parser
that lumps them in over-counts.

### Fenced code blocks are not structure

Every scan — headings, titles, bullets, links — ignores whatever sits inside a ` ``` ` or
`~~~` fence. A ticket that quotes the ticket format in its Question contains the line
`## Answer`, and must not thereby resolve itself. On a map about maps that is the likeliest
ticket anyone writes. The rule cuts one way only: an answer whose entire body is a code
fence is still an answer.

Any other reader of these files owes the same rule. A fence-blind `grep` for `^## Answer`
is a convenience, not the contract.

### One session at a time

The adapter forbids concurrent sessions: git merges a duplicate ticket number cleanly and
neither session can see it. `lint` still reports duplicates, because a merge can happen
anyway — that is the `fsck` role, after the fact.

Fog patches in the map's `## Not yet specified` are bullets with a bolded lead title and an
optional anchor — enough to give a patch identity and, when known, a position:

```markdown
- **Today's row.** Whether today is visually distinguished. <clears-with: 04>
```

Fog stays deliberately coarse. It is not promoted to one file per patch, because the skill
says to write it as loosely as the view allows.

## Checks

Errors: dangling or self `blocked_by`; `blocked_by` cycles; duplicate ticket numbers (two
parallel sessions both reaching for `10`); a ticket carrying both `## Answer` and
`## Ruled out`; a closing heading with nothing written under it; a resolved ticket
missing from Decisions-so-far; an out-of-scope ticket
listed as a decision, or missing from Out-of-scope; Decisions-so-far pointing at a ticket
that is not resolved; a fog patch duplicating a live ticket's title, or anchored to a
ticket already resolved; a missing Destination; an unknown type; a leftover `status:` field
that disagrees with the body, or that is not a status at all.

Warnings: loose (pre-frontmatter) headers; a leftover `status:` field that merely duplicates
the body; a claim with no owner or older than 72h; a claim left on a closed ticket; a ticket
blocked by something ruled out of scope, which can therefore never unblock.

Two of the skill's checks are absent here on purpose, because no parser can make them: that
a fog patch's title names a question no live ticket already holds *in substance*, and that a
decision's one-line gist says what its ticket's answer says. A third — that no progress count
is written down anywhere — ranges over the whole repository, which `lint` never reads.

Exit `0` clean, `1` errors found, `2` no map at that path.

## Reading legacy maps

Two older shapes still parse, each lints as a warning, and neither is written any more:

- the loose header (`Type:` / `Status:` / `Blocked by:` after the H1), from before frontmatter;
- a stored `status:` field, from before status was derived. Where it agrees with the body,
  `lint` says to delete it; where it disagrees, that is an error and the body wins.

## Build

```
go build ./cmd/wayfinder      # no dependencies
go test ./...
```

The module path is bare (`module wayfinder`) for local use. Give it a real path before
`go install`ing it from a remote.

## Not built yet

The GUI. The parser exists so a canvas *could* render the graph, with `undermined_by` nodes
amber and fog anchored where `clears-with` says. That is worth doing once the linter has run
clean across a couple of real efforts — and not before, because a canvas built on a format
that still drifts just renders the drift.
