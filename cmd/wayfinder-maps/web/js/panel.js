// The slide-in ticket detail panel: header, status chips, rendered markdown
// body, and cross-ticket links that jump the star-map to another node.
import {S, viewRect} from "./state.js";
import {el, pad2} from "./util.js";
import {mdToHtml} from "./markdown.js";
import {mode} from "./filter.js";
import {markSelected} from "./list.js";

// edgeRow builds one "blocked by" / "blocks" line: a plain label, then one
// pill per ticket number, coloured by that ticket's status, titled with its
// name, and clicking through to it. A number whose ticket is missing from the
// graph (a dangling edge) renders as a red, inert pill.
function edgeRow(label, nums) {
  var row = el("div", "erow");
  row.appendChild(el("span", "elabel", label));
  nums.forEach(function(num) {
    var t = S.byNum[num];
    var cls = t ? (t.status === "out_of_scope" ? "oos" : t.status) : "blocked";
    var a = el("span", "c " + cls + (t ? " tl" : ""), pad2(num));
    if (t) {
      a.title = t.title;
      a.onclick = function() { openPanel(t); };
    }
    row.appendChild(a);
  });
  return row;
}

export function fillPanel(n) {
  var p = document.getElementById("panel");
  var h = p.querySelector("h2"); h.innerHTML = ""; h.appendChild(el("span", "num", pad2(n.num))); h.appendChild(document.createTextNode(n.title));
  var meta = p.querySelector(".meta"); meta.innerHTML = "";
  meta.appendChild(el("span", "c " + (n.status === "out_of_scope" ? "oos" : n.status), n.status.replace(/_/g, " ")));
  if (n.type) meta.appendChild(el("span", "c type", n.type));
  meta.appendChild(el("span", "c type", mode(n)));
  if (n.claimedBy) meta.appendChild(el("span", "c type", "by " + n.claimedBy));
  if (n.undermined) meta.appendChild(el("span", "c undermined", "undermined"));
  if ((n.blockers || []).length) meta.appendChild(edgeRow("blocked by", n.blockers));
  var blocks = S.nodes.filter(function(o) { return (o.blockers || []).indexOf(n.num) >= 0; })
                      .map(function(o) { return o.num; });
  if (blocks.length) meta.appendChild(edgeRow("blocks", blocks));
  p.querySelector(".md").innerHTML = mdToHtml(n.body);
}

export function openPanel(n) {
  S.selected = n;
  // Ease the camera so the ticket sits centred in the map area left of the panel.
  var p = document.getElementById("panel"), r = viewRect();
  var x1 = Math.min(r.x1, innerWidth - (p.offsetWidth || 0));
  S.goal.x = (r.x0 + x1) / 2 - n.x * S.goal.s; S.goal.y = (r.y0 + r.y1) / 2 - n.y * S.goal.s;
  fillPanel(n); p.classList.add("open"); document.body.classList.add("panelopen");
  markSelected();
}

export function closePanel() {
  S.selected = null;
  document.getElementById("panel").classList.remove("open");
  document.body.classList.remove("panelopen");
  markSelected();
}

document.querySelector("#panel .x").onclick = closePanel;
// Cross-ticket links in the rendered body jump the star-map to that node.
document.querySelector("#panel .md").addEventListener("click", function(ev) {
  var a = ev.target.closest ? ev.target.closest("[data-goto]") : null;
  if (a) { ev.preventDefault(); var n = S.byNum[parseInt(a.getAttribute("data-goto"), 10)]; if (n) openPanel(n); }
});
