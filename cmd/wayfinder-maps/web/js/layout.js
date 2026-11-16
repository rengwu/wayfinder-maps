// Deterministic rank-biased force layout, plus the camera fit for a fresh map.
import {S} from "./state.js";
import {clamp, mulberry32} from "./util.js";

export function ringR(rank) { return 130 + rank * 165; }

// Implement tickets orbit in an outer belt past the outermost decision ring.
// Maps without implement tickets never reach the belt branch, so their layout
// is unchanged.
var BELT_GAP = 160, BELT_FRAC = 0.3, BELT_STEP = 40;
var BELT_PULL = 0.2;
var belt = false, decDepth = 0, beltR = 0;
// With a belt, depth is counted within each kind: a decision that waits on an
// implement ticket would otherwise be ranked past the belt it feeds.
function rankSpan() {
  belt = S.nodes.some(function(n) { return n.kind === "implement"; });
  decDepth = 0;
  if (!belt) return;
  var memo = {};
  function depth(n) {
    if (memo[n.num] != null) return memo[n.num];
    memo[n.num] = 0;
    var d = 0, k = n.kind === "implement";
    (n.blockers || []).forEach(function(b) {
      var o = S.byNum[b];
      if (o && (o.kind === "implement") === k) d = Math.max(d, depth(o) + 1);
    });
    return (memo[n.num] = d);
  }
  S.nodes.forEach(function(n) { n._depth = depth(n); if (n.kind !== "implement") decDepth = Math.max(decDepth, n._depth); });
  beltR = ringR(decDepth);
}
// The belt is a hard floor: past the outermost decision, an implement ticket is
// pushed back out to its radius and keeps its angle, so edges only steer it around.
function clampBelt(list) {
  if (!belt) return;
  beltR = ringR(decDepth);
  S.nodes.forEach(function(n) { if (n.kind !== "implement") beltR = Math.max(beltR, Math.hypot(n.x, n.y)); });
  list.forEach(function(n) {
    if (n.kind !== "implement") return;
    var d = Math.hypot(n.x, n.y) || 0.01, R = targetR(n);
    if (d < R) { n.x *= R / d; n.y *= R / d; }
  });
}
function recentre() {
  var cx = 0, cy = 0, k = 0;
  S.nodes.forEach(function(n) { if (n.kind !== "implement") { cx += n.x; cy += n.y; k++; } });
  if (k) S.nodes.forEach(function(n) { n.x -= cx / k; n.y -= cy / k; });
}
export function targetR(n) {
  if (!belt) return ringR(n.rank);
  if (n.kind !== "implement") return ringR(n._depth);
  return beltR + Math.max(BELT_GAP, beltR * BELT_FRAC) + n._depth * BELT_STEP;
}

export function layout() {
  rankSpan();
  var rnd = mulberry32(1337);
  S.nodes.forEach(function(n) {
    var ang = rnd() * 6.2831853;
    var jit = (rnd() - 0.5) * 70;
    var R = targetR(n) + jit;
    n.x = Math.cos(ang) * R; n.y = Math.sin(ang) * R;
  });
  var REP = 9000, SPRING = 0.02, REST = 150, RADIAL = 0.05;
  for (var it = 0; it < 420; it++) {
    for (var i = 0; i < S.nodes.length; i++) {
      var a = S.nodes[i];
      for (var j = i + 1; j < S.nodes.length; j++) {
        var b = S.nodes[j];
        var dx = a.x - b.x, dy = a.y - b.y; var d2 = dx * dx + dy * dy || 0.01; var d = Math.sqrt(d2);
        var f = REP / d2; var ux = dx / d, uy = dy / d;
        a.x += ux * f; a.y += uy * f; b.x -= ux * f; b.y -= uy * f;
      }
    }
    S.edges.forEach(function(e) {
      var a = S.byNum[e.from], b = S.byNum[e.to]; if (!a || !b) return;
      var dx = b.x - a.x, dy = b.y - a.y; var d = Math.hypot(dx, dy) || 0.01;
      var f = (d - REST) * SPRING; var ux = dx / d, uy = dy / d;
      a.x += ux * f; a.y += uy * f; b.x -= ux * f; b.y -= uy * f;
    });
    S.nodes.forEach(function(n) {
      var d = Math.hypot(n.x, n.y) || 0.01; var target = targetR(n);
      var f = (target - d) * (belt && n.kind === "implement" ? BELT_PULL : RADIAL); n.x += (n.x / d) * f; n.y += (n.y / d) * f;
    });
    if (belt) recentre();
    clampBelt(S.nodes);
  }
}

