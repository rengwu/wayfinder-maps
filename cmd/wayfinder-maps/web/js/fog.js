// Fog patches (open questions not yet sharp enough to ticket) sit as hatched,
// unsurveyed boxes beyond the tickets, each tethered to the ticket that will
// clear it. Positions are deterministic: an anchored patch sits out
// past its ticket's angle, an unanchored one is spread by the golden angle.
import {S, ctx} from "./state.js";
import {clamp, mulberry32} from "./util.js";

// R is the half-size of a fog box, in world units.
var R = 30;

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
      ctx.strokeStyle = "rgba(138,135,126,0.35)"; ctx.lineWidth = 1; ctx.setLineDash([2, 4]); ctx.lineDashOffset = -S.clock * 4;
      ctx.beginPath(); ctx.moveTo(f.x, f.y); ctx.lineTo(f.anchor._x, f.anchor._y); ctx.stroke();
      ctx.setLineDash([]); ctx.lineDashOffset = 0;
    }
    // The box: a dashed outline over diagonal hatching, clipped to the square.
    ctx.fillStyle = "rgba(14,15,16,0.9)"; ctx.fillRect(f.x - R, f.y - R, 2 * R, 2 * R);
    ctx.save(); ctx.beginPath(); ctx.rect(f.x - R, f.y - R, 2 * R, 2 * R); ctx.clip();
    ctx.strokeStyle = "rgba(138,135,126,0.12)"; ctx.lineWidth = 1; ctx.beginPath();
    for (var d = -2 * R; d < 2 * R; d += 8) { ctx.moveTo(f.x + d - R, f.y + R); ctx.lineTo(f.x + d + R, f.y - R); }
    ctx.stroke(); ctx.restore();
    ctx.strokeStyle = "rgba(138,135,126,0.32)"; ctx.setLineDash([2, 4]);
    ctx.strokeRect(f.x - R + 0.5, f.y - R + 0.5, 2 * R - 1, 2 * R - 1); ctx.setLineDash([]);
  }
}

export function drawFogLabels(a) {
  if (S.cam.s < 0.42 || !S.fogPts.length || a <= 0.002) return;
  var fs = clamp(10 * Math.pow(S.cam.s, 0.3), 7.5, 10.5);
  ctx.textAlign = "center"; ctx.font = "italic " + fs.toFixed(1) + "px 'Roboto Mono', ui-monospace, monospace";
  for (var i = 0; i < S.fogPts.length; i++) {
    var f = S.fogPts[i];
    var sx = f.x * S.ec.s + S.ec.x, sy = f.y * S.ec.s + S.ec.y;
    var t = f.title.length > 26 ? f.title.slice(0, 25) + "…" : f.title;
    var w = ctx.measureText(t).width;
    ctx.fillStyle = "rgba(14,15,16," + (0.85 * a).toFixed(3) + ")"; ctx.fillRect(sx - w / 2 - 4, sy + R * S.ec.s + 4, w + 8, fs + 6);
    ctx.fillStyle = "rgba(138,135,126," + a.toFixed(3) + ")";
    ctx.fillText(t, sx, sy + R * S.ec.s + fs + 6);
  }
}
