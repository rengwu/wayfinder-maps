// Shared mutable state for the star-map app. Everything that more than one
// module reads or reassigns lives as a field on the single S object: ES module
// bindings are read-only from the importing side, so writes go through S.
// State touched by only one module (starfield layers, drag tracking) stays
// local to that module instead.
export const canvas = document.getElementById("sky");
export const ctx = canvas.getContext("2d");
export const dpr = Math.max(1, window.devicePixelRatio || 1);

export const S = {
  cam: {x: 0, y: 0, s: 1},
  goal: {x: 0, y: 0, s: 1}, // cam eases toward goal each frame: pan, zoom, and select all move the goal
  ec: {x: 0, y: 0, s: 1},   // effective camera: cam folded with the map-transition zoom

  graph: null,
  nodes: [],
  edges: [],
  byNum: {},
  selected: null,

  clock: 0,
  lastClock: 0,

  // live-reload polling
  lastVersion: null,
  polling: false,

  screen: "splash", // splash | maplist | map
  currentEffort: null,
  lastProject: null,

  // Map-layer transition: mapAlpha (0..1) fades the constellation in/out and a
  // coupled gentle zoom (about the viewport centre) rides along, so opening,
  // leaving or switching a map dissolves over ~2s instead of snapping.
  mapAlpha: 1,
  fade: {on: false, t: 0, dur: 2.0, from: 1, to: 1, cb: null},

  // Fog patches (open questions not yet sharp enough to ticket), placed by
  // fog.js; also read by fitCamera so they stay inside the fitted view.
  fogPts: [],

  // The map screen's layout: "map" (canvas only), "hybrid" (canvas plus the
  // ticket list on the right) or "list" (cards only). See toolbar.js.
  view: "hybrid",
  // Ticket filters, applied to the list and (as dimming) to the map. type is
  // a ticket type, act an actionability bucket (filter.js), q a search string.
  filter: {type: "", act: "", q: ""},
  // The ticket under the pointer in the list, highlighted on the map.
  hover: null
};

// viewRect is the part of the window the map may use: below the toolbar and
// left of the hybrid ticket list. fitCamera and openPanel centre into it.
export function viewRect() {
  var top = document.getElementById("top"), list = document.getElementById("list");
  var y0 = shown(top) ? top.getBoundingClientRect().bottom : 0;
  var x1 = innerWidth;
  if (shown(list) && effectiveView() === "hybrid") x1 = list.getBoundingClientRect().left;
  return {x0: 0, y0: y0, x1: x1, y1: innerHeight};
}

// shown is true when el is laid out. offsetParent can't say: it is null for
// every position:fixed element.
function shown(el) { return !!el && getComputedStyle(el).display !== "none"; }

// NARROW is the window width below which hybrid has no room for both panes
// and shows the list alone. style.css hides the Hybrid button at the same width.
export const NARROW = 760;
export function effectiveView() { return S.view === "hybrid" && innerWidth < NARROW ? "list" : S.view; }
