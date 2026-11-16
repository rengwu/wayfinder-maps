---
type: implement
blocked_by: [06, 08]
---

# Group entries by label

## What to build

Entries are grouped by their label first and by title prefix only when no label is set.
This is what the audit found works for most PRs.

- [ ] a labelled PR lands in the group named by its label
- [ ] an unlabelled PR with a conventional prefix lands in the matching group
