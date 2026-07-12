// Boot: size the canvas, start the render loop and the poller, wire the
// top-level buttons, and open whatever /api/initial points at.
import {S, canvas, dpr} from "./state.js";
import {initStars, render} from "./draw.js";
import "./input.js"; // wires pan/zoom/click on the canvas
import "./panel.js"; // wires the panel close button and cross-ticket links
import {startPolling, showSplash, openProject, loadMap, leaveMap} from "./screens.js";

function resize() {
  canvas.width = innerWidth * dpr; canvas.height = innerHeight * dpr;
  canvas.style.width = innerWidth + "px"; canvas.style.height = innerHeight + "px";
}
window.addEventListener("resize", resize);

document.getElementById("openfolder").onclick = function() {
  fetch("/api/pick").then(function(r) { return r.json(); }).then(function(d) { if (d.path) openProject(d.path); });
};
document.getElementById("openanother").onclick = function() { leaveMap(showSplash); };
document.getElementById("backbtn").onclick = function() {
  leaveMap(function() { if (S.lastProject) openProject(S.lastProject); else showSplash(); });
};

resize(); initStars(); render(); startPolling();
fetch("/api/initial").then(function(r) { return r.json(); }).then(function(init) {
  if (init.effort) loadMap(init.effort);
  else if (init.project) openProject(init.project);
  else showSplash();
}).catch(showSplash);
