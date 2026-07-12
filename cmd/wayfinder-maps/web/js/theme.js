// Status palette: star core/glow colours, core radius and glow radius.
export const COL = {
  resolved:     {core: "#b9c9e0", glow: "#5d76ad", r: 6, gr: 26},
  frontier:     {core: "#ffd873", glow: "#ffb020", r: 9, gr: 54},
  claimed:      {core: "#f0c078", glow: "#e0a44b", r: 8, gr: 40},
  blocked:      {core: "#c07a7a", glow: "#7a3b3b", r: 5, gr: 22},
  out_of_scope: {core: "#7d7789", glow: "#4a4550", r: 5, gr: 20}
};

export const LABELCOL = {
  resolved: "#9fb2cc", frontier: "#ffe6a0", claimed: "#e8c288",
  blocked: "#c49a9a", out_of_scope: "#8a8496"
};

export function col(n) { return COL[n.status] || COL.blocked; }
