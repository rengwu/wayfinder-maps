// The render loop and everything painted on the canvas: the survey grid,
// dependency edges, status-coloured ticket markers, decluttered labels, the
// camera easing, and the map fade-in/out transition.
import {S, ctx, dpr, effectiveView} from "./state.js";
import {mod, clamp, hexA, pad2} from "./util.js";
import {UI, LABELCOL, col} from "./theme.js";
import {drawFog, drawFogLabels} from "./fog.js";
import {matches} from "./filter.js";

// DIM is the opacity of tickets (and their edges and labels) that fail the
// current filters: still placed, so the map keeps its shape.
var DIM = 0.14;

var EASE = 0.28;    // cam eases toward goal each frame
var POSEASE = 0.12; // node position tweening after a live update
// Reduced motion: the idle animations (pulses, the claimed spinner, bobbing)
// hold still; fades and camera easing still run.
var STILL = window.matchMedia && matchMedia("(prefers-reduced-motion: reduce)").matches;
var T0 = (window.performance && performance.now ? performance.now() : Date.now());

export function startFade(from, to, dur, cb) {
  var f = S.fade;
  f.from = from; f.to = to; f.t = 0; f.dur = dur; f.cb = cb || null; f.on = true;
  S.mapAlpha = from;
}

// w2s maps a node's DISPLAY position (base + idle bob, set each frame) to screen.
// It reads the effective camera ec (base cam folded with the transition zoom)
// so nodes, labels and hit-testing all stay aligned mid-transition.
export function w2s(n) {
  var x = (n._x != null ? n._x : n.x), y = (n._y != null ? n._y : n.y);
  return {x: x * S.ec.s + S.ec.x, y: y * S.ec.s + S.ec.y};
}

// --- survey grid --------------------------------------------------------------
// A dot grid tied to the camera, with a dashed major line every fifth dot. It
// sits in screen space so it stays crisp at any zoom; the spacing jumps by 5x
// when it would get denser than 18 px.
function drawGrid(W, H) {
  var step = 40;
  while (step * S.ec.s < 18) step *= 5;
  var px = step * S.ec.s, ox = mod(S.ec.x, px), oy = mod(S.ec.y, px);
  ctx.fillStyle = "#2c2e31";
  for (var x = ox; x < W; x += px) for (var y = oy; y < H; y += px) ctx.fillRect(Math.round(x) - 0.75, Math.round(y) - 0.75, 1.5, 1.5);
  var big = px * 5, bx = mod(S.ec.x, big), by = mod(S.ec.y, big);
  ctx.strokeStyle = UI.grid; ctx.lineWidth = 1; ctx.setLineDash([2, 4]); ctx.beginPath();
  for (x = bx; x < W; x += big) { ctx.moveTo(Math.round(x) + 0.5, 0); ctx.lineTo(Math.round(x) + 0.5, H); }
  for (y = by; y < H; y += big) { ctx.moveTo(0, Math.round(y) + 0.5); ctx.lineTo(W, Math.round(y) + 0.5); }
  ctx.stroke(); ctx.setLineDash([]);
}

// --- nodes, edges, labels ---------------------------------------------------
// drawEdge draws a straight line from the blocker (from) to the dependent (to),
// with an arrowhead at the midpoint so the direction reads without crowding
// either end: from --> to means "from unblocks to". A satisfied edge is solid
// and carries small square pulses travelling blocker->dependent; an open one
// is dashed.
function drawEdge(e) {
  var a = S.byNum[e.from], b = S.byNum[e.to]; if (!a || !b) return;
  ctx.globalAlpha = S.mapAlpha * (matches(a) && matches(b) ? 1 : DIM);
  var ax = a._x, ay = a._y, bx = b._x, by = b._y;
  var dx = bx - ax, dy = by - ay, len = Math.hypot(dx, dy) || 1, ux = dx / len, uy = dy / len;
  ctx.beginPath(); ctx.moveTo(ax, ay); ctx.lineTo(bx, by);
  if (e.satisfied) { ctx.strokeStyle = "rgba(228,225,216,0.45)"; ctx.lineWidth = 1.5; ctx.setLineDash([]); }
  else { ctx.strokeStyle = UI.axis; ctx.lineWidth = 1.2; ctx.setLineDash([2, 4]); }
  ctx.stroke(); ctx.setLineDash([]);
  if (e.satisfied) {
    for (var k = 0; k < 2; k++) {
      var u = mod(S.clock * 0.10 + k / 2 + (e.from * 0.13 + e.to * 0.07), 1);
      ctx.fillStyle = "rgba(228,225,216," + (0.12 + 0.5 * Math.sin(u * Math.PI)).toFixed(3) + ")";
      ctx.fillRect(ax + dx * u - 1.5, ay + dy * u - 1.5, 3, 3);
    }
  }
  var midx = (ax + bx) / 2, midy = (ay + by) / 2, ah = 7, aw = 3.8, px = -uy, py = ux;
  var tipx = midx + ux * (ah * 0.5), tipy = midy + uy * (ah * 0.5);
  ctx.beginPath();
  ctx.moveTo(tipx, tipy);
  ctx.lineTo(tipx - ux * ah + px * aw, tipy - uy * ah + py * aw);
  ctx.lineTo(tipx - ux * ah - px * aw, tipy - uy * ah - py * aw);
  ctx.closePath();
  ctx.fillStyle = e.satisfied ? "#9a978e" : "#5a5c5e";
  ctx.fill();
}