// relayoutWarm places a live update without disturbing the existing map. The
// force sim has free translation/rotation modes, so re-relaxing everything slides
// the whole constellation; instead, nodes already on screen are PINNED at their
// current position (px), and only freshly-added nodes are relaxed against them as
// fixed anchors. Survivors move exactly zero; a new star flies in and settles.
export function relayoutWarm() {
  var fresh = [];
  S.nodes.forEach(function(n) { if (n.px != null) { n.x = n.px; n.y = n.py; } else fresh.push(n); });
  if (!fresh.length) return;
  rankSpan();
  var rnd = mulberry32(97 + fresh.length);
  fresh.forEach(function(n) {
    var ax = null, ay = null;
    S.edges.forEach(function(e) {
      if (e.to === n.num && S.byNum[e.from] && S.byNum[e.from].px != null) { ax = S.byNum[e.from].x; ay = S.byNum[e.from].y; }
      else if (e.from === n.num && S.byNum[e.to] && S.byNum[e.to].px != null) { ax = S.byNum[e.to].x; ay = S.byNum[e.to].y; }
    });
    var ang = rnd() * 6.2831853;
    if (ax != null) { n.x = ax + Math.cos(ang) * 120; n.y = ay + Math.sin(ang) * 120; }
    else { var R = targetR(n); n.x = Math.cos(ang) * R; n.y = Math.sin(ang) * R; }
  });
  for (var it = 0; it < 200; it++) {
    for (var k = 0; k < fresh.length; k++) {
      var n = fresh[k], fx = 0, fy = 0, i;
      for (i = 0; i < S.nodes.length; i++) { var o = S.nodes[i]; if (o === n) continue; var dx = n.x - o.x, dy = n.y - o.y, d2 = dx * dx + dy * dy || 0.01, d = Math.sqrt(d2), f = 9000 / d2; fx += dx / d * f; fy += dy / d * f; }
      for (i = 0; i < S.edges.length; i++) { var e = S.edges[i], o2 = null; if (e.from === n.num) o2 = S.byNum[e.to]; else if (e.to === n.num) o2 = S.byNum[e.from]; if (!o2) continue; var dx2 = o2.x - n.x, dy2 = o2.y - n.y, dd2 = Math.hypot(dx2, dy2) || 0.01, f2 = (dd2 - 150) * 0.02; fx += dx2 / dd2 * f2; fy += dy2 / dd2 * f2; }
      var dd = Math.hypot(n.x, n.y) || 0.01, rf = (targetR(n) - dd) * (n.kind === "implement" ? BELT_PULL : 0.05); fx += n.x / dd * rf; fy += n.y / dd * rf;
      n.x += fx; n.y += fy;
    }
    clampBelt(fresh);
  }
}

export function fitCamera() {
  if (!S.nodes.length) { S.cam = {x: innerWidth / 2, y: innerHeight / 2, s: 1}; return; }
  var minx = 1e9, miny = 1e9, maxx = -1e9, maxy = -1e9;
  S.nodes.forEach(function(n) { minx = Math.min(minx, n.x); miny = Math.min(miny, n.y); maxx = Math.max(maxx, n.x); maxy = Math.max(maxy, n.y); });
  for (var i = 0; i < S.fogPts.length; i++) { var f = S.fogPts[i]; minx = Math.min(minx, f.x - 90); miny = Math.min(miny, f.y - 90); maxx = Math.max(maxx, f.x + 90); maxy = Math.max(maxy, f.y + 90); }
  var pad = 70; minx -= pad; miny -= pad; maxx += pad; maxy += pad;
  var spanx = maxx - minx || 1, spany = maxy - miny || 1;
  // Leave the top hint and the bottom HUD bar their own bands, and fit/centre the
  // constellation into what is left so no star hides behind the chrome.
  var topInset = 54, botInset = S.nodes.some(function(n) { return n.kind === "implement"; }) ? 96 : 72, availH = Math.max(120, innerHeight - topInset - botInset);
  var s = Math.min(innerWidth / spanx, availH / spany); s = clamp(s, 0.15, 1.4);
  S.cam.s = s;
  var cx = (minx + maxx) / 2, cy = (miny + maxy) / 2;
  S.cam.x = innerWidth / 2 - cx * s; S.cam.y = (topInset + availH / 2) - cy * s;
}
