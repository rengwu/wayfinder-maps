# Implement phase: design record

Wayfinder today is planning only: a map is done when nothing is left to decide (see "Plan, don't do" in [the skill](../skills/wayfinder-maps/SKILL.md)).
This fork extends a map past the decisions into the implementation: a map can end in an implementation plan whose tickets carry dependencies, show what can run in parallel, link their PRs, and render distinctly in the viewer.

Skills and viewer change together, because the viewer only reads what the format contract says.

## Why

- Real maps already override "Plan, don't do" in their Notes ("Execution is part of this map. Tickets may deliver code and data"), with no format or viewer support.
- `task` is explicitly not for delivering the destination, so build work has no honest ticket type.
- Upstream's handoff to implementation (`skills/.optional/to-tickets`) writes a flat `tickets.md` with title-based "Blocked by" that no tool reads.
- `Effort.Layers()` already ranks tickets by dependency depth; for implementation that rank is a parallel wave, so most of the planning logic exists.

## Decisions

1. **New ticket type `implement`.** Same map, same graph, same derived status. An implement ticket may be `blocked_by` the decision tickets it rests on (traceability), and decisions may be blocked by implement tickets (build, validate, learn). GitHub label: `wayfinder:implement`.
2. **Opt-in, backwards compatible.** No mode flag. A map with no implement tickets behaves exactly like today. Once it has any, the map is done when the destination is reached (all implement tickets resolved or ruled out), not merely when the way is clear.
3. **"Plan, don't do" becomes "Plan, then implement".** Decisions first. When the decision frontier and the fog are empty (or the user says the route is clear enough), a **Plan the implementation** session breaks the destination into implement tickets. `task` keeps its narrow meaning: manual work that unblocks a decision.
4. **Implementation planning rules** live in a new `skills/wayfinder-maps/IMPLEMENT-PLAN.md`, ported from `skills/.optional/to-tickets`: vertical tracer-bullet slices, prefactor first, expand-contract for wide refactors, each ticket sized to one fresh session, quiz the user on granularity and edges before creating tickets. Tickets are created and wired with the usual create-then-wire pass.
5. **PR links: `prs: [<url>, ...]`** in implement ticket frontmatter. URLs only; PR state is not stored. An implement ticket is resolved, as always, by an `## Answer` with prose: what shipped plus the PR link. Resolve on merge.
6. **Map index: new `## Implemented so far`** section, the counterpart of `## Decisions so far`: one line per resolved implement ticket, gist plus link. Decisions so far never lists implement tickets and vice versa.
7. **Parallel work vs "one session at a time".** The markdown adapter still allows one writer. Rule: only the coordinating (hub) session writes `.plan/`. It claims an implement ticket when it dispatches a worker, and records answer, `prs` and the index line after the merge. Workers never edit the map. GitHub-backed maps get this natively (assignee = claim, atomic ids).
8. **Waves.** A ticket's wave is its rank in the implement-only subgraph (decision blockers ignored, they are satisfied before the implementation starts or show as gates). Same wave and no edge between them means parallelizable. The implement frontier is "what workers can start now".

## Format changes

Ticket frontmatter:

```markdown
---
type: implement
blocked_by: [04, 09]
prs: [https://github.com/owner/repo/pull/12]   # optional, implement only
---
```

Map body adds, after Decisions so far:

```markdown
## Implemented so far

- [<implement ticket title>](./tickets/NN-slug.md): <one-line gist of what shipped>
```

## Skills changes

- `skills/wayfinder-maps/SKILL.md`: the `implement` ticket type (AFK, delivered by a worker session), "Plan, then implement", a third invocation mode **Plan the implementation**, and how working an implement ticket differs (PR, resolve on merge, hub records).
- `skills/wayfinder-maps/TRACKER-MARKDOWN.md`: `implement` type, `prs` field, `Implemented so far`, hub-only writes, checklist additions.
- `skills/wayfinder-maps/IMPLEMENT-PLAN.md`: new, see decision 4.
- `skills/README.md`: divergences from upstream; mention the implement phase.
- `README.md`: format contract, GitHub section (`wayfinder:implement`, PR import), fork notice.
- `.claude-plugin/*.json`: repository URL and description point at this fork.

## Viewer and CLI changes

### Model (`internal/wayfinder`)

- `TypeImplement`; `Ticket.PRs []string`; `Effort.Waves()` over the implement subgraph; `Map.Implemented []Decision`.
- Lint:
  - accept `implement`
  - error: a `prs` entry that is not an http(s) URL
  - warn: `prs` on a non-implement ticket
  - warn: resolved implement ticket without `prs`
  - warn: implement ticket listed in Decisions so far, or a non-implement ticket in Implemented so far
  - the existing index checks extended to Implemented so far

### Graph JSON (`/api/graph`)

- Node: `kind` (`decision` | `implement`), `prs`, `wave` (implement only).
- Counts split per kind, plus an implement frontier count ("parallel now").

### Rendering

- Star-map: implement tickets get a distinct shape and hue (stations, not stars) in an outer belt past the decisions, so the whole journey is one picture. Status colours keep their meaning.
- New **Implement view** toggle: left-to-right wave columns, cards with status and PR chips, decision blockers as faded gates on the left. One column is literally "what can run in parallel".
- HUD: decision and implement counts separately; "parallel now".
- Panel: PR chips that open the PR.

### CLI

- `status`: decisions and implementation as separate sections.
- New `waves <effort-dir> [--json]`: implement tickets per wave with status and PRs, for a hub deciding which workers to start.

### GitHub importer (`scripts/wayfinder-pull`)

- `wayfinder:implement` label maps to `type: implement`.
- Fetch `closedByPullRequestsReferences` (GraphQL) per implement issue and write `prs:`.
- A closed implement issue with no comment gets a synthesized answer ("Shipped in <PR url>"), so it reads as resolved.
- Keep it portable to macOS BSD `sed` (`sed -E`, no `\+`, no `\b`).

## Work breakdown

| Wave | Ticket | Blocked by |
| --- | --- | --- |
| 1 | A. Format contract: SKILL.md, TRACKER-MARKDOWN.md, IMPLEMENT-PLAN.md, README format section | none |
| 2 | B. Model: parse, lint, `Waves()`, graph JSON, Go tests | A |
| 2 | C. `wayfinder-pull`: label and PR import | A |
| 3 | D. Rendering: stations, Implement view, HUD, PR chips | B |
| 3 | E. CLI: `status` split and `waves` | B |
| 4 | F. Fixture map exercising every case, verified in the viewer (`.claude/skills/verify`) | D, E |
| 4 | G. Release: version bump, goreleaser under this fork, install docs | D, E, C |

## Open questions

- Does `wave` also need a manual override (e.g. two independent tickets that must not run at once because they touch the same files)? Leaning no: model that as a `blocked_by` edge.
- Should an implement ticket with an open (unmerged) PR read as its own status, or stay `claimed`? Leaning `claimed` plus a PR chip; no new stored state.
- Offer this upstream to rengwu once it has been used on a real map?
