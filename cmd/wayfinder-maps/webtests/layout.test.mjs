// Layout under every flow direction. The modules read the DOM at import, so
// stub just enough of it before importing them.
import { test, before } from "node:test";
import assert from "node:assert/strict";

globalThis.window = { devicePixelRatio: 1 };
globalThis.document = { getElementById: () => ({ getContext: () => ({}) }) };
globalThis.innerWidth = 1000;
globalThis.innerHeight = 700;

let S, layout, relayoutWarm, fitCamera, setupEndpoints, setupFog, D;
before(async () => {
  ({ S } = await import("../web/js/state.js"));
  ({ layout, relayoutWarm, fitCamera } = await import("../web/js/layout.js"));
  ({ setupEndpoints } = await import("../web/js/endpoints.js"));
  ({ setupFog } = await import("../web/js/fog.js"));
  D = await import("../web/js/direction.js");
});

// 1,2 roots; 3 needs 1; 4 needs 2 and 3; 5 needs 4; 6 is a root implement-free leaf.
function graph(withImplement) {
  const spec = [[1, []], [2, []], [3, [1]], [4, [2, 3]], [5, [4]], [6, []]];
  const nodes = spec.map(([num, blockers]) => ({ num, blockers, kind: "decision", rank: 0 }));
  const byNum = {};
  nodes.forEach((n) => (byNum[n.num] = n));
  const depth = (n) => (n.blockers.length ? 1 + Math.max(...n.blockers.map((b) => depth(byNum[b]))) : 0);
  nodes.forEach((n) => (n.rank = depth(n)));
  if (withImplement) {
    const impl = { num: 7, blockers: [5], kind: "implement", rank: 6 };
    nodes.push(impl);
    byNum[7] = impl;
  }
  const edges = [];
  nodes.forEach((n) => n.blockers.forEach((b) => edges.push({ from: b, to: n.num })));
  return { nodes, edges, byNum, fog: [{ title: "f", clearsWith: 5 }, { title: "g" }] };
}

function load(dir, withImplement) {
  const g = graph(withImplement);
  S.dir = dir;
  S.graph = { name: "t", fog: g.fog };
  S.nodes = g.nodes; S.edges = g.edges; S.byNum = g.byNum; S.fogPts = [];
  S.cam = { x: 0, y: 0, s: 1 };
  layout();
  S.nodes.forEach((n) => { n.px = n.x; n.py = n.y; });
  setupEndpoints();
  setupFog();
}

test("every blocker sits behind its dependent along the flow, in all 8 directions", () => {
  for (const d of D.DIRS) {
    load(d, true);
    const f = D.flow(d), proj = (p) => p.x * f.x + p.y * f.y;
    S.edges.forEach((e) => {
      assert.ok(proj(S.byNum[e.to]) > proj(S.byNum[e.from]), `${d}: ${e.from} -> ${e.to}`);
    });
  }
});

test("start is behind and the goal ahead of the whole map in every direction", () => {
  for (const d of D.DIRS) {
    load(d, false);
    const f = D.flow(d), proj = (p) => p.x * f.x + p.y * f.y;
    assert.ok(proj(S.startPt) < Math.min(...S.nodes.map(proj)), `${d} start`);
    assert.ok(proj(S.goalPt) > Math.max(...S.nodes.map(proj)), `${d} goal`);
    S.fogPts.forEach((p) => assert.ok(Math.hypot(p.x - S.goalPt.x, p.y - S.goalPt.y) <= S.goalPt.r, `${d} fog in goal`));
  }
});

test("S reads top to bottom and N bottom to top", () => {
  load("S", false);
  const top = S.byNum[1].y, bottom = S.byNum[5].y;
  assert.ok(bottom > top && S.startPt.y < top && S.goalPt.y > bottom);
  load("N", false);
  assert.ok(S.byNum[5].y < S.byNum[1].y && S.startPt.y > S.byNum[1].y && S.goalPt.y < S.byNum[5].y);
});

test("the layout is deterministic", () => {
  load("SE", true);
  const a = S.nodes.map((n) => [n.x, n.y]);
  load("SE", true);
  assert.deepEqual(S.nodes.map((n) => [n.x, n.y]), a);
});

test("a column direction yields a narrow map that fits a tall window", () => {
  load("E", false); fitCamera();
  const wide = S.cam.s;
  globalThis.innerWidth = 360; globalThis.innerHeight = 900;
  load("E", false); fitCamera();
  const eTall = S.cam.s;
  load("S", false); fitCamera();
  const sTall = S.cam.s;
  globalThis.innerWidth = 1000; globalThis.innerHeight = 700;
  assert.ok(sTall > eTall, `S (${sTall}) should fill a tall column better than E (${eTall})`);
  assert.ok(wide >= eTall);
});

test("a warm relayout keeps existing tickets put and places the new one ahead of its blocker", () => {
  for (const d of ["E", "S", "NW", D.NONE]) {
    load(d, false);
    const before = S.nodes.map((n) => [n.px, n.py]);
    const nn = { num: 8, blockers: [5], kind: "decision", rank: 5 };
    S.nodes.push(nn); S.byNum[8] = nn; S.edges.push({ from: 5, to: 8 });
    relayoutWarm();
    S.nodes.slice(0, before.length).forEach((n, i) => {
      assert.ok(Math.abs(n.x - before[i][0]) < 1e-6 && Math.abs(n.y - before[i][1]) < 1e-6, `${d}: ${n.num} moved`);
    });
    assert.ok(Number.isFinite(nn.x) && Number.isFinite(nn.y), `${d}: new ticket placed`);
    if (d !== D.NONE) {
      const f = D.flow(d), proj = (p) => p.x * f.x + p.y * f.y;
      assert.ok(proj(nn) > proj(S.byNum[5]), `${d}: new ticket ahead of its blocker`);
    }
  }
});

test("no direction restores the radial layout: rings by rank, no start or goal, fog on the rim", () => {
  load(D.NONE, false);
  assert.equal(S.startPt, null);
  assert.equal(S.goalPt, null);
  const mean = (r) => {
    const ns = S.nodes.filter((n) => n.rank === r);
    return ns.reduce((a, n) => a + Math.hypot(n.x, n.y), 0) / ns.length;
  };
  const maxRank = Math.max(...S.nodes.map((n) => n.rank));
  for (let r = 1; r <= maxRank; r++) assert.ok(mean(r) > mean(r - 1), `ring ${r} outside ring ${r - 1}`);
  const rim = Math.max(...S.nodes.map((n) => Math.hypot(n.x, n.y)));
  assert.equal(S.fogPts.length, 2);
  S.fogPts.forEach((p) => assert.ok(Math.hypot(p.x, p.y) > rim, "fog beyond the outermost ticket"));
});
