# Skills

General-purpose agent skills, adapted from [Matt Pocock's skills](https://github.com/mattpocock/skills) (MIT — see [LICENSE](./LICENSE)). Rewritten to be:

- **Standalone** — no setup skill, no per-repo configuration step. Each skill states its own conventions inline.
- **Language-agnostic** — examples are pseudocode or multi-language; tooling references defer to the project's own task runner and static checks.
- **Harness/model-agnostic** — no dependence on a specific agent product. Where a skill benefits from subagents, it degrades gracefully: run isolated parallel subagents if the environment supports them, otherwise run the same work as sequential, self-contained passes.

## Shared conventions

- **`.plan/`** — the project's planning memory, committed to version control. Specs (`.plan/<slug>/spec.md`), ticket breakdowns (`.plan/<slug>/tickets.md`), wayfinder maps (`.plan/<slug>/map.md` + `.plan/<slug>/tickets/NN-<slug>.md`), and handoffs (`.plan/handoffs/`).
- **`CONTEXT.md`** at the repo root — the domain glossary; **`docs/adr/`** — architecture decision records. Both created lazily by `domain-modeling`; skills that read them proceed silently when they don't exist.

## Skills

Five of these — `wayfinder-maps`, `grill-me`, `research`, `prototype`, `domain-modeling` — live directly in this directory, because they ship as the Claude Code plugin (wayfinder-maps plus the four skills its method invokes by name). The rest are optional extras under [`.optional/`](./.optional/) — copy over whichever you want alongside the core five.

**User-invoked** — reached by typing their name; their job is to orchestrate a session.

- **[grill-me](./grill-me/SKILL.md)** — a relentless interview to sharpen a plan or design, one question at a time.
- **[grill-with-docs](./.optional/grill-with-docs/SKILL.md)** — grill-me plus live domain-model upkeep: updates `CONTEXT.md` and ADRs as decisions crystallise.
- **[handoff](./.optional/handoff/SKILL.md)** — compact the current conversation into a handoff document (`.plan/handoffs/`) so a fresh agent can continue.
- **[improve-codebase-architecture](./.optional/improve-codebase-architecture/SKILL.md)** — scan for deepening opportunities, present as a visual HTML report, then grill through the chosen one.
- **[to-spec](./.optional/to-spec/SKILL.md)** — synthesize the current conversation into a spec at `.plan/<slug>/spec.md`.
- **[to-tickets](./.optional/to-tickets/SKILL.md)** — break a plan or spec into tracer-bullet tickets at `.plan/<slug>/tickets.md`.
- **[wayfinder-maps](./wayfinder-maps/SKILL.md)** — chart a big, foggy effort as a map of investigation tickets, resolved one per session, optionally through to implementation ([IMPLEMENT-PLAN.md](./wayfinder-maps/IMPLEMENT-PLAN.md)). Storage is adapter-specific; the default is [local markdown](./wayfinder-maps/TRACKER-MARKDOWN.md) under `.plan/<slug>/`.
- **[writing-great-skills](./.optional/writing-great-skills/SKILL.md)** — reference for writing and editing skills well.

**Model-invoked** — reachable by the agent on its own (and by other skills), or by typing their name.

- **[codebase-design](./.optional/codebase-design/SKILL.md)** — shared vocabulary and principles for designing deep modules.
- **[domain-modeling](./domain-modeling/SKILL.md)** — actively build and sharpen the project's domain model (`CONTEXT.md`, `docs/adr/`).
- **[prototype](./prototype/SKILL.md)** — throwaway code that answers a design question: an interactive terminal app for logic/state questions, or radically different UI variants behind one switcher.
- **[research](./research/SKILL.md)** — investigate a question against primary sources and capture cited findings as markdown.
- **[review-code](./.optional/review-code/SKILL.md)** — two-axis review of a diff: Standards (repo conventions + smell baseline) and Spec (does it implement what was asked?).

## Divergences from upstream

Beyond the retargeting described in [LICENSE](./LICENSE), `wayfinder-maps` makes four deliberate changes to Pocock's method. The first three exist because upstream stores tickets on an **issue tracker**, and a tracker supplies guarantees a directory of files does not.

- **A ticket is never deleted.** Upstream says "update or delete those tickets." A tracker's issue id is never reused, so deleting is unusual there and closing is the norm; a filename offers no such protection, and removing one dangles every `blocked_by` that named it. Ruled out is the way off the map. This lives in the [markdown adapter](./wayfinder-maps/TRACKER-MARKDOWN.md), not the skill — a tracker-backed adapter need not adopt it.
- **`undermined_by`.** Upstream has no equivalent: a decision whose premise a later ticket destroyed is simply `resolved`, and reads as settled. The field records what broke, so a green checkmark cannot launder a live problem.
- **Out of scope never unblocks.** Upstream says a ticket is unblocked when every ticket blocking it is *closed*, and out-of-scope tickets are closed — so a dependent goes takeable on the strength of a decision nobody made. Here `out_of_scope` satisfies no blocking edge, and a ticket blocked by one is flagged: one of the two is mis-scoped.
- **An implement phase.** Upstream stops at the decisions and hands off to `to-tickets`. Here a map can carry `implement` tickets that ship the work, usually as PRs (planned by a grilling ticket following [IMPLEMENT-PLAN.md](./wayfinder-maps/IMPLEMENT-PLAN.md), ported from `to-tickets`), a `## Implemented so far` index beside Decisions so far, per-ticket AFK/HITL on the frontier, and a hub/worker split for running AFK tickets in parallel. A map with no implement tickets behaves as upstream.

Upstream's own layering is preserved. The skill is tracker-agnostic method; storage mechanics live in an adapter, and the local-markdown adapter is the default upstream names when no tracker is wired up.
