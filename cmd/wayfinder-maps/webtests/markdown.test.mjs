// Unit tests for the detail panel's markdown renderer. Zero dependencies:
//   node --test cmd/wayfinder-maps/webtests/*.test.mjs
// Lives outside web/ on purpose — //go:embed web would ship anything in there.
import { test } from "node:test";
import assert from "node:assert/strict";
import { mdToHtml } from "../web/js/markdown.js";

test("empty body falls back to the muted placeholder", () => {
  assert.equal(mdToHtml(""), "<p class='muted'>(no body)</p>");
  assert.equal(mdToHtml(null), "<p class='muted'>(no body)</p>");
  assert.equal(mdToHtml(undefined), "<p class='muted'>(no body)</p>");
});

test("HTML in prose is escaped", () => {
  assert.equal(mdToHtml("a <script> & b"), "<p>a &lt;script&gt; &amp; b</p>");
});

test("consecutive lines join into one paragraph, a blank line splits", () => {
  assert.equal(mdToHtml("one\ntwo\n\nthree"), "<p>one two</p>\n<p>three</p>");
});

test("leading whitespace is stripped outside fences", () => {
  assert.equal(mdToHtml("  hello"), "<p>hello</p>");
});

test("bold, italic, inline code", () => {
  assert.equal(mdToHtml("**b**"), "<p><strong>b</strong></p>");
  assert.equal(mdToHtml("*i*"), "<p><em>i</em></p>");
  assert.equal(mdToHtml("`c`"), "<p><code>c</code></p>");
  assert.equal(
    mdToHtml("**b** and *i* and `c`"),
    "<p><strong>b</strong> and <em>i</em> and <code>c</code></p>"
  );
});

test("italic can wrap bold", () => {
  assert.equal(mdToHtml("*a **b** c*"), "<p><em>a <strong>b</strong> c</em></p>");
});

test("bold can wrap an inline code span", () => {
  assert.equal(mdToHtml("**`go vet`**"), "<p><strong><code>go vet</code></strong></p>");
});

test("a * inside code is never italicised", () => {
  assert.equal(
    mdToHtml("`*Ticket` and *real*"),
    "<p><code>*Ticket</code> and <em>real</em></p>"
  );
});

test("HTML inside a code span stays escaped", () => {
  assert.equal(mdToHtml("`a < b`"), "<p><code>a &lt; b</code></p>");
});

test("heading levels 1 through 6; seven hashes is prose", () => {
  for (let lv = 1; lv <= 6; lv++) {
    assert.equal(mdToHtml("#".repeat(lv) + " T"), "<h" + lv + ">T</h" + lv + ">");
  }
  assert.equal(mdToHtml("####### T"), "<p>####### T</p>");
});

test("headings take inline formatting", () => {
  assert.equal(mdToHtml("## `code` **b**"), "<h2><code>code</code> <strong>b</strong></h2>");
});

test("hr from ---, ***, ___; two dashes is prose", () => {
  assert.equal(mdToHtml("---"), "<hr>");
  assert.equal(mdToHtml("***"), "<hr>");
  assert.equal(mdToHtml("___"), "<hr>");
  assert.equal(mdToHtml("--"), "<p>--</p>");
});

test("unordered list from -, * and +", () => {
  assert.equal(mdToHtml("- a\n- b"), "<ul>\n<li>a</li>\n<li>b</li>\n</ul>");
  assert.equal(mdToHtml("* a\n+ b"), "<ul>\n<li>a</li>\n<li>b</li>\n</ul>");
});

test("ordered list", () => {
  assert.equal(mdToHtml("1. a\n2. b"), "<ol>\n<li>a</li>\n<li>b</li>\n</ol>");
});

test("switching list type closes the open list", () => {
  assert.equal(
    mdToHtml("- a\n1. b"),
    "<ul>\n<li>a</li>\n</ul>\n<ol>\n<li>b</li>\n</ol>"
  );
});

test("prose after a list closes it", () => {
  assert.equal(mdToHtml("- a\ntext"), "<ul>\n<li>a</li>\n</ul>\n<p>text</p>");
});

test("list items take inline formatting", () => {
  assert.equal(
    mdToHtml("- **b** `c`"),
    "<ul>\n<li><strong>b</strong> <code>c</code></li>\n</ul>"
  );
});

test("one blockquote element per quoted line", () => {
  assert.equal(mdToHtml("> q"), "<blockquote>q</blockquote>");
  assert.equal(
    mdToHtml("> a\n> b"),
    "<blockquote>a</blockquote>\n<blockquote>b</blockquote>"
  );
});

test("fenced block quotes markdown-looking content verbatim", () => {
  assert.equal(
    mdToHtml("```\n## Answer\n**not bold**\n```"),
    "<pre><code>## Answer\n**not bold**</code></pre>"
  );
});

test("fenced block escapes HTML, keeps indentation, drops the language tag", () => {
  assert.equal(
    mdToHtml("```go\nif a < b {\n\treturn\n}\n```"),
    "<pre><code>if a &lt; b {\n\treturn\n}</code></pre>"
  );
});

test("an unclosed fence still flushes at end of input", () => {
  assert.equal(mdToHtml("```\ndangling"), "<pre><code>dangling</code></pre>");
});

test("a fence closes an open paragraph first", () => {
  assert.equal(mdToHtml("para\n```\nc\n```"), "<p>para</p>\n<pre><code>c</code></pre>");
});

test("external links open in a new tab", () => {
  assert.equal(
    mdToHtml("[site](https://example.com/x)"),
    "<p><a href='https://example.com/x' target='_blank' rel='noopener'>site</a></p>"
  );
  assert.equal(
    mdToHtml("[plain](http://example.com/)"),
    "<p><a href='http://example.com/' target='_blank' rel='noopener'>plain</a></p>"
  );
});

test("ticket links become data-goto with leading zeros stripped", () => {
  assert.equal(
    mdToHtml("[t](04-ci-module-syntax-check.md)"),
    "<p><a class='xlink' data-goto='4'>t</a></p>"
  );
  assert.equal(
    mdToHtml("[t](./tickets/012-foo.md)"),
    "<p><a class='xlink' data-goto='12'>t</a></p>"
  );
});

test("task list items render as disabled checkboxes", () => {
  assert.equal(
    mdToHtml("- [ ] open one\n- [x] done *one*\n- [X] done two\n- plain"),
    "<ul>\n<li class='task'><input type='checkbox' disabled> open one</li>\n" +
    "<li class='task'><input type='checkbox' disabled checked> done <em>one</em></li>\n" +
    "<li class='task'><input type='checkbox' disabled checked> done two</li>\n" +
    "<li>plain</li>\n</ul>"
  );
});

test("other relative links render as inert xlink spans", () => {
  assert.equal(
    mdToHtml("[readme](../README.md)"),
    "<p><span class='xlink'>readme</span></p>"
  );
});
