// The map screen's top bar: map title and progress, the Map / Hybrid / List
// switch, the type and actionability filter chips, the search box, and the
// lint strip. The static parts live in index.html; buildToolbar refreshes the
// rest after every graph load or live update.
import {S, effectiveView} from "./state.js";
import {el} from "./util.js";
import {ACTS, TYPES, filtering, matches} from "./filter.js";
import {renderList} from "./list.js";
import {fitCamera} from "./layout.js";

var $ = function(id) { return document.getElementById(id); };

// Per-viewer conveniences: the last view and filters survive a reload. Storage
// can be missing or throw (private windows, blocked site data), so every
// access is guarded and the defaults stand in.
function load(key, dflt) {
  try { var v = localStorage.getItem("wayfinder." + key); return v ? JSON.parse(v) : dflt; } catch (e) { return dflt; }
}
function save(key, v) { try { localStorage.setItem("wayfinder." + key, JSON.stringify(v)); } catch (e) {} }

S.view = load("view", "hybrid");
var saved = load("filter", null);
if (saved) S.filter = {type: saved.type || "", act: saved.act || "", q: ""};

export function buildToolbar() {
  var g = S.graph, c = g.counts;
  $("mname").textContent = g.name;
  $("mdest").textContent = g.destination || "";
  $("progtxt").textContent = c.resolved + "/" + c.total + " resolved";
  $("progbar").style.width = (c.total ? Math.round(c.resolved * 100 / c.total) : 0) + "%";

  var tc = $("typechips"); tc.innerHTML = "";
  TYPES.forEach(function(t) {
    var all = S.nodes.filter(function(n) { return n.type === t; });
    if (!all.length) return;
    var done = all.filter(function(n) { return n.status === "resolved" || n.status === "out_of_scope"; }).length;
    var b = chip(t, done + "/" + all.length, S.filter.type === t);
    b.title = done + " of " + all.length + " " + t + " tickets closed";
    b.onclick = function() { setFilter("type", S.filter.type === t ? "" : t); };
    tc.appendChild(b);
  });

  var ac = $("actchips"); ac.innerHTML = "";
  ACTS.forEach(function(a) {
    var n = S.nodes.filter(a.test).length;
    if (!n && a.key === "oos") return; // out of scope is rare; hide it at zero
    var b = chip(a.label, String(n), S.filter.act === a.key, a.sq);
    if (!n) b.classList.add("zero");
    b.onclick = function() { setFilter("act", S.filter.act === a.key ? "" : a.key); };
    ac.appendChild(b);
  });

  var lint = g.lint || [], box = $("lint");
  box.style.display = lint.length ? "block" : "none";
  if (lint.length) {
    var errs = lint.filter(function(d) { return d.level === "error"; }).length;
    var sum = box.querySelector("summary");
    sum.className = errs ? "critical" : "warning";
    sum.textContent = (errs ? errs + " lint error" + (errs > 1 ? "s" : "") + ", " : "") +
      (lint.length - errs) + " warning" + (lint.length - errs === 1 ? "" : "s") + " in this map";
    var rows = box.querySelector(".lrows"); rows.innerHTML = "";
    lint.forEach(function(d) {
      var r = el("div", "lrow " + (d.level === "error" ? "critical" : "warning"));
      r.appendChild(el("span", "lfile", d.file)); r.appendChild(el("span", "lmsg", d.msg));
      rows.appendChild(r);
    });
  }
  refreshCounts();
}

// chip builds one filter toggle: an optional status square, a label, a count.
function chip(label, count, on, sq) {
  var b = el("button", "fchip" + (on ? " on" : ""));
  if (sq) b.appendChild(el("span", "sq " + sq));
  b.appendChild(el("span", "flabel", label));
  b.appendChild(el("span", "fcount", count));
  return b;
}

function refreshCounts() {
  var shown = S.nodes.filter(matches).length;
  $("shown").textContent = filtering() ? shown + " of " + S.nodes.length + " tickets" : S.nodes.length + " tickets";
  $("clearf").style.display = filtering() ? "" : "none";
}

export function setFilter(key, val) {
  S.filter[key] = val;
  save("filter", {type: S.filter.type, act: S.filter.act});
  if (S.graph) { buildToolbar(); renderList(); }
}

export function setView(v) {
  S.view = v; save("view", v);
  applyView();
  // The map's free area changed shape: refit so the whole map sits inside it.
  if (S.nodes.length && effectiveView() !== "list") { fitCamera(); S.goal.x = S.cam.x; S.goal.y = S.cam.y; S.goal.s = S.cam.s; }
}

// applyView puts the effective view on <body> (style.css lays the panes out
// from it) and marks the active switch button.
export function applyView() {
  var ev = effectiveView();
  document.body.classList.remove("view-map", "view-hybrid", "view-list");
  document.body.classList.add("view-" + ev);
  document.querySelectorAll("#viewseg button").forEach(function(b) {
    b.classList.toggle("on", b.getAttribute("data-view") === S.view);
  });
}

document.querySelectorAll("#viewseg button").forEach(function(b) {
  b.onclick = function() { setView(b.getAttribute("data-view")); };
});
$("search").addEventListener("input", function(ev) { setFilter("q", ev.target.value.trim()); });
$("clearf").onclick = function() {
  $("search").value = ""; S.filter.q = ""; S.filter.type = ""; setFilter("act", "");
};
