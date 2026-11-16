// Status palette: star core/glow colours, core radius and glow radius.
export const COL = {
  resolved:     {core: "#b9d6c4", glow: "#5b9077", r: 5.4, gr: 24},
  frontier:     {core: "#8ad8ff", glow: "#2f9be0", r: 8.1, gr: 49},
  claimed:      {core: "#ffd873", glow: "#ffb020", r: 7.2, gr: 36},
  blocked:      {core: "#e2c3c3", glow: "#9a6f6f", r: 4.5, gr: 20},
  out_of_scope: {core: "#7d7789", glow: "#4a4550", r: 4.5, gr: 18}
};

export const COL_IMPL = {
  resolved:     {core: "#a4d9cb", glow: "#2f8f7c", r: 6.2, gr: 25},
  frontier:     {core: "#86f5d2", glow: "#12b88f", r: 9.2, gr: 50},
  claimed:      {core: "#ffd873", glow: "#ffb020", r: 8.2, gr: 37},
  blocked:      {core: "#e2c3c3", glow: "#9a6f6f", r: 5.2, gr: 21},
  out_of_scope: {core: "#7d7789", glow: "#4a4550", r: 5.2, gr: 19}
};

export const LABELCOL = {
  resolved: "#a2c1ac", frontier: "#b3e5ff", claimed: "#ffe6a0",
  blocked: "#d0b3b3", out_of_scope: "#8a8496"
};

export const STATION = "#5fd1c0";

// Who can take a frontier ticket, as 10x10 SVG paths shared by the canvas
// (Path2D) and the HUD legend (inline SVG). Filled with the evenodd rule.
export const AFKCOL = "#8fe0cf", HITLCOL = "#e6dcc8";
export const GLYPH = {
  afk: "M4.5 0.4h1v2.6h-1zM2 3h6a1.4 1.4 0 0 1 1.4 1.4v3.6a1.4 1.4 0 0 1-1.4 1.4h-6a1.4 1.4 0 0 1-1.4-1.4v-3.6a1.4 1.4 0 0 1 1.4-1.4zM3.1 5.2a0.9 0.9 0 1 0 1.8 0a0.9 0.9 0 1 0-1.8 0zM5.1 5.2a0.9 0.9 0 1 0 1.8 0a0.9 0.9 0 1 0-1.8 0z",
  station: "M5 0.4L9 2.7V7.3L5 9.6L1 7.3V2.7zM5 2.7L7 3.85V6.15L5 7.3L3 6.15V3.85z",
  hitl: "M2.8 2.8a2.2 2.2 0 1 0 4.4 0a2.2 2.2 0 1 0-4.4 0zM1 9.6c0-2.8 1.7-4.2 4-4.2s4 1.4 4 4.2z"
};

export function col(n) { return (n.kind === "implement" ? COL_IMPL : COL)[n.status] || COL.blocked; }
