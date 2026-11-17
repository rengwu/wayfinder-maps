// Boot: size the canvas, start the render loop and the poller, wire the
// top-level buttons, and open whatever /api/initial points at.
import {S, canvas, dpr} from "./state.js";
import {initStars, render} from "./draw.js";
import "./input.js"; // wires pan/zoom/click on the canvas
import "./panel.js"; // wires the panel close button and cross-ticket links
import {startPolling, showSplash, openProject, loadMap, leaveMap, pullFromGithub, setDirection} from "./screens.js";
import {mountCompass} from "./compass.js";
import {DEFAULT, isDir} from "./direction.js";

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

document.getElementById("pullbtn").onclick = pullFromGithub;

var setCompass = mountCompass(document.getElementById("compass"), function(d) { setDirection(d); setCompass(S.dir); });
setCompass(S.dir);

resize(); initStars(); render(); startPolling();
// The saved direction has to be known before the first layout runs.
fetch("/api/settings").then(function(r) { return r.json(); }).then(function(st) {
  S.dir = isDir(st.direction) ? st.direction : DEFAULT; setCompass(S.dir);
}).catch(function() {}).then(function() { return fetch("/api/initial"); }).then(function(r) { return r.json(); }).then(function(init) {
  if (init.effort) loadMap(init.effort);
  else if (init.project) openProject(init.project);
  else showSplash();
}).catch(showSplash);
