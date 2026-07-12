// The slide-in ticket detail panel: header, status chips, rendered markdown
// body, and cross-ticket links that jump the star-map to another node.
import {S} from "./state.js";
import {el, pad2} from "./util.js";
import {mdToHtml} from "./markdown.js";

export function fillPanel(n) {
  var p = document.getElementById("panel");
  var h = p.querySelector("h2"); h.innerHTML = ""; h.appendChild(el("span", "num", pad2(n.num))); h.appendChild(document.createTextNode(n.title));
  var meta = p.querySelector(".meta"); meta.innerHTML = "";
  meta.appendChild(el("span", "c " + (n.status === "out_of_scope" ? "oos" : n.status), n.status.replace(/_/g, " ")));
  if (n.type) meta.appendChild(el("span", "c blocked", n.type));
  (n.blockers || []).forEach(function(b) {
    var dep = S.byNum[b], ok = dep && dep.status === "resolved";
    meta.appendChild(el("span", "c " + (ok ? "resolved" : "blocked"), "blocked by " + pad2(b)));
  });
  if (n.undermined) meta.appendChild(el("span", "c blocked", "undermined"));
  p.querySelector(".md").innerHTML = mdToHtml(n.body);
}

export function openPanel(n) {
  S.selected = n;
  // Ease the camera so the star sits centred in the space left of the panel.
  var p = document.getElementById("panel"), pw = p.offsetWidth || 0, vx = (innerWidth - pw) / 2, vy = innerHeight / 2;
  S.goal.x = vx - n.x * S.goal.s; S.goal.y = vy - n.y * S.goal.s;
  fillPanel(n); p.classList.add("open"); document.body.classList.add("panelopen");
}

export function closePanel() {
  S.selected = null;
  document.getElementById("panel").classList.remove("open");
  document.body.classList.remove("panelopen");
}

document.querySelector("#panel .x").onclick = closePanel;
// Cross-ticket links in the rendered body jump the star-map to that node.
document.querySelector("#panel .md").addEventListener("click", function(ev) {
  var a = ev.target.closest ? ev.target.closest("[data-goto]") : null;
  if (a) { ev.preventDefault(); var n = S.byNum[parseInt(a.getAttribute("data-goto"), 10)]; if (n) openPanel(n); }
});
