---
type: implement
blocked_by: [06, 07]
claimed_by: worker-2
claimed_at: 2026-11-10T09:00:00Z
---

# Render entries as Markdown

## What to build

The tool prints entries as one Markdown release file, grouped by entry type, newest group first.
Each entry is one bullet with its title and a link to its PR.

- [x] entries print grouped by type
- [ ] each bullet links to its PR
- [ ] an empty range prints a clear "no changes" line
