// Small pure helpers with no app state.
export function mod(a, n) { return ((a % n) + n) % n; }
export function clamp(v, a, b) { return v < a ? a : (v > b ? b : v); }
export function pad2(n) { return (n < 10 ? "0" : "") + n; }

// Deterministic PRNG: same seed, same sequence, so layouts are reproducible.
export function mulberry32(a) {
  return function() {
    a |= 0; a = a + 0x6D2B79F5 | 0;
    var t = Math.imul(a ^ a >>> 15, 1 | a);
    t = t + Math.imul(t ^ t >>> 7, 61 | t) ^ t;
    return ((t ^ t >>> 14) >>> 0) / 4294967296;
  };
}

export function hexA(hex, al) {
  var h = hex.replace("#", "");
  var r = parseInt(h.substr(0, 2), 16), g = parseInt(h.substr(2, 2), 16), b = parseInt(h.substr(4, 2), 16);
  return "rgba(" + r + "," + g + "," + b + "," + al + ")";
}

export function el(tag, cls, txt) {
  var e = document.createElement(tag);
  if (cls) e.className = cls;
  if (txt != null) e.textContent = txt;
  return e;
}
