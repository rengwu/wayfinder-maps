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
  // fog.js; also read by fitCamera so nebulae stay inside the fitted view.
  fogPts: []
};
