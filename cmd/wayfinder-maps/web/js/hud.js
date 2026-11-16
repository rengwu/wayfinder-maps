// The bottom HUD bar: map title, status count chips, progress bar and legend.
import {S} from "./state.js";
import {el, glyphSvg} from "./util.js";
import {COL, GLYPH, AFKCOL, HITLCOL, STATION} from "./theme.js";

export function buildHud() {
  var hud = document.getElementById("hud"); hud.innerHTML = "";
  var title = el("div", "htitle");
  title.appendChild(el("h1", null, S.graph.name));
  if (S.graph.destination) title.appendChild(el("div", "dest", S.graph.destination));
  hud.appendChild(title);
  hud.appendChild(el("div", "spacer"));
  var c = S.graph.counts, impl = c.implement && c.implement.total > 0;
  hud.classList.toggle("hasimpl", !!impl);
  var counts = el("div", "counts"), f = frontierSplit();
  if (impl) {
    var front = el("div", "cfront");
    frontierChips(front, f);
    counts.appendChild(front);
    var kinds = el("div", "ckinds");
    kinds.appendChild(kindRow("decisions", "#9fb4d6", null, c.decisions, "decision"));
    kinds.appendChild(kindRow("implement", STATION, GLYPH.station, c.implement, "implement"));
    counts.appendChild(kinds);
  } else {
    chip(counts, "resolved", c.resolved + " resolved");
    frontierChips(counts, f);
    if (c.claimed) chip(counts, "claimed", c.claimed + " claimed");
    chip(counts, "blocked", blockedCount() + " blocked");
    if (c.outOfScope) chip(counts, "oos", c.outOfScope + " out of scope");
  }
  hud.appendChild(counts);
  var bar = el("div", "bar"); var span = document.createElement("span");
  span.style.width = (c.total ? Math.round(c.resolved * 100 / c.total) : 0) + "%"; bar.appendChild(span); hud.appendChild(bar);
  var lg = el("div", "legend");
  [["frontier", "frontier"], ["claimed", "claimed"], ["blocked", "blocked"], ["resolved", "resolved"]].forEach(function(p) {
    var b = el("b"); var d = el("span", "dot"); d.style.background = COL[p[0]].core; b.appendChild(d); b.appendChild(document.createTextNode(p[1])); lg.appendChild(b);
  });
  if (impl) legendGlyph(lg, GLYPH.station, STATION, "implement");
  hud.appendChild(lg);
}

function chip(parent, cls, txt, glyph, color) {
  var s = el("span", "c " + cls);
  if (glyph) s.appendChild(glyphSvg(glyph, color));
  s.appendChild(document.createTextNode(txt));
  parent.appendChild(s);
}

function legendGlyph(lg, path, color, txt) {
  var b = el("b"); b.appendChild(glyphSvg(path, color)); b.appendChild(document.createTextNode(txt)); lg.appendChild(b);
}

function frontierChips(parent, f) {
  if (!f.total) { chip(parent, "frontier", "0 frontier"); return; }
  if (f.afk) chip(parent, "frontier", f.afk + (f.afk === 1 ? " agent can take" : " agents can take"), GLYPH.afk, AFKCOL);
  if (f.hitl) chip(parent, "frontier", f.hitl + " needs you", GLYPH.hitl, HITLCOL);
}

function kindRow(label, color, glyph, k, kind) {
  var row = el("div", "krow"), lab = el("span", "klabel");
  if (glyph) lab.appendChild(glyphSvg(glyph, color));
  else { var d = el("span", "dot"); d.style.background = color; lab.appendChild(d); }
  lab.appendChild(document.createTextNode(label));
  row.appendChild(lab);
  chip(row, "resolved", k.resolved + "/" + k.total + " resolved");
  if (k.claimed) chip(row, "claimed", k.claimed + " claimed");
  chip(row, "blocked", blockedCount(kind) + " blocked");
  if (k.outOfScope) chip(row, "oos", k.outOfScope + " out of scope");
  return row;
}

function frontierSplit() {
  var f = {afk: 0, hitl: 0, total: 0};
  S.nodes.forEach(function(x) { if (x.status === "frontier") { f.total++; if (x.afk) f.afk++; else f.hitl++; } });
  return f;
}
function blockedCount(kind) { var n = 0; S.nodes.forEach(function(x) { if (x.status === "blocked" && (!kind || (x.kind || "decision") === kind)) n++; }); return n; }
