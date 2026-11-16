# Implement phase: design record

Wayfinder today is planning only: a map is done when nothing is left to decide (see "Plan, don't do" in [the skill](../skills/wayfinder-maps/SKILL.md)).
This fork extends a map past the decisions into the implementation: a map can carry implement tickets that can link their PR, show which frontier tickets agents can take on their own (and so run in parallel), and render distinctly in the viewer.

Skills and viewer change together, because the viewer only reads what the format contract says.

## Why

- Real maps already override "Plan, don't do" in their Notes ("Execution is part of this map. Tickets may deliver code and data"), with no format or viewer support.
- `task` is explicitly not for delivering the destination, so build work has no honest ticket type.
- Upstream's handoff to implementation (`skills/.optional/to-tickets`) writes a flat `tickets.md` with title-based "Blocked by" that no tool reads.
- The frontier already says what can start now; it only lacks whether an agent can take a ticket alone.

## Decisions

1. **New ticket type `implement`.** Same map, same graph, same derived status. An implement ticket may be `blocked_by` the decision tickets it rests on (traceability), and decisions may be blocked by implement tickets (build, validate, learn). GitHub label: `wayfinder:implement`.
2. **Opt-in, backwards compatible.** No mode flag. A map with no implement tickets behaves exactly like today. Once it has any, the map is done when the destination is reached (no open or claimed tickets left), not merely when the way is clear.
3. **"Plan, don't do" becomes "Plan, then implement".** Settling the specs before implementing is encouraged, not required: not every decision can be known up front, and an implementation that raises new questions is normal. `task` keeps its narrow meaning: manual work that unblocks a decision.
4. **Implementation is planned in a ticket.** A grilling ticket "Plan the implementation of X", `blocked_by` the decisions it rests on, reaches the frontier like any other. Resolving it follows `skills/wayfinder-maps/IMPLEMENT-PLAN.md` and creates the implement tickets (create-then-wire). A map may hold several, one per part of the destination. No new invocation mode; one ticket per session holds.
5. **IMPLEMENT-PLAN.md** is ported from `skills/.optional/to-tickets`: vertical tracer-bullet slices, prefactor first, expand-contract for wide refactors, each ticket sized to one fresh session, quiz the user on granularity and edges before creating tickets.
6. **Implement ticket body: `## What to build`** (end-to-end behaviour) plus `- [ ]` acceptance criteria, instead of `## Question`. `## Answer` (what shipped, plus the PR or commit link when there is one) closes it as usual.
7. **AFK or HITL is readable per ticket.** Derived from type: research and implement are AFK, grilling and prototype are HITL, task is HITL unless it sets `afk: true`. `afk` is meaningful only on task. The frontier, shown per ticket as "agent can take" or "needs you", is what tells a hub which tickets can run in parallel. GitHub: label `wayfinder:afk` on a task.
8. **No waves.** Dropped: the frontier plus the AFK marker answers "what can run in parallel now", and dependency-aware breakdown stays a habit the skill encourages for every ticket type, not a computed field.
9. **Parallel work: hub claims, workers write their own ticket.**
   - The hub (coordinating session) owns ticket numbers and `map.md`. It claims a ticket on main before dispatching a worker.
   - A worker writes only its own ticket file, on its branch: `## Answer`, `pr` (when there is a PR), ticked criteria. Questions it raises go in its Answer; it creates no tickets.
   - The ticket resolves when the work has shipped (usually when the worker's PR merges), with no hub step. The hub then writes the index line and turns raised questions into tickets or fog.
   - One ticket per session binds the worker. The hub only claims and records, so it may record many.
   - Without a hub, a single session works one ticket and records everything itself, as today.
   - GitHub-backed maps get this natively (assignee = claim, atomic ids).
10. **One repo, usually one PR: optional `pr: <url>`** in implement ticket frontmatter. An implement ticket targets one repo and usually ships one PR; work that needs more repos is more implement tickets. Work without a PR (direct commit, config change, manual rollout) leaves `pr` out and links the commit in its Answer. URL only; PR state is not stored. While the PR is open, main shows the ticket `claimed` with no PR link (the link lives on the branch); no new status.
11. **Claims keep the 72h rule.** A stale implement claim is the hub's cue to check the worker or PR and refresh `claimed_at` or release it. Lint warns, never errors.
12. **Map index: new `## Implemented so far`**, the counterpart of `## Decisions so far`: one line per resolved implement ticket, gist plus link. Decisions so far never lists implement tickets and vice versa.
13. **Upstream-friendly.** Upstreaming is not a goal of this effort, but keep changes easy to offer to rengwu later: additive prose over rewrites, fork-only bits (URLs, notices) in their own commits.

## Format changes

Ticket frontmatter:

```markdown
---
type: implement
blocked_by: [04, 09]
pr: https://github.com/owner/repo/pull/12   # optional, implement only
---

# <Ticket title>

## What to build

<end-to-end behaviour>

- [ ] <acceptance criterion>
```

A task an agent can drive alone:

```markdown
---
type: task
afk: true
---
```

Map body adds, after Decisions so far:

```markdown
## Implemented so far

- [<implement ticket title>](./tickets/NN-slug.md): <one-line gist of what shipped>
```

## Skills changes

- `skills/wayfinder-maps/SKILL.md`: the `implement` ticket type (AFK, ships a PR), "Plan, then implement", the planning ticket, AFK/HITL on the frontier, and how a hub runs AFK frontier tickets in parallel.
- `skills/wayfinder-maps/TRACKER-MARKDOWN.md`: `implement` type, `pr` and `afk` fields, the What-to-build body, `Implemented so far`, hub and worker writes, checklist additions.
- `skills/wayfinder-maps/IMPLEMENT-PLAN.md`: new, see decision 5.
- `skills/README.md`: divergences from upstream; mention the implement phase.
- `README.md`: format contract, GitHub section (`wayfinder:implement`, `wayfinder:afk`, PR import), fork notice.
- `.claude-plugin/*.json`: repository URL and description point at this fork.

## Viewer and CLI changes

### Model (`internal/wayfinder`)

- `TypeImplement`; `Ticket.PR string`; `Ticket.AFK()` derived per decision 7; `Map.Implemented []Decision`.
- Lint:
  - accept `implement`
  - error: a `pr` that is not an http(s) URL, or a list
  - warn: `pr` on a non-implement ticket
  - warn: `afk` on a non-task ticket
  - warn: implement ticket listed in Decisions so far, or a non-implement ticket in Implemented so far
  - the existing index checks extended to Implemented so far

### Graph JSON (`/api/graph`)

- Node: `kind` (`decision` | `implement`), `pr`, `afk`.
- Counts split per kind; frontier split into AFK and HITL.

### Rendering

- Star-map: implement tickets get a distinct shape and hue (stations, not stars) in an outer belt past the decisions, so the whole journey is one picture. Status colours keep their meaning.
- Frontier tickets show whether an agent can take them or they need you.
- HUD: decision and implement counts separately; frontier as "agents can take" and "needs you".
- Panel: a PR chip that opens the PR, shown only when the ticket has one.

### CLI

- `status`: decisions and implementation as separate sections; each frontier ticket marked AFK or HITL, so a hub can see which can run in parallel.

### GitHub importer (`scripts/wayfinder-pull`)

- `wayfinder:implement` label maps to `type: implement`; `wayfinder:afk` maps to `afk: true`.
- Fetch `closedByPullRequestsReferences` (GraphQL) per implement issue and write `pr:` (the merged one; more than one means the ticket was mis-sliced).
- A closed implement issue with no comment gets a synthesized answer (`Shipped in <PR url>`) when a PR was found, else the usual `_Closed on GitHub with no answer comment._`; it counts as implemented either way.
- Keep it portable to macOS BSD `sed` (`sed -E`, no `\+`, no `\b`). The script currently uses both, in `slugify` and the URL rewrites.

## Work breakdown

| Ticket | Blocked by |
| --- | --- |
| A. Format contract: SKILL.md, TRACKER-MARKDOWN.md, IMPLEMENT-PLAN.md, README format section | none |
| B. Model: parse, lint, `AFK()`, graph JSON, Go tests | A |
| C. `wayfinder-pull`: labels, PR import, BSD sed | A |
| D. Rendering: stations, AFK/HITL frontier, HUD, PR chip | B |
| E. CLI: `status` split and AFK/HITL frontier | B |
| F. Fixture map exercising every case, verified in the viewer (`.claude/skills/verify`) | D, E |
| G. Release: version bump, plugin URLs, goreleaser under this fork, install docs, fork notice | D, E, C |

## Out of scope

- Offering this upstream to rengwu. Revisit after the fork has carried one real map through implementation (see decision 13).
