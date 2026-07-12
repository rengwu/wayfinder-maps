// The render loop and everything painted on the canvas: parallax starfield,
// curved dependency edges, glowing status-coloured stars, decluttered labels,
// the camera easing, and the map fade-in/out transition.
import {S, ctx, dpr} from "./state.js";
import {mod, clamp, hexA, pad2, mulberry32} from "./util.js";
import {LABELCOL, col} from "./theme.js";
import {drawFog, drawFogLabels} from "./fog.js";

var EASE = 0.28;    // cam eases toward goal each frame
var POSEASE = 0.12; // node position tweening after a live update
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

// --- parallax starfield ----------------------------------------------------
var starLayers = [];
export function initStars() {
  starLayers = [];
  var specs = [{f: 0.15, n: 140, sz: 0.7, a: 0.45}, {f: 0.30, n: 80, sz: 1.1, a: 0.65}, {f: 0.50, n: 34, sz: 1.7, a: 0.9}];
  var rnd = mulberry32(9001);
  specs.forEach(function(sp) {
    var arr = [];
    for (var i = 0; i < sp.n; i++) arr.push({x: rnd(), y: rnd(), t: rnd()});
    starLayers.push({f: sp.f, sz: sp.sz, a: sp.a, stars: arr});
  });
}
function drawStars(W, H) {
  starLayers.forEach(function(L) {
    ctx.fillStyle = "rgba(255,255,255," + L.a + ")";
    for (var i = 0; i < L.stars.length; i++) {
      var s = L.stars[i];
      var x = mod(s.x * W + S.cam.x * L.f, W), y = mod(s.y * H + S.cam.y * L.f, H);
      var tw = 0.65 + 0.35 * Math.sin(s.t * 6.2831853);
      ctx.globalAlpha = L.a * tw;
      ctx.fillRect(x, y, L.sz, L.sz);
    }
  });
  ctx.globalAlpha = 1;
}

// --- nodes, edges, labels ---------------------------------------------------
// drawEdge draws a curved line from the blocker (from) to the dependent (to),
// with an arrowhead landing on the dependent so the blocking direction reads at
// a glance: from --> to means "from unblocks to".
function drawEdge(e) {
  var a = S.byNum[e.from], b = S.byNum[e.to]; if (!a || !b) return;
  var ax = a._x, ay = a._y, bx = b._x, by = b._y;
  var mx = (ax + bx) / 2, my = (ay + by) / 2, dx = bx - ax, dy = by - ay, len = Math.hypot(dx, dy) || 1;
  var nx = -dy / len, ny = dx / len, bow = Math.min(46, len * 0.13);
  var cx = mx + nx * bow, cy = my + ny * bow;
  ctx.beginPath(); ctx.moveTo(ax, ay); ctx.quadraticCurveTo(cx, cy, bx, by);
  if (e.satisfied) { ctx.strokeStyle = "rgba(160,192,166,0.62)"; ctx.lineWidth = 1.8; ctx.setLineDash([]); }
  else { ctx.strokeStyle = "rgba(132,146,168,0.34)"; ctx.lineWidth = 1.3; ctx.setLineDash([4, 6]); }
  ctx.stroke(); ctx.setLineDash([]);
  // Flow particles: on a SATISFIED edge, motes drift blocker->dependent, reading
  // as energy having unlocked the next step. Staggered phases, dim at the ends.
  if (e.satisfied) {
    for (var k = 0; k < 3; k++) {
      var u = mod(S.clock * 0.14 + k / 3 + (e.from * 0.13 + e.to * 0.07), 1);
      var m = 1 - u, fx = m * m * ax + 2 * m * u * cx + u * u * bx, fy = m * m * ay + 2 * m * u * cy + u * u * by;
      ctx.fillStyle = "rgba(190,225,200," + (0.16 + 0.44 * Math.sin(u * Math.PI)) + ")";
      ctx.beginPath(); ctx.arc(fx, fy, 1.7, 0, 6.2831853); ctx.fill();
    }
  }
  // Arrowhead at the curve MIDPOINT, pointing blocker->dependent, so direction
  // reads without crowding either vertex. For a quadratic Bezier the midpoint is
  // B(0.5)=0.25a+0.5c+0.25b and its tangent is simply b-a. The triangle is
  // centred on that point (tip half a length ahead, base half behind).
  var midx = 0.25 * ax + 0.5 * cx + 0.25 * bx, midy = 0.25 * ay + 0.5 * cy + 0.25 * by;
  var adx = bx - ax, ady = by - ay, al = Math.hypot(adx, ady) || 1, ux = adx / al, uy = ady / al;
  var ah = 7, aw = 3.8, px = -uy, py = ux;
  var tipx = midx + ux * (ah * 0.5), tipy = midy + uy * (ah * 0.5);
  ctx.beginPath();
  ctx.moveTo(tipx, tipy);
  ctx.lineTo(tipx - ux * ah + px * aw, tipy - uy * ah + py * aw);
  ctx.lineTo(tipx - ux * ah - px * aw, tipy - uy * ah - py * aw);
  ctx.closePath();
  // Opaque fill: a translucent arrow over the translucent line (or over another
  // arrow near a hub) compounds alpha and reads as two stacked shapes. The
  // satisfied/unsatisfied distinction rides on hue, not transparency.
  ctx.fillStyle = e.satisfied ? "#aecdb6" : "#6f7889";
  ctx.fill();
}

