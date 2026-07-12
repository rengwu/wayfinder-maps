# Tracker: local markdown

The default adapter for [`wayfinder`](SKILL.md), and the one to use when the repo has no issue tracker wired up. The map and its tickets are files under `.plan/`, committed to version control — the shared memory that future sessions orient to, so commit map and ticket changes promptly.

The skill holds the method. This file holds everything the method defers: where files live, what carries structure, how status is read, what a claim is, and the checklist to run before committing. **A tool that reads a map reads it by this file.**

Structure only what has exactly one correct value **and no other home**: type, edges, claims, anchors. A machine can check those, and a second copy of one is a bug waiting to happen. A ticket's status is not among them — it is already written in the ticket, and so it is derived rather than stored. Everything else is prose — see **What stays prose** in the skill.

## One session at a time

This adapter **does not support concurrent sessions.** The skill permits them for trackers that can honour them; files in git cannot. Two sessions reach for the same ticket number, both commit, and git merges the collision cleanly — neither session could see it coming, and the loser's `blocked_by` edges now name the wrong ticket.

Serialising also earns the delta check below: if nothing merges, the invariants hold by induction from the last commit, and a session need only verify what it changed.

If a merge happens anyway, that induction is void. **Verify the whole graph, not your delta**, and run the linter if the repo has one.

## Layout

```
.plan/<effort-slug>/
  map.md                    the map body — see the skill
  tickets/NN-<slug>.md      one file per ticket, numbered from 01
  assets/                   research notes, approved markup, prototypes
```

The map's Decisions-so-far and Out-of-scope sections link tickets as `[<title>](./tickets/NN-slug.md)`.

## Tickets

A ticket's number is its identity. It is **never reused and never retired** — a ticket file is never deleted, only closed. Deleting one dangles every edge that named it, anywhere in the graph, in files no session has open; that is the single change a delta check cannot see. Where the skill says "update those tickets or close them," this adapter says: close them.

The frontmatter holds the ticket's facts, and is the only place any of them is written. The body is the question, sized to one fresh agent session.

```markdown
---
type: research | prototype | grilling | task
blocked_by: [NN, NN]                 # [] when none
claimed_by: <session or agent id>    # set while a session holds the ticket
claimed_at: <RFC 3339 timestamp>     # set alongside claimed_by
undermined_by: [NN]                  # optional — see the skill
assets: [<repo-relative path>]       # optional
---

# <Ticket title>

## Question

<the decision or investigation this ticket resolves>
```

A closing section — `## Answer`, or `## Ruled out` — is appended on closure, not written up front. Assets are saved in the repo, linked from `assets:`, and not pasted in.

### Status is derived, never stored

There is no `status:` field. Every value one could hold is already written in the file, and a field would be a second copy — the kind that goes stale in one home and lies from the other.

| Derived status | When |
|---|---|
| `resolved` | the body has an `## Answer` **with prose under it** |
| `out_of_scope` | the body has a `## Ruled out` with prose under it |
| `claimed` | neither, and `claimed_by` is set |
| `open` | none of the above |

Closure is read **first**, so a `claimed_by` left behind on a closed ticket is inert litter rather than a broken invariant — it can never hold the frontier. The one state that must not exist is both closing sections at once: a ticket is either a step on the route or a boundary of it, never both.

Writing the answer **is** the act of resolving. There is no second edit to forget, so a resolved ticket carrying no answer — and an answer sitting on an unresolved ticket — are unrepresentable rather than merely checked.

It is the *prose*, not the heading, that closes the ticket. A session that types `## Answer` and then dies has resolved nothing, and the ticket stays exactly where it was: still claimed, its claim going stale, its owner still nameable. Were the bare heading enough, that ticket would read as finished and its stale claim would look like harmless litter — a dead session laundered into a decision.

### Fenced code blocks are not structure

**Every scan for structure — headings, titles, bullets, links — ignores whatever sits inside a fenced code block** (` ``` ` or `~~~`). A ticket that quotes the ticket format in its Question contains the line `## Answer`, and must not thereby resolve itself. On a map about maps, that is the likeliest ticket you will ever write.

The rule cuts one way only: a closing section whose entire body is a code fence is still written, and still closes the ticket.

Finding the frontier is one scan. The grep below is fence-blind, so it is a convenience and not the contract — a tool must strip fences first:

```sh
grep -LE '^## (Answer|Ruled out)' tickets/*.md    # every ticket still open or claimed
```

### Claims

`claimed_by` is the claim: set it, and commit it, before any work, so a later session skips the ticket. `claimed_at` is what tells a live session apart from one that died mid-ticket; without it the frontier steps around that ticket forever.

**A claim older than 72 hours is stale.** Say so out loud when you find one, rather than silently taking or skipping the ticket.

## Fog patches

One bullet per patch in the map's **Not yet specified**. The bolded lead sentence is the patch's **title** — its identity, so it can be referred to and struck once it graduates. Anchor it to the open ticket that will clear it where you know which; leave the anchor off where you don't. Title and anchor are a patch's only structure; the rest is prose, as loose as the view allows.

```markdown
- **<Patch title>.** <prose> <clears-with: NN>
```

## Verify before you commit

The map is shared memory: the next session trusts it without re-deriving it, so drift misleads silently.

But these invariants held at the last commit, and one session touches few files. **Verify your delta, not the graph** — only what you changed can have broken them. The induction bottoms out at the charting session, which creates a map with nothing resolved and nothing closed, and it stands only because this adapter serialises sessions.

Checks 1, 2, 4 and 6 are a grep. Checks 3 and 5 need judgment, and no tool can supply it.

1. **Edges.** Every `blocked_by` you wrote names a ticket that exists, and not itself. No cycle — the whole edge set is one grep over `tickets/`.
2. **Closure.** No ticket carries both an `## Answer` and a `## Ruled out`, and no closing heading is left empty.
3. **The index.** The ticket you resolved appears exactly once in **Decisions so far**, and its gist says what its answer says. A ticket you ruled out appears once in **Out of scope**, and nowhere in Decisions-so-far.
4. **Claims.** Any ticket still carrying `claimed_by` also carries `claimed_at`, and that claim is under 72 hours old.
5. **Fog.** Every patch title names a question no live ticket holds, and every `<clears-with: NN>` names a ticket not yet resolved — a patch anchored to a resolved ticket should have graduated into a ticket, or been struck.
6. **Numbers.** Each ticket number is used once, and no ticket file was deleted.
7. **Counts.** Progress is written down nowhere; it is derived. Grep the repo for a stated count before you commit one.

Where the repo has a tool that performs these, run it — but the skill needs no tool and assumes none. A tool's job here is `fsck`, not verification: it re-establishes the base case after the things a delta check cannot see, which are edits made outside this protocol and any merge that happened despite the rule above. It runs after the fact, and its absence costs you recovery, not correctness. A tool reading only `.plan/` also cannot see check 7's grep, nor make the judgments in checks 3 and 5.
