// The ticket cards: every ticket that passes the filters, grouped by what can
// happen to it next. In hybrid view the list sits right of the map; hovering a
// card highlights its ticket there, and selecting either side selects both.
import {S} from "./state.js";
import {el, pad2} from "./util.js";
import {mode, matches} from "./filter.js";
import {openPanel} from "./panel.js";

// SECTIONS group by status, most actionable first.
var SECTIONS = [
  {status: "frontier",     title: "Takeable now"},
  {status: "claimed",      title: "In progress"},
  {status: "blocked",      title: "Blocked"},
  {status: "resolved",     title: "Resolved"},
  {status: "out_of_scope", title: "Out of scope"}
];

var collapsed = {};
try { collapsed = JSON.parse(localStorage.getItem("wayfinder.collapsed") || "{}"); } catch (e) {}

export function renderList() {
  var box = document.getElementById("list"), top = box.scrollTop;
  box.innerHTML = "";
  var shown = S.nodes.filter(matches), any = false;
  SECTIONS.forEach(function(sec) {
    var ns = shown.filter(function(n) { return n.status === sec.status; });
    if (!ns.length) return;
    any = true;
    var s = el("section", "lsec" + (collapsed[sec.status] ? " shut" : ""));
    var h = el("button", "lhead");
    h.appendChild(el("span", "sq " + (sec.status === "out_of_scope" ? "oos" : sec.status)));
    h.appendChild(el("span", null, sec.title));
    h.appendChild(el("span", "lcount", String(ns.length)));
    h.onclick = function() {
      collapsed[sec.status] = !collapsed[sec.status]; s.classList.toggle("shut");
      try { localStorage.setItem("wayfinder.collapsed", JSON.stringify(collapsed)); } catch (e) {}
    };
    s.appendChild(h);
    var grid = el("div", "lcards");
    ns.forEach(function(n) { grid.appendChild(card(n)); });
    s.appendChild(grid);
    box.appendChild(s);
  });
  if (!any) box.appendChild(el("div", "lempty", S.nodes.length ? "No tickets match these filters." : "This map has no tickets yet."));
  box.scrollTop = top;
}

function card(n) {
  var st = n.status === "out_of_scope" ? "oos" : n.status;
  var b = el("button", "tcard s-" + st + (S.selected === n ? " sel" : ""));
  b.setAttribute("data-num", n.num);
  var row = el("div", "trow");
  row.appendChild(el("span", "tnum", pad2(n.num)));
  if (n.type) row.appendChild(el("span", "ttype", n.type));
  row.appendChild(el("span", "tspace"));
  if (n.undermined) row.appendChild(el("span", "tund", "undermined"));
  row.appendChild(el("span", "tmode", mode(n)));
  b.appendChild(row);
  b.appendChild(el("div", "ttitle", n.title));
  b.appendChild(el("div", "tfoot", footer(n)));
  b.onmouseenter = function() { S.hover = n; };
  b.onmouseleave = function() { if (S.hover === n) S.hover = null; };
  b.onclick = function() { openPanel(n); };
  return b;
}

// footer is the one line that says what is holding a ticket or who has it.
function footer(n) {
  switch (n.status) {
    case "frontier": return "ready to claim";
    case "claimed": return "claimed by " + (n.claimedBy || "unknown") + (n.claimedAt ? " · " + ago(n.claimedAt) : "");
    case "blocked":
      var open = (n.blockers || []).filter(function(b) { var t = S.byNum[b]; return !t || t.status !== "resolved"; });
      return "waits on " + open.map(pad2).join(", ");
    case "resolved": return n.undermined && n.underminedBy ? "resolved · premise broken by " + n.underminedBy.map(pad2).join(", ") : "resolved";
    default: return "out of scope";
  }
}

function ago(iso) {
  var t = Date.parse(iso); if (isNaN(t)) return iso;
  var m = Math.max(0, Math.round((Date.now() - t) / 60000));
  if (m < 60) return m + " min ago";
  if (m < 48 * 60) return Math.round(m / 60) + " h ago";
  return Math.round(m / 1440) + " d ago";
}

// markSelected moves the card highlight to the selected ticket and scrolls its
// card into view, after a pick on the map.
export function markSelected() {
  var box = document.getElementById("list");
  box.querySelectorAll(".tcard.sel").forEach(function(c) { c.classList.remove("sel"); });
  if (!S.selected) return;
  var c = box.querySelector('.tcard[data-num="' + S.selected.num + '"]');
  if (c) { c.classList.add("sel"); c.scrollIntoView({block: "nearest"}); }
}
