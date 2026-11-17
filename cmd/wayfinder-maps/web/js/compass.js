// The compass control: a 3x3 pad. The eight outer buttons point the map's flow
// that way, the centre one turns the direction off (radial layout). The active
// button glows.
import {el} from "./util.js";
import {NONE, flow} from "./direction.js";

var GRID = ["NW", "N", "NE", "W", NONE, "E", "SW", "S", "SE"];
var NAME = {N: "north", NE: "north-east", E: "east", SE: "south-east", S: "south", SW: "south-west", W: "west", NW: "north-west"};
var NS = "http://www.w3.org/2000/svg";

function svg(markup, rotate) {
  var s = document.createElementNS(NS, "svg");
  s.setAttribute("viewBox", "0 0 20 20"); s.setAttribute("class", "cglyph");
  if (rotate != null) s.style.transform = "rotate(" + rotate + "deg)";
  s.innerHTML = markup;
  return s;
}

// mountCompass fills root with the pad and returns set(dir) to move the glow.
export function mountCompass(root, onPick) {
  var buttons = {};
  root.innerHTML = "";
  GRID.forEach(function(d) {
    var b = el("button", "cbtn" + (d === NONE ? " cnone" : ""));
    b.type = "button";
    if (d === NONE) {
      b.title = "No direction: radial layout"; b.setAttribute("aria-label", "No direction");
      b.appendChild(svg('<circle cx="10" cy="10" r="3.2"/>'));
    } else {
      var f = flow(d);
      b.title = "Flow " + NAME[d]; b.setAttribute("aria-label", "Flow " + NAME[d]);
      b.appendChild(svg('<path d="M3 10h12M10.5 4.5L16 10l-5.5 5.5"/>', Math.round(Math.atan2(f.y, f.x) * 180 / Math.PI)));
    }
    b.dataset.dir = d;
    b.onclick = function() { onPick(d); };
    buttons[d] = b; root.appendChild(b);
  });
  return function set(dir) {
    GRID.forEach(function(d) {
      buttons[d].classList.toggle("on", d === dir);
      buttons[d].setAttribute("aria-pressed", d === dir ? "true" : "false");
    });
  };
}