function drawNode(n) {
  var c = col(n), x = n._x, y = n._y;
  var en = (n.enter != null) ? n.enter : 1; // 0..1 grow-in for a newly-added ticket
  var fl = n.flare || 0;                    // 1..0 burst on a status change
  // Frontier stars breathe: glow and ring pulse gently so the eye is drawn to
  // what is takeable now.
  var isF = n.status === "frontier";
  var beat = 0.5 + 0.5 * Math.sin(S.clock * 2.8);
  var pulse = isF ? (0.8 + 0.2 * beat) : 1;
  var gr = (isF ? c.gr * (0.92 + 0.16 * beat) : c.gr) * (0.55 + 0.45 * en) * (1 + fl * 0.5);
  var g = ctx.createRadialGradient(x, y, 0, x, y, gr);
  g.addColorStop(0, hexA(c.glow, Math.min(1, (0.85 * pulse + fl * 0.5) * en))); g.addColorStop(0.4, hexA(c.glow, 0.22 * pulse * en)); g.addColorStop(1, hexA(c.glow, 0));
  ctx.fillStyle = g; ctx.beginPath(); ctx.arc(x, y, gr, 0, 6.2831853); ctx.fill();
  ctx.fillStyle = c.core; ctx.beginPath(); ctx.arc(x, y, c.r * (0.4 + 0.6 * en), 0, 6.2831853); ctx.fill();
  // Flare: an expanding ring pulse marking a just-changed status.
  if (fl > 0) { ctx.strokeStyle = hexA(c.core, fl * 0.7); ctx.lineWidth = 1.5 + 2 * fl; ctx.beginPath(); ctx.arc(x, y, c.r + (1 - fl) * 40, 0, 6.2831853); ctx.stroke(); }
  if (isF) { ctx.strokeStyle = hexA("#ffd873", 0.4 + 0.3 * beat); ctx.lineWidth = 1.5; ctx.beginPath(); ctx.arc(x, y, c.r + 6 + 1.5 * beat, 0, 6.2831853); ctx.stroke(); }
  // Undermined: a red cracked halo — an uneven dashed ring, its gaps travelling,
  // marking a decision resting on a premise a later ticket destroyed.
  if (n.undermined) {
    ctx.strokeStyle = hexA("#e06c75", 0.55 + 0.2 * Math.sin(S.clock * 3.5 + n.num));
    ctx.lineWidth = 1.6; ctx.setLineDash([5, 4, 2, 4]); ctx.lineDashOffset = -S.clock * 9;
    ctx.beginPath(); ctx.arc(x, y, c.r + 9, 0, 6.2831853); ctx.stroke();
    ctx.setLineDash([]); ctx.lineDashOffset = 0;
  }
  if (S.selected === n) { ctx.strokeStyle = "rgba(255,255,255,0.85)"; ctx.lineWidth = 1.5; ctx.beginPath(); ctx.arc(x, y, c.r + 11, 0, 6.2831853); ctx.stroke(); }
}

function drawLabels(a) {
  if (S.cam.s < 0.22 || a <= 0.002) return; // very far out: no text at all
  var numOnly = S.cam.s < 0.42;             // far out: ticket number only
  var fs = clamp(11 * Math.pow(S.cam.s, 0.3), 8, 13); // subtle shrink as you zoom out
  ctx.textAlign = "center"; ctx.font = fs.toFixed(1) + "px ui-sans-serif,system-ui,sans-serif";
  ctx.shadowColor = "rgba(0,0,0," + (0.85 * a).toFixed(3) + ")"; ctx.shadowBlur = 4;
  // Greedy declutter: each label prefers to sit just under its star, but if that
  // box would collide with one already placed it is nudged down (then up) until
  // it clears. Deterministic order (nodes are number-sorted) keeps it stable.
  var placed = [], h = fs + 2, step = h + 2, tries = [0, step, -step, 2 * step, -2 * step, 3 * step];
  S.nodes.forEach(function(n) {
    var s = w2s(n), c = col(n);
    var label = pad2(n.num);
    if (!numOnly) { var t = n.title.length > 30 ? n.title.slice(0, 29) + "…" : n.title; label += "  " + t; }
    var w = ctx.measureText(label).width, cx = s.x, cy = s.y + c.r * S.cam.s + fs + 3, fy = cy;
    for (var ti = 0; ti < tries.length; ti++) {
      var ty = cy + tries[ti], ok = true;
      for (var pi = 0; pi < placed.length; pi++) {
        var p = placed[pi];
        if (Math.abs(cx - p.x) < (w + p.w) / 2 + 4 && Math.abs(ty - p.y) < h + 2) { ok = false; break; }
      }
      if (ok) { fy = ty; break; }
    }
    placed.push({x: cx, y: fy, w: w});
    ctx.fillStyle = hexA(LABELCOL[n.status] || "#c8ccd6", a);
    ctx.fillText(label, cx, fy);
  });
  ctx.shadowBlur = 0;
}

export function render() {
  var W = innerWidth, H = innerHeight;
  var now = ((window.performance && performance.now ? performance.now() : Date.now()) - T0) / 1000;
  var dt = now - S.lastClock; if (dt < 0 || dt > 0.1) dt = 0.016; S.lastClock = now; S.clock = now;
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
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.fillStyle = "#05070d"; ctx.fillRect(0, 0, W, H);
  drawStars(W, H); // starfield backdrop never fades
  ctx.save(); ctx.globalAlpha = S.mapAlpha; ctx.translate(S.ec.x, S.ec.y); ctx.scale(S.ec.s, S.ec.s);
  drawFog(); S.edges.forEach(drawEdge); S.nodes.forEach(drawNode);
  ctx.restore();
  // Labels take the fade as an argument rather than through globalAlpha: WebKit's
  // canvas ignores globalAlpha on fillText once a shadow is set, so text drawn that
  // way stays fully opaque while the constellation dissolves. Baking the alpha into
  // the text and shadow colours is honoured by every engine.
  drawLabels(S.mapAlpha); drawFogLabels(S.mapAlpha);
  requestAnimationFrame(render);
}
