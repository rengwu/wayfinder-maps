// Camera interaction: drag to pan, scroll to zoom, click to select a ticket.
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
  if (!down) { if (ev.target === canvas) canvas.style.cursor = nodeAt(ev.clientX, ev.clientY) ? "pointer" : ""; return; }
  var dx = ev.clientX - last.x, dy = ev.clientY - last.y;
  S.goal.x += dx; S.goal.y += dy; moved += Math.abs(dx) + Math.abs(dy); last = {x: ev.clientX, y: ev.clientY};
});
window.addEventListener("mouseup", function(ev) {
  canvas.classList.remove("drag"); if (!down) return; down = false;
  if (moved < 5) hitTest(ev.clientX, ev.clientY);
});
canvas.addEventListener("wheel", function(ev) {
  ev.preventDefault();
  var f = Math.exp(-ev.deltaY * 0.0012);
  zoomAt(ev.clientX, ev.clientY, clamp(S.goal.s * f, 0.15, 4));
}, {passive: false});
window.addEventListener("keydown", function(ev) { if (ev.key === "Escape") closePanel(); });

// --- touch: one finger pans (a short tap selects), two fingers pinch-zoom ---
// The canvas carries touch-action:none, so the browser never pans or zooms the
// page; preventDefault also stops the synthesized mouse events that would
// otherwise double-drive the handlers above.
var pinch = null; // {cx,cy,d} of the current two-finger gesture
function pinchState(ev) {
  var a = ev.touches[0], b = ev.touches[1];
  return {cx: (a.clientX + b.clientX) / 2, cy: (a.clientY + b.clientY) / 2,
          d: Math.hypot(a.clientX - b.clientX, a.clientY - b.clientY) || 1};
}
// zoomAt rescales the goal so the world point under (px,py) stays put — the
// shared core of wheel, pinch and gesture zoom.
function zoomAt(px, py, ns, nx, ny) {
  var wx = (px - S.goal.x) / S.goal.s, wy = (py - S.goal.y) / S.goal.s;
  S.goal.s = ns; S.goal.x = (nx != null ? nx : px) - wx * ns; S.goal.y = (ny != null ? ny : py) - wy * ns;
}
canvas.addEventListener("touchstart", function(ev) {
  ev.preventDefault();
  if (ev.touches.length === 1) {
    var t = ev.touches[0];
    down = true; moved = 0; last = {x: t.clientX, y: t.clientY};
  } else if (ev.touches.length === 2) {
    down = false; pinch = pinchState(ev); // second finger: drag becomes pinch, no tap
  }
}, {passive: false});
canvas.addEventListener("touchmove", function(ev) {
  ev.preventDefault();
  if (pinch && ev.touches.length === 2) {
    // Zoom about the OLD midpoint, land it on the NEW one: pinch and two-finger
    // pan are the same gesture, so both ride through zoomAt in one step.
    var np = pinchState(ev);
    zoomAt(pinch.cx, pinch.cy, clamp(S.goal.s * np.d / pinch.d, 0.15, 4), np.cx, np.cy);
    pinch = np;
  } else if (down && ev.touches.length === 1) {
    var t = ev.touches[0];
    var dx = t.clientX - last.x, dy = t.clientY - last.y;
    S.goal.x += dx; S.goal.y += dy; moved += Math.abs(dx) + Math.abs(dy); last = {x: t.clientX, y: t.clientY};
  }
}, {passive: false});
canvas.addEventListener("touchend", function(ev) {
  ev.preventDefault();
  if (ev.touches.length === 0) {
    if (pinch) { pinch = null; down = false; return; } // pinch release is never a tap
    if (down) {
      down = false;
      if (moved < 8 && ev.changedTouches.length) {
        var t = ev.changedTouches[0];
        hitTest(t.clientX, t.clientY, 24); // fingers are wider than cursors
      }
    }
  } else if (ev.touches.length === 1) {
    // Pinch collapsed to one finger: hand off to a fresh drag, not a stale one.
    pinch = null;
    var t = ev.touches[0];
    down = true; moved = 0; last = {x: t.clientX, y: t.clientY};
  }
}, {passive: false});
canvas.addEventListener("touchcancel", function() { down = false; pinch = null; });

// --- Safari trackpad pinch -------------------------------------------------
// Safari doesn't translate trackpad pinch into ctrl+wheel the way Chrome and
// Firefox do; it fires nonstandard gesture* events instead, and unhandled they
// zoom the page. ev.scale is cumulative from gesturestart, so anchor on the
// scale at start rather than compounding per event.
var gestureS0 = 1;
canvas.addEventListener("gesturestart", function(ev) { ev.preventDefault(); gestureS0 = S.goal.s; });
canvas.addEventListener("gesturechange", function(ev) {
  ev.preventDefault();
  zoomAt(ev.clientX, ev.clientY, clamp(gestureS0 * ev.scale, 0.15, 4));
});
canvas.addEventListener("gestureend", function(ev) { ev.preventDefault(); });

function hitTest(mx, my, pad) {
  var best = nodeAt(mx, my, pad);
  if (best) openPanel(best); else closePanel();
}

// nodeAt is the ticket nearest (mx,my) within its hit radius, or null.
function nodeAt(mx, my, pad) {
  var best = null, bd = 1e9, minHit = pad || 15;
  S.nodes.forEach(function(n) {
    var s = w2s(n); var d = Math.hypot(s.x - mx, s.y - my);
    var hit = Math.max(minHit, col(n).r * S.cam.s + 8);
    if (d < hit && d < bd) { bd = d; best = n; }
  });
  return best;
}
