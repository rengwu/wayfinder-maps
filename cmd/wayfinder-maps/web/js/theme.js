// Palette, after alex-style's dark instrument panel. Status colours only mean
// state: frontier takes the terracotta accent because it is the work to take
// next; claimed is in progress (warning yellow), resolved is good (green),
// blocked and out-of-scope are off (grey). r is the marker's half-size and
// also feeds hit-testing in input.js.
export const UI = {
  bg: "#0e0f10", panel: "#141516", border: "#2a2c2f", grid: "#26282b", axis: "#4a4c4e",
  ink: "#e4e1d8", muted: "#8a877e", accent: "#e2683c", critical: "#ff5a4a",
  mono: "'Roboto Mono', ui-monospace, Menlo, monospace",
  sans: "Roboto, system-ui, sans-serif"
};

export const COL = {
  resolved:     {core: "#4fbf5a", r: 5},
  frontier:     {core: "#e2683c", r: 7},
  claimed:      {core: "#e6c229", r: 6.5},
  blocked:      {core: "#7a7c7e", r: 5},
  out_of_scope: {core: "#5a5c5e", r: 4}
};

export const LABELCOL = {
  resolved: "#b7b4ab", frontier: "#e4e1d8", claimed: "#e4e1d8",
  blocked: "#8a877e", out_of_scope: "#5f5d57"
};

export function col(n) { return COL[n.status] || COL.blocked; }
