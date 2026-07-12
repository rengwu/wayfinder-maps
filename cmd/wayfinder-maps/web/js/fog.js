// Fog patches (open questions not yet sharp enough to ticket) sit as faint
// nebulae in the void beyond the constellation, each tethered to the ticket
// that will clear it. Positions are deterministic: an anchored patch sits out
// past its ticket's angle, an unanchored one is spread by the golden angle.
import {S, ctx} from "./state.js";
import {clamp, mulberry32} from "./util.js";

export function setupFog() {
  S.fogPts = [];
  if (!S.graph || !S.graph.fog || !S.graph.fog.length || !S.nodes.length) return;
  var cx = 0, cy = 0; S.nodes.forEach(function(n) { cx += n.x; cy += n.y; }); cx /= S.nodes.length; cy /= S.nodes.length;
  var rim = 0; S.nodes.forEach(function(n) { rim = Math.max(rim, Math.hypot(n.x - cx, n.y - cy)); }); rim += 130;
  var rnd = mulberry32(4242);
  S.graph.fog.forEach(function(f, i) {
    var anc = f.clearsWith ? S.byNum[f.clearsWith] : null;
    var ang = anc ? Math.atan2(anc.y - cy, anc.x - cx) + (rnd() - 0.5) * 0.3 : (i * 2.399963 + rnd() * 0.4);
    S.fogPts.push({title: f.title, anchor: anc || null, x: cx + Math.cos(ang) * rim, y: cy + Math.sin(ang) * rim});
  });
}

export function drawFog() {
  for (var i = 0; i < S.fogPts.length; i++) {
    var f = S.fogPts[i];
    if (f.anchor) {
      ctx.strokeStyle = "rgba(150,132,205,0.28)"; ctx.lineWidth = 1; ctx.setLineDash([3, 7]); ctx.lineDashOffset = -S.clock * 4;
      ctx.beginPath(); ctx.moveTo(f.x, f.y); ctx.lineTo(f.anchor._x, f.anchor._y); ctx.stroke();
      ctx.setLineDash([]); ctx.lineDashOffset = 0;
    }
    var breathe = 0.85 + 0.18 * Math.sin(S.clock * 0.8 + f.x * 0.01), R = 92;
    var g = ctx.createRadialGradient(f.x, f.y, 0, f.x, f.y, R);
    g.addColorStop(0, "rgba(140,112,216," + (0.36 * breathe) + ")");
    g.addColorStop(0.45, "rgba(104,92,186," + (0.16 * breathe) + ")");
    g.addColorStop(1, "rgba(92,82,170,0)");
    ctx.fillStyle = g; ctx.beginPath(); ctx.arc(f.x, f.y, R, 0, 6.2831853); ctx.fill();
    ctx.fillStyle = "rgba(196,180,232,0.62)"; ctx.beginPath(); ctx.arc(f.x, f.y, 2.4, 0, 6.2831853); ctx.fill();
  }
}

export function drawFogLabels(a) {
  if (S.cam.s < 0.42 || !S.fogPts.length || a <= 0.002) return;
  var fs = clamp(10 * Math.pow(S.cam.s, 0.3), 7.5, 11.5);
  ctx.textAlign = "center"; ctx.font = "italic " + fs.toFixed(1) + "px ui-sans-serif,system-ui,sans-serif";
  ctx.shadowColor = "rgba(0,0,0," + (0.8 * a).toFixed(3) + ")"; ctx.shadowBlur = 4;
  ctx.fillStyle = "rgba(184,168,220," + (0.8 * a).toFixed(3) + ")";
  for (var i = 0; i < S.fogPts.length; i++) {
    var f = S.fogPts[i];
    var sx = f.x * S.ec.s + S.ec.x, sy = f.y * S.ec.s + S.ec.y;
    var t = f.title.length > 26 ? f.title.slice(0, 25) + "…" : f.title;
    ctx.fillText(t, sx, sy + 18);
  }
  ctx.shadowBlur = 0;
}
