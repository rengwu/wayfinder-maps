---
name: wayfinder-maps
description: Plan a huge chunk of work — more than one agent session can hold — as a shared map of investigation tickets, and resolve them one at a time until the way to the destination is clear.
disable-model-invocation: true
---

A loose idea has arrived — too big for one agent session, and wrapped in fog: the way from here to the **destination** isn't visible yet. Wayfinding is about finding that way, not charging at the destination. This skill charts the way as a **shared map**, then works its tickets one at a time until the route is clear.

The destination varies per effort, and naming it is the first act of charting — it shapes every ticket. It might be a spec to hand off and iterate on, a decision to lock before planning starts, or a change made in place like a data-structure migration. The map is domain-agnostic — engineering work, course content, whatever fits the shape.

## Plan, then implement

Wayfinder starts with **decisions**: most tickets resolve a question, and the pull to just do the work is usually the signal you've reached the edge of the decisions and it's time to plan the implementation. Settling the specs before implementing is encouraged, not required: not every decision can be known up front, and an implementation that raises new questions is normal.

A map can carry **implement** tickets that deliver the destination, usually as PRs, in the same graph as the decisions. A map with no implement tickets behaves exactly as it always did: done when the way is clear, nothing left to decide. Once it has any, it is done when the destination is reached: no open or claimed tickets left, not merely a clear way.

## Refer by name

Every map and ticket has a **name** — its title. In everything the human reads — narration, the map's Decisions-so-far and Implemented-so-far — refer to it by that name, never by a bare id, number, or filename. A wall of `03, 04, 05` is illegible; names read at a glance. The id doesn't vanish — a name wraps its link — but it rides *inside* the name, never stands in for it.

## The Map

The map is the canonical artifact. Its tickets hang off it, one per ticket.

The map is an **index**, not a store. It lists the decisions made (and, once implementation starts, what has shipped) and points at the tickets that hold their detail; a decision lives in exactly one place — its ticket — so the map never restates it, only gists it and links.

**Where the map, its tickets, blocking, claims, and frontier queries physically live is adapter-specific.** Consult the adapter for this repo before writing anything. Absent one, default to the local-markdown adapter: [`TRACKER-MARKDOWN.md`](TRACKER-MARKDOWN.md), which stores the map under `.plan/` and states its own shapes, invariants, and verification checklist.

Two rules keep the index honest, because a map that has drifted misleads every session that trusts it, and it does so silently:

- **The map gists a decision and links it; a ticket's own facts — its type, edges, status — are read from the ticket, never copied into prose.** A gist is lossy on purpose and cannot be derived — that is why the map reads in one pass. A fact copied has two homes, and goes stale in one of them.
- **Progress is derived from the tickets** — count them when the question is asked, wherever it is asked: a README, a contributor guide, the map itself. Never write the count down. "Four of nine resolved" is true for a week and wrong forever after, and the reader who trusts it cannot tell. A *dated* handoff is exempt where it **records** what one session did — that is history, and history never claimed to be current. The exemption stops at the record: an instruction written into a handoff ("claim it by setting…") goes stale exactly like a count, so point at the skill rather than restating it.

### The map body

The whole map at low resolution, loaded once per session. Open tickets are **not** listed here — they are found by querying the adapter.

```markdown
## Destination

<what reaching the end of this map looks like — the spec, decision, or change this effort is finding its way to. One or two lines; every session orients to it before choosing a ticket.>

## Notes

<domain; skills every session should consult; standing preferences for this effort>

## Decisions so far

<!-- the index — one line per resolved ticket: enough to judge relevance, then zoom the link for the detail the ticket holds -->

- [<resolved ticket title>](<link>) — <one-line gist of the answer>

## Implemented so far

<!-- the same index for resolved implement tickets; empty until implementation starts -->

- [<resolved implement ticket title>](<link>): <one-line gist of what shipped>

## Not yet specified

<!-- see "Fog of war": in-scope fog you can't ticket yet; graduates as the frontier advances -->

## Out of scope

<!-- see "Out of scope": work ruled beyond the destination; closed, never graduates -->
```

### Tickets

A ticket's body is the question, sized to one fresh agent session.

A session **claims** a ticket, **first**, before any work, so concurrent sessions skip it. How a claim is expressed, and how a dead session's stale claim is told apart from live work, is the adapter's business.

Blocking is a ticket's list of the tickets it waits on. A ticket is **unblocked** when every ticket it lists is `resolved`; the **frontier** is the open, unblocked, unclaimed tickets — the edge of the known. Prefer an adapter whose blocking is native, because it renders the frontier *visually* in the tracker's own interface, and the human sees what's takeable without opening the map.

