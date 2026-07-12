// Deterministic rank-biased force layout, plus the camera fit for a fresh map.
import {S} from "./state.js";
import {clamp, mulberry32} from "./util.js";

export function ringR(rank) { return 130 + rank * 165; }

export function layout() {
  var rnd = mulberry32(1337);
  S.nodes.forEach(function(n) {
    var ang = rnd() * 6.2831853;
    var jit = (rnd() - 0.5) * 70;
    var R = ringR(n.rank) + jit;
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
      var d = Math.hypot(n.x, n.y) || 0.01; var target = ringR(n.rank);
      var f = (target - d) * RADIAL; n.x += (n.x / d) * f; n.y += (n.y / d) * f;
    });
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
  var rnd = mulberry32(97 + fresh.length);
  fresh.forEach(function(n) {
    var ax = null, ay = null;
    S.edges.forEach(function(e) {
      if (e.to === n.num && S.byNum[e.from] && S.byNum[e.from].px != null) { ax = S.byNum[e.from].x; ay = S.byNum[e.from].y; }
      else if (e.from === n.num && S.byNum[e.to] && S.byNum[e.to].px != null) { ax = S.byNum[e.to].x; ay = S.byNum[e.to].y; }
    });
    var ang = rnd() * 6.2831853;
    if (ax != null) { n.x = ax + Math.cos(ang) * 120; n.y = ay + Math.sin(ang) * 120; }
    else { var R = ringR(n.rank); n.x = Math.cos(ang) * R; n.y = Math.sin(ang) * R; }
  });
  for (var it = 0; it < 200; it++) {
    for (var k = 0; k < fresh.length; k++) {
      var n = fresh[k], fx = 0, fy = 0, i;
      for (i = 0; i < S.nodes.length; i++) { var o = S.nodes[i]; if (o === n) continue; var dx = n.x - o.x, dy = n.y - o.y, d2 = dx * dx + dy * dy || 0.01, d = Math.sqrt(d2), f = 9000 / d2; fx += dx / d * f; fy += dy / d * f; }
      for (i = 0; i < S.edges.length; i++) { var e = S.edges[i], o2 = null; if (e.from === n.num) o2 = S.byNum[e.to]; else if (e.to === n.num) o2 = S.byNum[e.from]; if (!o2) continue; var dx2 = o2.x - n.x, dy2 = o2.y - n.y, dd2 = Math.hypot(dx2, dy2) || 0.01, f2 = (dd2 - 150) * 0.02; fx += dx2 / dd2 * f2; fy += dy2 / dd2 * f2; }
      var dd = Math.hypot(n.x, n.y) || 0.01, rf = (ringR(n.rank) - dd) * 0.05; fx += n.x / dd * rf; fy += n.y / dd * rf;
      n.x += fx; n.y += fy;
    }
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
  var topInset = 54, botInset = 72, availH = Math.max(120, innerHeight - topInset - botInset);
  var s = Math.min(innerWidth / spanx, availH / spany); s = clamp(s, 0.15, 1.4);
  S.cam.s = s;
  var cx = (minx + maxx) / 2, cy = (miny + maxy) / 2;
  S.cam.x = innerWidth / 2 - cx * s; S.cam.y = (topInset + availH / 2) - cy * s;
}
