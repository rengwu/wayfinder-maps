---
type: task
afk: true
blocked_by: []
---

# Export last quarter's merged PRs to a fixture file

## Question

Save the merged PRs of the last quarter as JSON so the parser can be built offline.

## Answer

Exported 212 PRs to `testdata/prs.json` with the `gh` CLI.
Each record has the title, labels, author and merge date.