`out_of_scope` is closed, and closed is not resolved: it satisfies no blocking edge. A ticket blocked by an out-of-scope ticket therefore never unblocks — one of the two is mis-scoped, and you should say which.

The answer isn't part of the body as written — it is recorded on resolution. Assets created while resolving (research notes, prototype code) are saved in the repo and linked, not pasted in.

### Undermined decisions

A resolved ticket sometimes rests on a premise that a later ticket destroys. Mark it and say what broke: record the ticket that broke it, and open its answer with a line saying so. The decision still stands — nobody has reopened it — but every session that reads the map can now see what it is standing on. A decision recorded as simply `resolved` reads as settled, and that is how a live problem gets laundered into a checkmark.

## Ticket Types

Every ticket is either **HITL** — human in the loop, worked *with* a human who speaks for themselves — or **AFK**, driven by the agent alone. A HITL ticket only resolves through that live exchange; the agent never stands in for the human's side of it (a grilling agent that answers its own questions has broken this).

- **Research** (AFK): Reading documentation, third-party APIs, or local resources like knowledge bases. Use the `research` skill: investigate against primary sources and create a cited markdown summary as a linked asset. Use when knowledge outside the current working directory is required.
- **Prototype** (HITL): Raise the fidelity of the discussion by making a cheap, rough, concrete artifact to react to — an outline, a rough take, a stub, or UI/logic code via the `prototype` skill. Links the prototype as an asset. Use when "how should it look" or "how should it behave" is the key question.
- **Grilling** (HITL): A relentless interview — one question at a time, a recommended answer offered with each, decision branches resolved one by one, facts looked up rather than asked — with the `domain-modeling` skill applied so terms and decisions are captured as they crystallise. The default case.
- **Implement** (AFK): Delivers a vertical slice of the destination in one repo, usually as one PR; work that touches several repos is several implement tickets. The body is **What to build** (the end-to-end behaviour) plus acceptance criteria. Resolved when the work has shipped (usually via a PR); the answer records what shipped and links the PR or commit when there is one. Implement tickets are not charted by hand: they come from a grilling ticket, "Plan the implementation of X", blocked by the decisions it rests on, whose resolution follows [`IMPLEMENT-PLAN.md`](IMPLEMENT-PLAN.md). A map may hold several such planning tickets, one per part of the destination.
- **Task** (HITL or AFK): Manual work that must happen before a *decision* can be made — nothing to decide, prototype, or research, but the discussion is blocked until it's done. Signing up for a service so its API can be judged, provisioning access, moving data so its shape can be seen. This is the one type that *does* rather than decides — and it earns its place by unblocking a decision, not by delivering the destination. The ticket records whether an agent can drive it alone (AFK) or it needs the human (HITL); the adapter says how. Where the agent can, it does the work; otherwise it hands the human a precise checklist. Resolved when the work is done; the answer records what was done and any resulting facts (credentials location, new URLs, row counts) later tickets depend on.

## Parallel work

The frontier says, per ticket, whether an agent can take it (AFK: research, implement, and any task marked so) or it needs the human (HITL: grilling, prototype, other tasks). A **hub** session can dispatch AFK frontier tickets to parallel workers; without a hub nothing changes, and a single session works one ticket and records everything itself.

- The hub **claims** a ticket before dispatching its worker, and owns ticket creation and the map: new tickets, Implemented-so-far and Decisions-so-far lines, fog.
- A worker writes only its own ticket: its answer and whatever the adapter says an implement ticket records. A question it raises goes in its answer; the hub turns it into a ticket or fog.
- **One ticket per session** binds the worker. The hub only claims and records, so it may record many.

Break work so independent tickets do not block each other, whatever the ticket type: edges only where one ticket genuinely gates another. Where claims, branches, and ticket numbers live is the adapter's business.

## Fog of war

The map is _deliberately_ incomplete: don't chart what you can't yet see. Beyond the live tickets lies the **fog of war** — the dim view of decisions and investigations you can tell are coming but can't yet pin down, because they hang on questions still open. Resolving a ticket clears the fog ahead of it, graduating whatever's now specifiable into fresh tickets — one at a time, until the way to the destination is clear and no tickets remain.

The map's **Not yet specified** section is where that dim view is written down: the suspected question, the area to revisit later. It's the undiscovered frontier _toward_ the destination — everything here is in scope, just not sharp enough to ticket. Write as loosely or as fully as the view allows; it doubles as a signpost for collaborators reading where the effort is headed.

