// Unit tests for the flow-direction transforms. Zero dependencies:
//   node --test cmd/wayfinder-maps/webtests/*.test.mjs
import { test } from "node:test";
import assert from "node:assert/strict";
import { DIRS, NONE, DEFAULT, isDir, flow, toWorld, toLocal, placeEndpoints } from "../web/js/direction.js";

const near = (a, b) => assert.ok(Math.abs(a - b) < 1e-9, `${a} != ${b}`);
const H = Math.SQRT1_2;

test("eight compass directions, E by default, plus an explicit none", () => {
  assert.deepEqual([...DIRS].sort(), ["E", "N", "NE", "NW", "S", "SE", "SW", "W"]);
  assert.equal(DEFAULT, "E");
  assert.ok(DIRS.every(isDir) && isDir(NONE));
  assert.ok(!isDir("up") && !isDir("") && !isDir(undefined));
});

test("flow points the way the graph moves on screen (y grows downward)", () => {
  assert.deepEqual(flow("E"), { x: 1, y: 0 });
  assert.deepEqual(flow("W"), { x: -1, y: 0 });
  assert.deepEqual(flow("S"), { x: 0, y: 1 });
  assert.deepEqual(flow("N"), { x: 0, y: -1 });
  assert.deepEqual(flow("SE"), { x: H, y: H });
  assert.deepEqual(flow("NW"), { x: -H, y: -H });
  assert.equal(flow(NONE), null);
});

test("E is the identity and none leaves positions alone", () => {
  assert.deepEqual(toWorld("E", 12, -7), { x: 12, y: -7 });
  assert.deepEqual(toWorld(NONE, 12, -7), { x: 12, y: -7 });
  assert.deepEqual(toLocal(NONE, 12, -7), { x: 12, y: -7 });
});

test("the rank axis lands on the flow vector", () => {
  for (const d of DIRS) {
    const w = toWorld(d, 100, 0), f = flow(d);
    near(w.x, f.x * 100);
    near(w.y, f.y * 100);
  }
});

test("S puts rank 0 at the top and deeper ranks lower", () => {
  assert.ok(toWorld("S", 190, 0).y > toWorld("S", 0, 0).y);
  assert.ok(toWorld("N", 190, 0).y < toWorld("N", 0, 0).y);
});

test("tickets in one rank spread perpendicular to the flow", () => {
  for (const d of DIRS) {
    const a = toWorld(d, 190, -60), b = toWorld(d, 190, 60), f = flow(d);
    near((b.x - a.x) * f.x + (b.y - a.y) * f.y, 0);
    near(Math.hypot(b.x - a.x, b.y - a.y), 120);
  }
});

test("toLocal inverts toWorld", () => {
  for (const d of [...DIRS, NONE]) {
    const w = toWorld(d, 31, -17), l = toLocal(d, w.x, w.y);
    near(l.x, 31);
    near(l.y, -17);
  }
});

test("start sits behind every ticket and the goal ahead, in every direction", () => {
  const local = [{ x: 0, y: -40, root: true }, { x: 0, y: 40, root: true }, { x: 190, y: 0 }, { x: 380, y: 10 }];
  for (const d of DIRS) {
    const pts = local.map((p) => ({ ...toWorld(d, p.x, p.y), root: p.root }));
    const { start, goal } = placeEndpoints(d, pts, 0);
    const f = flow(d), proj = (p) => p.x * f.x + p.y * f.y;
    assert.ok(proj(start) < Math.min(...pts.map(proj)), `${d}: start behind`);
    assert.ok(proj(goal) > Math.max(...pts.map(proj)), `${d}: goal ahead`);
  }
});

test("start sits level with the roots and the goal grows with the fog", () => {
  const pts = [{ x: 0, y: 30, root: true }, { x: 0, y: 50, root: true }, { x: 190, y: 0 }];
  const a = placeEndpoints("E", pts, 0), b = placeEndpoints("E", pts, 4);
  near(a.start.y, 40);
  assert.ok(b.goal.r > a.goal.r);
  near(a.goal.y, 0);
});
