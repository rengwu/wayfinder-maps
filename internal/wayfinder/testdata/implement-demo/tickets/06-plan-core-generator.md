---
type: grilling
blocked_by: [01, 03, 04, 05]
---

# Plan the implementation of the core generator

## Question

How do we slice the generator into implement tickets?

## Answer

Four vertical slices, each one PR in the generator repo.
Parsing comes first and everything else builds on it.
Rendering, release ranges and label grouping follow, with grouping waiting on the renderer.
Created [Parse merged PRs into entries](./07-parse-prs-into-entries.md), [Render entries as Markdown](./08-render-markdown.md), [Add a since flag to select a release range](./09-since-flag.md) and [Group entries by label](./10-group-by-label.md).
