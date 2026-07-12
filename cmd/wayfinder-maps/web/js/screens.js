// The three screens (splash, map list, star-map), the transitions between
// them, graph loading, and the live-reload poller.
import {S} from "./state.js";
import {el} from "./util.js";
import {layout, relayoutWarm, fitCamera} from "./layout.js";
import {setupFog} from "./fog.js";
import {buildHud} from "./hud.js";
import {fillPanel, closePanel} from "./panel.js";
import {startFade} from "./draw.js";

function sortNodes(g) { return (g.nodes || []).slice().sort(function(a, b) { return a.num - b.num; }); }

// applyGraph loads a graph cold: fresh node objects (no px) so layout runs from
// scratch, then fit the camera. Used when a map is first opened or switched to.
function applyGraph(g) {
  S.graph = g; S.nodes = sortNodes(g); S.edges = g.edges || []; S.byNum = {}; S.nodes.forEach(function(n) { S.byNum[n.num] = n; });
  layout(); S.nodes.forEach(function(n) { n.px = n.x; n.py = n.y; n._x = n.x; n._y = n.y; }); setupFog(); fitCamera();
  S.goal.x = S.cam.x; S.goal.y = S.cam.y; S.goal.s = S.cam.s; buildHud();
}

// updateGraph folds a freshly-fetched graph into the live scene: surviving nodes
// are mutated in place (so selection and screen position persist), status changes
// fire a flare, new tickets grow in, and layout re-runs with a warm start so
// positions tween. The camera is deliberately left alone.
function updateGraph(ng) {
  var incoming = sortNodes(ng), keep = [];
  incoming.forEach(function(nn) {
    var old = S.byNum[nn.num];
    if (old) {
      if (old.status !== nn.status || old.undermined !== nn.undermined) old.flare = 1;
      old.status = nn.status; old.undermined = nn.undermined; old.claimedBy = nn.claimedBy;
      old.title = nn.title; old.type = nn.type; old.body = nn.body; old.rank = nn.rank; old.blockers = nn.blockers;
      keep.push(old);
    } else { nn.enter = 0; nn.flare = 1; keep.push(nn); }
  });
  S.nodes = keep; S.byNum = {}; S.nodes.forEach(function(n) { S.byNum[n.num] = n; });
  S.edges = ng.edges || []; S.graph = ng;
  relayoutWarm();                                                       // survivors pinned; only new nodes placed
  S.nodes.forEach(function(n) { if (n.px == null) { n.px = n.x; n.py = n.y; } }); // new nodes appear at their final spot
  setupFog(); buildHud();                                               // note: no fitCamera — keep the user's view
  if (S.selected) { if (S.byNum[S.selected.num]) fillPanel(S.selected); else closePanel(); }
}

// One poller for the whole app: only runs while a map is open, and re-checks the
// effort mid-flight so switching maps never applies stale data.
export function startPolling() {
  setInterval(function() {
    var eff = S.currentEffort;
    if (!eff || S.polling) return; S.polling = true;
    fetch("/api/version?effort=" + encodeURIComponent(eff)).then(function(r) { return r.text(); }).then(function(v) {
      v = v.trim();
      if (S.currentEffort !== eff) return;
      if (S.lastVersion === null) { S.lastVersion = v; return; }
      if (v === S.lastVersion) return;
      S.lastVersion = v;
      return fetch("/api/graph?effort=" + encodeURIComponent(eff)).then(function(r) { return r.json(); }).then(function(g) { if (S.currentEffort === eff) updateGraph(g); });
    }).then(function() { S.polling = false; }).catch(function() { S.polling = false; });
  }, 1500);
}

// unmountMap tears the constellation down cleanly: no stale nodes/edges/fog can
// bleed through the (semi-transparent) splash or map-list overlays, and the
// poller goes idle because currentEffort is cleared.
function unmountMap() {
  S.currentEffort = null; S.lastVersion = null; S.selected = null;
  S.graph = null; S.nodes = []; S.edges = []; S.byNum = {}; S.fogPts = [];
  closePanel();
}

