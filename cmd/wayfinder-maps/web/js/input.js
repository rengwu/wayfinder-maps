// Camera interaction: drag to pan, scroll to zoom, click to select a star.
// All input moves the camera GOAL; the render loop eases the camera toward it.
import {S, canvas} from "./state.js";
import {clamp} from "./util.js";
import {col} from "./theme.js";
import {w2s} from "./draw.js";
import {openPanel, closePanel} from "./panel.js";

var down = false, moved = 0, last = {x: 0, y: 0};

canvas.addEventListener("mousedown", function(ev) {
  down = true; moved = 0; last = {x: ev.clientX, y: ev.clientY}; canvas.classList.add("drag");
});
window.addEventListener("mousemove", function(ev) {
  if (!down) return;
  var dx = ev.clientX - last.x, dy = ev.clientY - last.y;
  S.goal.x += dx; S.goal.y += dy; moved += Math.abs(dx) + Math.abs(dy); last = {x: ev.clientX, y: ev.clientY};
});
window.addEventListener("mouseup", function(ev) {
  canvas.classList.remove("drag"); if (!down) return; down = false;
  if (moved < 5) hitTest(ev.clientX, ev.clientY);
});
canvas.addEventListener("wheel", function(ev) {
  ev.preventDefault();
  var f = Math.exp(-ev.deltaY * 0.0012); var ns = clamp(S.goal.s * f, 0.15, 4);
  var wx = (ev.clientX - S.goal.x) / S.goal.s, wy = (ev.clientY - S.goal.y) / S.goal.s;
  S.goal.s = ns; S.goal.x = ev.clientX - wx * ns; S.goal.y = ev.clientY - wy * ns;
}, {passive: false});
window.addEventListener("keydown", function(ev) { if (ev.key === "Escape") closePanel(); });

function hitTest(mx, my) {
  var best = null, bd = 1e9;
  S.nodes.forEach(function(n) {
    var s = w2s(n); var d = Math.hypot(s.x - mx, s.y - my);
    var hit = Math.max(15, col(n).r * S.cam.s + 8);
    if (d < hit && d < bd) { bd = d; best = n; }
  });
  if (best) openPanel(best); else closePanel();
}
