---
type: task
blocked_by: []
---

# Unit tests for the markdown renderer

## Question

`web/js/markdown.js` renders arbitrary ticket prose into the detail panel. It
became a pure importable module in the web/ refactor, so it can finally be
tested with `node --test` and zero dependencies — and it's the frontend code
most likely to regress subtly (escaping, nested emphasis around code spans,
fences quoting `## Answer`, cross-ticket link extraction).

Write that test file, covering at minimum: HTML escaping, bold/italic/inline
code and their nesting, fenced blocks (including markdown-looking content
inside them), ordered/unordered lists, blockquotes, hr, heading levels,
external links vs `data-goto` ticket links, and the empty-body fallback.
Decide where tests live and how they run (likely `node --test` invoked from
CI once the [CI syntax check](04-ci-module-syntax-check.md) exists).
