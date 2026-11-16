# Star-map view — design record

The map view is being redesigned from a server-rendered layered-DAG diagram into
a **2.5D, pannable star-map** with an RTS/game feel: nodes are stars in space, a
camera pans and zooms over a parallax starfield, and the graph's meaning (the
frontier, dependency direction, progress) is preserved through colour, glow, and
motion rather than a grid.

This records the decisions so a session that picks the work up cold does not
re-litigate them. Each was chosen deliberately against alternatives.

## What is kept from the current tool

The whole model layer is untouched. `wayfinder.Load(dir)` still parses `map.md`
+ `tickets/` into an `Effort`; `Layers()`, `Frontier()`, derived `Status`, and
the `serve`/`app` commands all stay. The star-map is a **new renderer** over the
same data, exactly as the first web UI was. Re-read-on-refresh stays. The
`webview_go`/cgo arrangement is unchanged.

## Decisions

1. **Tech: 2.5D hand-rolled HTML Canvas.** Camera pan (drag) + zoom (scroll), a
   multi-layer parallax starfield for depth. No engine, no JS framework, no build
   step — vanilla JS inlined by Go. Rejected true 3D/WebGL (Three.js dependency,
   build step, and a fundamentally 2D dependency graph gains only occlusion and
   depth-ambiguity from a third axis). Rejected enriched-SVG (no real camera).

2. **Layout: force-directed, rank-biased.** A physics sim spreads nodes into an
   organic constellation, but each node is soft-pulled toward a radius set by its
   dependency depth (`Layers()` rank) — roots drift inward, deeper tickets to the
   rim. Edges stay mostly monotonic so "what unblocks what" still reads. Rejected
   strict concentric rings (too clockwork, unlike the reference images) and pure
   force-directed (position would carry no meaning).

3. **Motion: deterministic + idle drift.** Positions are seeded from the ticket
   data itself (seed each node from its number, relax a fixed number of steps), so
   the same map lays out the same way every load — no persistence file, and a
   ticket stays where you learned it. Nodes bob and the starfield parallaxes for
   life; topology never reshuffles. Resolving a ticket perturbs its neighbourhood
   locally. Rejected live-physics (reshuffles every load, destroys spatial memory)
   and hard-frozen/saved positions (adds a state file that can drift).

4. **Nodes: status is the whole star; type rides in the label.** Status drives
   colour + size + glow + pulse:
   - resolved → dim blue-white ember (burned down, done)
   - frontier → bright gold, pulsing
   - claimed → amber, steady
   - blocked (open, not frontier) → small dim red
   - out-of-scope → cold gray husk
   - undermined → a cracked red halo over whatever it is
   Ticket type (grilling/prototype/research/task) appears only as a tag on the
   label and in the detail panel. Two additions ride on top of status (see
   [the implement phase](implement-phase-design.md)):
   - implement tickets are **stations**: a lit hexagon in a teal frame instead of a
     round star, laid out in an outer belt past the decision rings (a faint dotted
     orbit marks it). The belt is a hard floor: after every relaxation step a
     station inside the outermost decision's radius is pushed back out, keeping its
     angle, so edges only steer it around the ring. Status colours, glow, pulse and
     rings keep their meaning.
   - a frontier ticket carries a small badge at its upper right: a robot when an
     agent can take it alone, a person when it needs the human.

   Rejected type-as-celestial-body (richer, more
   Star-Citizen, but two visual languages to learn — deferred as a possible v2)
   and uniform stars (loses the at-a-glance status read the tool exists for).

5. **Edges: animated flow hyperlanes.** Curved lines; faint particles drift from
   blocker toward dependent, so direction is unmistakable. A **satisfied** edge
   (blocker resolved) glows solid and flows; an **unsatisfied** edge (blocker
   still open) is a faint dashed static thread. The frontier visibly ignites as
   paths clear. Rejected static gradient curves (calmer but less alive) and plain
   undirected hairlines (direction and cleared-path go dark).

6. **Interaction: in-canvas select + slide-in panel.** Click a star → it flares,
   the camera eases toward it, and a panel slides in over the map (side on
   desktop, bottom on mobile) with the ticket's title, status, blockers, and body.
   Never leave the star field. Esc / click-away deselects. The plain `/ticket`
   HTML page retires; the panel replaces it. Rejected navigate-to-page (breaks
   immersion every read) and hover-tooltip-only (can't read a ticket from the
   tool).

7. **Context: full HUD + Fog as rim nebula; decisions fold into the stars.**
   Counts + progress are an RTS resource bar (the frontier reads "agents can take"
   and "needs you"; with implement tickets the counts split into a decisions row
   and an implement row); the Destination is named in the HUD and anchors the bright core the layout radiates from. Fog patches render as
   faint nebulae drifting at the map's edge, clickable to read the patch, with a
   dashed tether to their anchor ticket when they have one. Decisions-so-far stops
   being a separate list — a resolved star *is* a decision, and its one-line gist
   shows in its panel. Rejected minimal-HUD-with-text-panels-below (staples a
   document under the map) and pure-diegetic (nothing glanceable as text).

## Architecture consequence

`/` stops serving the layered HTML diagram and serves a **canvas shell** (HTML +
inlined vanilla JS). A new data endpoint (`/graph.json`) serves the graph: nodes
(num, title, type, status, frontier flag, undermined flag, claimed-by, gist), the
blocked-by edges with a satisfied flag, fog patches (title, anchor), destination,
and counts. Each ticket's **body is inlined** into its node so the detail panel
needs no second fetch (maps are ~10–30 tickets; this is cheap). The `/ticket`
route is removed once the panel replaces it. Layout can be computed client-side
(deterministic seed) so the JSON stays pure data; the rank comes from `Layers()`.

## v1 scope — the beautiful vertical slice

Ship the look and core interaction first, then layer the "alive" motion.

**In v1:** `/graph.json`; canvas shell with pan/zoom camera + parallax starfield;
deterministic rank-biased layout; glowing status-coloured stars; static curved
edges; HUD with counts + destination; click → detail panel.

**Deferred to the immediately-following pass:** flow particles on edges, frontier
pulse, undermined cracked halo, fog nebulae + tethers, camera easing on select,
idle bob.

## Open risk

The layout algorithm — deterministic **and** rank-biased **and** force-spread
**and** non-overlapping — is the hard part and where the look lives or dies.
Expect to iterate here once it's on screen. The canvas *feel* can't be verified
from a headless session (no browser to screenshot); it's confirmed by eye via
`wayfinder-maps app <effort-dir>`. Plumbing (the JSON endpoint, the layout math) is
covered by Go tests as usual.