Each patch carries a **title** and, where you know which open ticket will clear it, an **anchor**. A title names a question no live ticket holds; a question tracked both sharply and vaguely rots in its vague copy.

**Fog or ticket?** The test is whether you can state the question precisely now — _not_ whether you can answer it now.

- **Ticket when** the question is already sharp — even if it's blocked and you can't act on it yet.
- **Not yet specified when** you can't yet phrase it that sharply. Don't pre-slice the fog into ticket-sized pieces: it's coarser than a ticket, and one patch may graduate into several tickets, or none, once the frontier reaches it.

**Not yet specified** excludes what's already decided (Decisions so far), what's already a live ticket, and what's out of scope (the next section).

## Out of scope

Fog only ever gathers _toward_ the destination. The destination fixes the scope, so work beyond it is **out of scope** — it isn't fog, and it doesn't belong in **Not yet specified**. It gets its own **Out of scope** section on the map: work you've consciously ruled out of _this_ effort. Scope, not sharpness, lands it here.

Out-of-scope work never graduates — the frontier stops at the destination — so it returns only if the destination is redrawn, and then as a fresh effort, not a resumption.

Ruling something out of scope is a scoping act, not a step on the route. When a ticket that already exists turns out to sit past the destination — mis-scoped in while charting, or exposed by a resolution — **close it** (a closed ticket is unambiguously off the frontier) and leave one line in the map's **Out of scope** section: the gist plus why it's out of scope, linking the closed ticket. It stays out of **Decisions so far**, which records the route actually walked — a scope boundary isn't a step on it. That is why `out_of_scope` is its own state and not a flavour of `resolved`: anything counting the decisions made would otherwise count a boundary as a step.

## What stays prose

The **Destination**, the **Notes**, a ticket's **Question**, its **Answer**, and a decision's one-line gist in Decisions-so-far stay unstructured, permanently. They are lossy human summaries — nothing can derive them, and no field should try to hold them. Give the prose a schema and you have built a ticket tracker and thrown away the thing this skill is for.

## Invocation

Two modes. Either way, **never resolve more than one ticket per session.** (A hub that only claims, dispatches, and records resolves none itself; see Parallel work.)

### Chart the map

User invokes with a loose idea.

1. **Name the destination.** Run a grilling session (see the Grilling ticket type — one question at a time, recommended answers, `domain-modeling` applied) to pin down what this map is finding its way to — the spec, decision, or change. The destination fixes the scope, so it's settled first.
2. **Map the frontier.** Grill again, **breadth-first** this time: fan out across the whole space rather than deep on any one thread, surfacing the open decisions and the first steps takeable now. **If this surfaces no fog** — the way to the destination is already clear, the whole journey small enough for one session — you don't need a map. Stop and ask the user how they'd like to proceed.
3. **Create the map**: Destination and Notes filled in, Decisions-so-far empty, the fog sketched into **Not yet specified**.
4. **Create the tickets you can specify now** — then wire the blocking edges in a **second pass** (tickets need ids before they can reference each other). Wiring sorts them into the frontier and the blocked; everything you can't yet specify stays in the fog — the **Not yet specified** section.
5. **Verify** against the adapter's checklist before committing.
6. Stop — charting the map is one session's work; do not also resolve tickets.

### Work through the map

User invokes with a map. A ticket is **optional** — without one, you pick the next decision, not the user.

1. Load the **map** — the low-res view, not every ticket body.
2. Choose the ticket. If the user named one, use it. Otherwise take the first frontier ticket in order. **Claim it** before any work.
3. Resolve it — **zoom as needed**: read the full body of any related or resolved ticket on demand; invoke the skills the `## Notes` block names. If in doubt, grill (with `domain-modeling` applied).
4. Record the resolution: write the answer, release the claim, and **append a context pointer** (gist + link) to the map's Decisions-so-far, or to Implemented-so-far if the ticket was an implement ticket.
5. Add newly-surfaced tickets (create-then-wire); graduate any fog the answer has made specifiable, clearing each graduated patch from **Not yet specified** so it lives only as its new ticket. If the answer reveals a ticket — this one or another — sits beyond the destination, **rule it out of scope** rather than resolving it on the route. If the decision invalidates other parts of the map, update those tickets or close them — consult the adapter before removing one. If it breaks the premise of a decision already made, mark that ticket undermined rather than quietly re-deciding it.
6. **Verify** against the adapter's checklist before committing.

Whether two sessions may work the map at once, and how a hub and its workers split the writes, is the adapter's to say. A tracker that allocates ids atomically can permit it; a directory of files generally cannot.