function square(x, y, h) { ctx.rect(x - h, y - h, 2 * h, 2 * h); }

// brackets strokes the four corners of a square of half-size h, each arm k long.
function brackets(x, y, h, k) {
  ctx.beginPath();
  [[-1, -1], [1, -1], [1, 1], [-1, 1]].forEach(function(c) {
    var cx = x + c[0] * h, cy = y + c[1] * h;
    ctx.moveTo(cx - c[0] * k, cy); ctx.lineTo(cx, cy); ctx.lineTo(cx, cy - c[1] * k);
  });
  ctx.stroke();
}

function drawNode(n) {
  var c = col(n), x = n._x, y = n._y;
  ctx.globalAlpha = S.mapAlpha * (matches(n) ? 1 : DIM);
  var en = (n.enter != null) ? n.enter : 1; // 0..1 grow-in for a newly-added ticket
  var fl = n.flare || 0;                    // 1..0 burst on a status change
  var beat = 0.5 + 0.5 * Math.sin(S.clock * 2.8);
  var h = c.r * (0.4 + 0.6 * en);
  // Frontier is the work to take next: a soft terracotta glow breathes behind
  // it and an outline square pulses around it.
  if (n.status === "frontier") {
    var gr = 26 + 6 * beat, g = ctx.createRadialGradient(x, y, 0, x, y, gr);
    g.addColorStop(0, hexA(c.core, 0.32 * en)); g.addColorStop(1, hexA(c.core, 0));
    ctx.fillStyle = g; ctx.fillRect(x - gr, y - gr, 2 * gr, 2 * gr);
    ctx.strokeStyle = hexA(c.core, 0.35 + 0.4 * beat); ctx.lineWidth = 1;
    ctx.beginPath(); square(x, y, h + 4 + 2 * beat); ctx.stroke();
  }
  if (n.status === "blocked" || n.status === "out_of_scope") {
    ctx.fillStyle = UI.bg; ctx.strokeStyle = c.core; ctx.lineWidth = 1.5;
    if (n.status === "out_of_scope") ctx.setLineDash([2, 2]);
    ctx.beginPath(); square(x, y, h); ctx.fill(); ctx.stroke(); ctx.setLineDash([]);
  } else {
    ctx.fillStyle = c.core; ctx.beginPath(); square(x, y, h); ctx.fill();
  }
  // Claimed means someone is working it, so it gets a spinner: a dashed ring
  // turning around the marker.
  if (n.status === "claimed") {
    ctx.save(); ctx.translate(x, y); ctx.rotate(S.clock * 1.4);
    ctx.strokeStyle = hexA(c.core, 0.8); ctx.lineWidth = 1.5; ctx.setLineDash([6, 5]);
    ctx.beginPath(); ctx.arc(0, 0, h + 7, 0, 6.2831853); ctx.stroke();
    ctx.restore(); ctx.setLineDash([]);
  }
  // Flare: an expanding square outline marking a just-changed status.
  if (fl > 0) { ctx.strokeStyle = hexA(c.core, fl * 0.7); ctx.lineWidth = 1 + 2 * fl; ctx.beginPath(); square(x, y, h + (1 - fl) * 36); ctx.stroke(); }
  // Undermined: a critical-red dashed square, its dashes travelling, marking a
  // decision resting on a premise a later ticket broke.
  if (n.undermined) {
    ctx.strokeStyle = UI.critical; ctx.lineWidth = 1.5; ctx.setLineDash([4, 3]); ctx.lineDashOffset = -S.clock * 9;
    ctx.beginPath(); square(x, y, h + 6); ctx.stroke();
    ctx.setLineDash([]); ctx.lineDashOffset = 0;
  }
  if (S.selected === n) { ctx.strokeStyle = UI.accent; ctx.lineWidth = 2; brackets(x, y, h + 12, 6); }
  else if (S.hover === n) { ctx.strokeStyle = UI.ink; ctx.lineWidth = 1.5; brackets(x, y, h + 12, 6); }
}

