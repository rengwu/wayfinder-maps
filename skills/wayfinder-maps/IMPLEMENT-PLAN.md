# Plan the implementation

How to resolve a "Plan the implementation of X" ticket of [`wayfinder-maps`](SKILL.md): a grilling ticket whose answer is a set of **implement tickets**, tracer-bullet vertical slices, each declaring the tickets that block it. Ported from the optional `to-tickets` skill, with the output made native to the map.

The planning ticket is resolved like any grilling ticket, one session, claimed first. Its answer lists the implement tickets it created, by name and link.

## Process

### 1. Gather context

Read the map and the decisions the planning ticket is blocked by, in full: they are the spec. If the user points at another document, read that too.

### 2. Explore the codebase

If you have not already, explore the code to understand its current state. If a `CONTEXT.md` glossary exists at the repo root, ticket titles and bodies should use its vocabulary; respect ADRs in `docs/adr/` in the area you're touching. If neither exists, proceed silently.

Look for opportunities to prefactor the code to make the implementation easier. "Make the change easy, then make the easy change."

### 3. Draft vertical slices

Break the work into **tracer bullet** tickets.

- Each slice cuts a narrow but COMPLETE path through every layer the change touches (e.g. schema, API, UI, tests): vertical, NOT a horizontal slice of one layer.
- A completed slice is demoable or verifiable on its own.
- Each slice is sized to fit one fresh agent session, because that is who takes it.
- Each slice targets one repo and lands as one PR. A change spanning several repos, or too big for one PR, is several slices, wired with edges where one must merge first.
- Any prefactoring is done first.

Give each ticket its **blocking edges**: the other tickets that must merge before it can start. A ticket with no blockers can start immediately. Add an edge only where one slice genuinely gates another. Slices with no edge between them can run in parallel, so keep independent slices free of edges.

**Wide refactors are the exception to vertical slicing.** A **wide refactor** is one mechanical change (rename a column, retype a shared symbol) whose **blast radius** fans across the whole codebase, so a single edit breaks thousands of call sites at once and no vertical slice can land green. Don't force it into a tracer bullet; sequence it as **expand-contract**. First expand: add the new form beside the old so nothing breaks. Then migrate the call sites over in batches sized by blast radius (per package, per directory), each batch its own ticket blocked by the expand, keeping the build green batch to batch because the old form still exists. Finally contract: delete the old form once no caller remains, in a ticket blocked by every migrate batch. When even the batches can't stay green alone, keep the sequence but let them share an integration branch that all block a final integrate-and-verify ticket; green is promised only there.

### 4. Quiz the user

Present the proposed breakdown as a numbered list. For each ticket, show:

- **Title**: short descriptive name
- **Blocked by**: which other tickets (if any) must merge first
- **What it delivers**: the end-to-end behaviour this ticket makes work

Ask the user:

- Does the granularity feel right? (too coarse / too fine)
- Are the blocking edges correct: does each ticket only depend on tickets that genuinely gate it?
- Should any tickets be merged or split further?

Iterate until the user approves the breakdown.

### 5. Create the tickets

Create the approved tickets as implement tickets, per the adapter, **create-then-wire**: tickets need ids before they can reference each other, so create them all, then wire the edges in a second pass.

Each ticket is `type: implement`, with a body of **What to build** and acceptance criteria (the adapter shows the shape), in place of a Question:

- **What to build**: the end-to-end behaviour this ticket makes work, from the user's perspective, not a layer-by-layer implementation list.
- **Acceptance criteria**: a `- [ ]` checklist the worker ticks as it goes.

Each ticket's `blocked_by` names the decisions it rests on (so the reason for the work stays traceable) and the slices that gate it.

Avoid specific file paths or code snippets: they go stale fast. Exception: if a prototype produced a snippet that encodes a decision more precisely than prose can (state machine, reducer, schema, type shape), inline it and note briefly that it came from a prototype. Trim to the decision-rich parts: not a working demo, just the important bits.

Then resolve the planning ticket as usual: write its answer, release the claim, and record it in Decisions so far. The new implement tickets reach Implemented so far only as each one's PR merges.

Each implement ticket is worked in its own fresh session: implement the slice, run the project's static checks and tests, review the work (with the `review-code` skill if it's available in your environment), tick off the acceptance criteria, write the answer and `pr` in the ticket, and open the PR. Independent AFK tickets on the frontier can run in parallel; see **Parallel work** in the skill.
