// Flow direction of the map. The layout always runs in a local frame where
// ranks advance along +x and tickets of one rank spread along y; toWorld
// rotates that frame so the flow points the chosen way. No DOM, no state.
import {clamp} from "./util.js";

export var DIRS = ["N", "NE", "E", "SE", "S", "SW", "W", "NW"];
export var NONE = "none"; // radial: rings around the centre, no start or goal
export var DEFAULT = "E";

var H = Math.SQRT1_2;
// Unit flow vectors in screen space (y grows downward).
var FLOW = {E: [1, 0], SE: [H, H], S: [0, 1], SW: [-H, H], W: [-1, 0], NW: [-H, -H], N: [0, -1], NE: [H, -H]};

export function isDir(d) { return d === NONE || Object.prototype.hasOwnProperty.call(FLOW, d); }

export function flow(dir) {
  var f = FLOW[dir];
  return f ? {x: f[0], y: f[1]} : null;
}

// Rotation, not reflection: local +x maps to the flow vector, local +y to its
// quarter turn clockwise.
export function toWorld(dir, x, y) {
  var f = FLOW[dir];
  if (!f) return {x: x, y: y};
  return {x: x * f[0] - y * f[1], y: x * f[1] + y * f[0]};
}

export function toLocal(dir, x, y) {
  var f = FLOW[dir];
  if (!f) return {x: x, y: y};
  return {x: x * f[0] + y * f[1], y: -x * f[1] + y * f[0]};
}

var GAP = 190;

// placeEndpoints puts the virtual start behind the map and the goal cloud ahead
// of it. pts are world positions of the tickets, root ones flagged; nf is the
// number of fog patches the goal has to hold. Rank 0 sits at local x = 0.
export function placeEndpoints(dir, pts, nf) {
  var minx = 0, maxx = -1e9, ry = 0, rn = 0;
  pts.forEach(function(p) {
    var l = toLocal(dir, p.x, p.y);
    minx = Math.min(minx, l.x); maxx = Math.max(maxx, l.x);
    if (p.root) { ry += l.y; rn++; }
  });
  var R = clamp(130 + 24 * nf, 130, 260);
  var s = toWorld(dir, minx - GAP, rn ? ry / rn : 0);
  var g = toWorld(dir, maxx + GAP + R * 0.6, 0);
  return {start: {x: s.x, y: s.y, r: 22}, goal: {x: g.x, y: g.y, r: R}};
}