// drawLabels sets each ticket's label on a dark plate under its marker: the
// number in mono ink, the title in Roboto and a status-tinted grey.
function drawLabels(a) {
  if (S.cam.s < 0.22 || a <= 0.002) return; // very far out: no text at all
  var numOnly = S.cam.s < 0.42;             // far out: ticket number only
  var fs = clamp(11 * Math.pow(S.cam.s, 0.3), 8, 12);
  ctx.textAlign = "left"; ctx.textBaseline = "alphabetic";
  // Greedy declutter: each label prefers to sit just under its marker, but if
  // that box would collide with one already placed it is nudged down (then up)
  // until it clears. Deterministic order (nodes are number-sorted) keeps it stable.
  var placed = [], h = fs + 4, step = h + 2, tries = [0, step, -step, 2 * step, -2 * step, 3 * step];
  var a0 = a;
  S.nodes.forEach(function(n) {
    var s = w2s(n), c = col(n);
    a = a0 * (matches(n) ? 1 : DIM);
    var num = pad2(n.num), title = "";
    if (!numOnly) title = "  " + (n.title.length > 30 ? n.title.slice(0, 29) + "…" : n.title);
    ctx.font = "500 " + fs.toFixed(1) + "px " + UI.mono;
    var nw = ctx.measureText(num).width;
    ctx.font = fs.toFixed(1) + "px " + UI.sans;
    var w = nw + (title ? ctx.measureText(title).width : 0);
    var cx = s.x, cy = s.y + c.r * S.cam.s + fs + 6, fy = cy;
    for (var ti = 0; ti < tries.length; ti++) {
      var ty = cy + tries[ti], ok = true;
      for (var pi = 0; pi < placed.length; pi++) {
        var p = placed[pi];
        if (Math.abs(cx - p.x) < (w + p.w) / 2 + 6 && Math.abs(ty - p.y) < h + 2) { ok = false; break; }
      }
      if (ok) { fy = ty; break; }
    }
    placed.push({x: cx, y: fy, w: w});
    var lx = cx - w / 2;
    ctx.fillStyle = hexA(UI.bg, 0.82 * a); ctx.fillRect(lx - 4, fy - fs - 1, w + 8, fs + 6);
    ctx.font = "500 " + fs.toFixed(1) + "px " + UI.mono;
    ctx.fillStyle = hexA(UI.ink, a); ctx.fillText(num, lx, fy);
    if (title) {
      ctx.font = fs.toFixed(1) + "px " + UI.sans;
      ctx.fillStyle = hexA(LABELCOL[n.status] || UI.muted, a); ctx.fillText(title, lx + nw, fy);
    }
  });
}

export function render() {
  var W = innerWidth, H = innerHeight;
  var now = ((window.performance && performance.now ? performance.now() : Date.now()) - T0) / 1000;
  var dt = now - S.lastClock; if (dt < 0 || dt > 0.1) dt = 0.016; S.lastClock = now; S.clock = STILL ? 0 : now;
  // Per node: ease the render base (px,py) toward the true layout position so a
  // structural update tweens rather than jumps; add idle bob on top; decay the
  // status-change flare and ramp a new node's grow-in. Edges/labels/hit-testing
  // all read _x/_y, so nothing detaches.
  for (var i = 0; i < S.nodes.length; i++) {
    var n = S.nodes[i], ph = n.num * 1.7;
    if (n.px == null) { n.px = n.x; n.py = n.y; }
    n.px += (n.x - n.px) * POSEASE; n.py += (n.y - n.py) * POSEASE;
    n._x = n.px + Math.sin(S.clock * 0.7 + ph) * 2.4; n._y = n.py + Math.cos(S.clock * 0.55 + ph) * 2.4;
    if (n.flare > 0) n.flare = Math.max(0, n.flare - dt / 1.1);
    if (n.enter != null && n.enter < 1) n.enter = Math.min(1, n.enter + dt / 0.55);
  }
  // Camera eases toward its goal every frame — pan, zoom and select all just
  // move the goal, giving weighty motion instead of instant jumps.
  S.cam.x += (S.goal.x - S.cam.x) * EASE; S.cam.y += (S.goal.y - S.cam.y) * EASE; S.cam.s += (S.goal.s - S.cam.s) * EASE;
  // Advance the load/unload fade (smoothstep), then derive the effective camera:
  // a gentle zoom coupled to mapAlpha, taken about the viewport centre.
  if (S.fade.on) {
    S.fade.t += dt; var u = clamp(S.fade.t / S.fade.dur, 0, 1); var e = u * u * (3 - 2 * u);
    S.mapAlpha = S.fade.from + (S.fade.to - S.fade.from) * e;
    if (u >= 1) { S.fade.on = false; if (S.fade.cb) { var cb = S.fade.cb; S.fade.cb = null; cb(); } }
  }
  var zt = 0.95 + 0.05 * S.mapAlpha, Cx = W / 2, Cy = H / 2;
  S.ec.s = S.cam.s * zt; S.ec.x = Cx + (S.cam.x - Cx) * zt; S.ec.y = Cy + (S.cam.y - Cy) * zt;
  // List view hides the canvas; keep the clock and camera running but skip
  // the painting.
  if (effectiveView() === "list" && S.screen === "map") { requestAnimationFrame(render); return; }
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.fillStyle = UI.bg; ctx.fillRect(0, 0, W, H);
  drawGrid(W, H); // the grid never fades
  ctx.save(); ctx.globalAlpha = S.mapAlpha; ctx.translate(S.ec.x, S.ec.y); ctx.scale(S.ec.s, S.ec.s);
  drawFog(); S.edges.forEach(drawEdge); S.nodes.forEach(drawNode);
  ctx.restore();
  // Labels take the fade as an argument and bake it into their colours, so text
  // fades with the map in every engine.
  drawLabels(S.mapAlpha); drawFogLabels(S.mapAlpha);
  requestAnimationFrame(render);
}
