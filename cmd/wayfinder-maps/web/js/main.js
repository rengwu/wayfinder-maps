// Boot: size the canvas, start the render loop and the poller, wire the
// top-level buttons, and open whatever /api/initial points at (the page URL's
// ?path= wins over the path serve was started with).
import {S, canvas, dpr} from "./state.js";
import {render} from "./draw.js";
import "./input.js"; // wires pan/zoom/click on the canvas
import "./panel.js"; // wires the panel close button and cross-ticket links
import {startPolling, showSplash, openProject, loadMap, leaveMap} from "./screens.js";
import {applyView, setView} from "./toolbar.js";

function resize() {
  canvas.width = innerWidth * dpr; canvas.height = innerHeight * dpr;
  canvas.style.width = innerWidth + "px"; canvas.style.height = innerHeight + "px";
}
window.addEventListener("resize", function() { resize(); if (S.screen === "map") applyView(); });

// --top tracks the toolbar's height (chips wrap, the lint strip opens), so
// the list, panel and hint always start just below it.
var top = document.getElementById("top");
new ResizeObserver(function() {
  document.documentElement.style.setProperty("--top", top.getBoundingClientRect().height + "px");
}).observe(top);

// Keys on the map screen: / focuses the filter box, 1 2 3 pick Map, Hybrid,
// List, and Escape in the filter box clears it (instead of closing the panel,
// which input.js does otherwise).
window.addEventListener("keydown", function(ev) {
  if (S.screen !== "map" || ev.metaKey || ev.ctrlKey || ev.altKey) return;
  var search = document.getElementById("search");
  if (document.activeElement === search) {
    if (ev.key === "Escape") { search.value = ""; search.dispatchEvent(new Event("input")); search.blur(); ev.stopImmediatePropagation(); }
    return;
  }
  if (ev.key === "/") { ev.preventDefault(); search.focus(); }
  else if (ev.key === "1") setView("map");
  else if (ev.key === "2") setView("hybrid");
  else if (ev.key === "3") setView("list");
}, true);

document.getElementById("openfolder").onclick = function() {
  fetch("/api/pick").then(function(r) { return r.json(); }).then(function(d) { if (d.path) openProject(d.path); });
};
document.getElementById("openanother").onclick = function() { leaveMap(showSplash); };
document.getElementById("backbtn").onclick = function() {
  leaveMap(function() { if (S.lastProject) openProject(S.lastProject); else showSplash(); });
};

resize(); render(); startPolling();
fetch("/api/initial" + location.search).then(function(r) { return r.json(); }).then(function(init) {
  if (init.effort) loadMap(init.effort);
  else if (init.project) openProject(init.project);
  else showSplash();
}).catch(showSplash);
