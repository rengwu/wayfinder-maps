// The virtual start (left) and goal (right) of the map. Start feeds every root
// ticket; the goal is an abstract cloud that every dead-end ticket flows into
// and that holds the fog. Neither is a ticket, so neither is selectable.
import {S, ctx} from "./state.js";
import {clamp, hexA} from "./util.js";
import {colX} from "./layout.js";

var GAP = 190;

export function setupEndpoints() {
  S.startPt = S.goalPt = null;
  if (!S.nodes.length) return;
  var minx = 1e9, maxx = -1e9, ry = 0, rn = 0, nf = (S.graph && S.graph.fog ? S.graph.fog.length : 0);
  S.nodes.forEach(function(n) {
    minx = Math.min(minx, n.x); maxx = Math.max(maxx, n.x);
    if (!(n.blockers || []).length) { ry += n.y; rn++; }
  });
  var R = clamp(130 + 24 * nf, 130, 260);
  S.startPt = {x: Math.min(minx, colX(0)) - GAP, y: rn ? ry / rn : 0, r: 22};
  S.goalPt = {x: maxx + GAP + R * 0.6, y: 0, r: R};
}

function curve(ax, ay, bx, by) {
  var dx = bx - ax, dy = by - ay, len = Math.hypot(dx, dy) || 1, bow = Math.min(40, len * 0.1);
  ctx.beginPath(); ctx.moveTo(ax, ay);
  ctx.quadraticCurveTo((ax + bx) / 2 - dy / len * bow, (ay + by) / 2 + dx / len * bow, bx, by); ctx.stroke();
}

export function drawEndpoints() {
  var s = S.startPt, g = S.goalPt; if (!s || !g) return;
  var hasDep = {};
  S.edges.forEach(function(e) { hasDep[e.from] = true; });
  ctx.strokeStyle = "rgba(132,146,168,0.22)"; ctx.lineWidth = 1.1; ctx.setLineDash([2, 8]);
  S.nodes.forEach(function(n) {
    if (!(n.blockers || []).length) curve(s.x, s.y, n._x, n._y);
    if (!hasDep[n.num]) curve(n._x, n._y, g.x, g.y);
  });
  ctx.setLineDash([]);
  var breathe = 0.9 + 0.1 * Math.sin(S.clock * 0.6);
  var cg = ctx.createRadialGradient(g.x, g.y, 0, g.x, g.y, g.r * 1.15);
  cg.addColorStop(0, "rgba(255,226,150," + (0.2 * breathe) + ")");
  cg.addColorStop(0.5, "rgba(230,170,110," + (0.08 * breathe) + ")");
  cg.addColorStop(1, "rgba(230,170,110,0)");
  ctx.fillStyle = cg; ctx.beginPath(); ctx.arc(g.x, g.y, g.r * 1.15, 0, 6.2831853); ctx.fill();
  var sg = ctx.createRadialGradient(s.x, s.y, 0, s.x, s.y, s.r * 2.2);
  sg.addColorStop(0, "rgba(220,235,255,0.95)"); sg.addColorStop(0.4, "rgba(150,190,240,0.35)"); sg.addColorStop(1, "rgba(150,190,240,0)");
  ctx.fillStyle = sg; ctx.beginPath(); ctx.arc(s.x, s.y, s.r * 2.2, 0, 6.2831853); ctx.fill();
  ctx.strokeStyle = hexA("#a8c8f0", 0.6); ctx.lineWidth = 1.4;
  ctx.beginPath(); ctx.arc(s.x, s.y, s.r * 0.6, 0, 6.2831853); ctx.stroke();
}

export function drawEndpointLabels(a) {
  var s = S.startPt, g = S.goalPt; if (!s || !g || S.cam.s < 0.22 || a <= 0.002) return;
  var fs = clamp(12 * Math.pow(S.cam.s, 0.3), 9, 14);
  ctx.textAlign = "center"; ctx.font = "600 " + fs.toFixed(1) + "px ui-sans-serif,system-ui,sans-serif";
  ctx.shadowColor = "rgba(0,0,0," + (0.85 * a).toFixed(3) + ")"; ctx.shadowBlur = 4;
  ctx.fillStyle = "rgba(168,200,240," + (0.9 * a).toFixed(3) + ")";
  ctx.fillText("START", s.x * S.ec.s + S.ec.x, (s.y + s.r * 1.6) * S.ec.s + S.ec.y + fs);
  ctx.fillStyle = "rgba(255,214,150," + (0.9 * a).toFixed(3) + ")";
  var t = "GOAL" + (S.graph && S.graph.name ? ": " + S.graph.name : ""); if (t.length > 40) t = t.slice(0, 39) + "…";
  ctx.fillText(t, g.x * S.ec.s + S.ec.x, (g.y - g.r * 0.85) * S.ec.s + S.ec.y);
  ctx.shadowBlur = 0;
}
