// A small markdown renderer for the detail panel. No dependency, no build
// step. HTML is escaped first, then a line-based block pass (fences, headings,
// lists, hr, blockquotes, paragraphs) wraps it, and an inline pass handles
// code / links / bold / italic.
function esc(s) { return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;"); }

function mdFmt(t) {
  t = t.replace(/\[([^\]]+)\]\(([^)]+)\)/g, function(m, txt, url) {
    if (/^https?:\/\//.test(url)) return "<a href='" + url + "' target='_blank' rel='noopener'>" + txt + "</a>";
    var tm = url.match(/(?:^|\/)0*(\d+)-[^/)]*\.md/);
    if (tm) return "<a class='xlink' data-goto='" + tm[1] + "'>" + txt + "</a>";
    return "<span class='xlink'>" + txt + "</span>";
  });
  t = t.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  t = t.replace(/\*([^*\n]+)\*/g, "<em>$1</em>");
  return t;
}

// Protect code spans as placeholders before emphasis, so bold can WRAP an
// inline-code span, and a * inside code (Go pointer types) is never italicised.
function mdInline(s) {
  var codes = [];
  s = s.replace(/`([^`]+)`/g, function(m, c) { codes.push(c); return "\x00" + (codes.length - 1) + "\x00"; });
  s = mdFmt(s);
  return s.replace(/\x00(\d+)\x00/g, function(m, i) { return "<code>" + codes[i] + "</code>"; });
}

export function mdToHtml(src) {
  if (!src) return "<p class='muted'>(no body)</p>";
  var lines = src.split("\n"), out = [], para = [], list = null, inCode = false, code = [];
  function fp() { if (para.length) { out.push("<p>" + mdInline(esc(para.join(" "))) + "</p>"); para = []; } }
  function fl() { if (list) { out.push("</" + list + ">"); list = null; } }
  for (var i = 0; i < lines.length; i++) {
    var line = lines[i], t = line.replace(/^\s+/, "");
    if (t.indexOf("```") === 0) {
      if (inCode) { out.push("<pre><code>" + esc(code.join("\n")) + "</code></pre>"); inCode = false; code = []; }
      else { fp(); fl(); inCode = true; code = []; }
      continue;
    }
    if (inCode) { code.push(line); continue; }
    if (t === "") { fp(); fl(); continue; }
    var h = t.match(/^(#{1,6})\s+(.*)$/);
    if (h) { fp(); fl(); var lv = h[1].length; out.push("<h" + lv + ">" + mdInline(esc(h[2])) + "</h" + lv + ">"); continue; }
    if (/^(---+|\*\*\*+|___+)$/.test(t)) { fp(); fl(); out.push("<hr>"); continue; }
    var li = t.match(/^([-*+]|\d+\.)\s+(.*)$/);
    if (li) { fp(); var ty = /\d/.test(li[1]) ? "ol" : "ul"; if (list && list !== ty) fl(); if (!list) { out.push("<" + ty + ">"); list = ty; } out.push("<li>" + mdInline(esc(li[2])) + "</li>"); continue; }
    var bq = t.match(/^>\s?(.*)$/);
    if (bq) { fp(); fl(); out.push("<blockquote>" + mdInline(esc(bq[1])) + "</blockquote>"); continue; }
    if (list) fl();
    para.push(t);
  }
  if (inCode) out.push("<pre><code>" + esc(code.join("\n")) + "</code></pre>");
  fp(); fl();
  return out.join("\n");
}