// leaveMap runs next (splash or the map list) straight away and lets the
// constellation dissolve behind the overlay, which is semi-transparent — so the
// reader gets the destination immediately and the stars fade out under it rather
// than making them wait. The unmount is deferred to the end of the fade.
//
// Opening another map mid-dissolve is safe: loadMap starts a new fade, and
// startFade replaces fade.cb, so this fade's pending unmountMap is dropped and
// cannot wipe the freshly loaded graph. The map chrome is hidden up front so it
// doesn't hang over the dissolving stars.
export function leaveMap(next) {
  S.currentEffort = null; // stop the poller applying mid-dissolve
  document.getElementById("hud").style.display = "none";
  document.getElementById("backbtn").style.display = "none";
  document.getElementById("hint").style.display = "none";
  closePanel();
  if (S.screen !== "map" || !S.nodes.length) { unmountMap(); next(); return; }
  next();                                  // destination appears at once
  startFade(S.mapAlpha, 0, 1.8, unmountMap); // stars dissolve behind it
}

function setScreen(s) {
  S.screen = s;
  document.getElementById("splash").style.display = s === "splash" ? "flex" : "none";
  document.getElementById("maplist").style.display = s === "maplist" ? "flex" : "none";
  document.getElementById("hud").style.display = s === "map" ? "flex" : "none";
  document.getElementById("backbtn").style.display = s === "map" ? "flex" : "none";
  document.getElementById("hint").style.display = s === "map" ? "block" : "none";
  if (s !== "map") closePanel();
}

// renderRecents paints the list. The remove endpoint answers with the trimmed
// list, so forgetting an entry re-renders from the server's truth rather than
// from a guess about what is left.
function renderRecents(rs) {
  var rc = document.getElementById("recents"); rc.innerHTML = "";
  if (!rs || !rs.length) { rc.innerHTML = "<div class='muted' style='text-align:left;font-size:12px'>No recent projects yet.</div>"; return; }
  rs.forEach(function(p) {
    var b = el("button", "recent");
    b.appendChild(el("span", "rname", p.name));
    b.appendChild(el("span", "rmeta", p.maps + " map" + (p.maps === 1 ? "" : "s")));
    // A span, not a button: this row is itself a button, and nesting one inside
    // another is invalid. stopPropagation keeps the dismiss from opening the
    // project it is dismissing.
    var x = el("span", "rx", "×"); x.title = "Forget this project";
    x.onclick = function(ev) {
      ev.stopPropagation();
      fetch("/api/recents/remove?project=" + encodeURIComponent(p.path), {method: "POST"})
        .then(function(r) { return r.json(); }).then(renderRecents);
    };
    b.appendChild(x);
    b.onclick = function() { openProject(p.path); };
    rc.appendChild(b);
  });
}

export function showSplash() {
  setScreen("splash");
  document.getElementById("recents").innerHTML = "";
  fetch("/api/recents").then(function(r) { return r.json(); }).then(renderRecents);
}

export function openProject(path) {
  S.lastProject = path;
  fetch("/api/maps?project=" + encodeURIComponent(path)).then(function(r) { return r.json(); }).then(function(maps) {
    document.getElementById("projname").textContent = path.replace(/\/+$/, "").split("/").pop();
    var g = document.getElementById("cards"); g.innerHTML = "";
    if (!maps || !maps.length) { g.innerHTML = "<div class='muted'>No wayfinder maps in this project's .plan/</div>"; }
    (maps || []).forEach(function(m) {
      var c = el("button", "card");
      c.appendChild(el("div", "cname", m.name));
      var meta = el("div", "cmeta");
      meta.appendChild(el("span", null, m.resolved + "/" + m.total + " resolved"));
      if (m.frontier) meta.appendChild(el("span", "cfront", m.frontier + " on frontier"));
      c.appendChild(meta);
      var bar = el("div", "cbar"), sp = document.createElement("span");
      sp.style.width = (m.total ? Math.round(m.resolved * 100 / m.total) : 0) + "%"; bar.appendChild(sp); c.appendChild(bar);
      c.onclick = function() { loadMap(m.path); };
      g.appendChild(c);
    });
    setScreen("maplist");
  });
}

export function loadMap(effort) {
  S.currentEffort = effort; S.lastVersion = null; S.selected = null;
  fetch("/api/graph?effort=" + encodeURIComponent(effort)).then(function(r) { return r.json(); }).then(function(g) {
    applyGraph(g); setScreen("map"); startFade(0, 1, 2.2); // dissolve the constellation in
  });
}
