---
type: implement
blocked_by: [06, 07]
---

# Add a since flag to select a release range

## What to build

A user can pass `--since <tag>` and get notes only for PRs merged after that tag.
Without the flag the tool covers every PR in the input.

- [ ] `--since v1.2.0` drops PRs merged before that tag
- [ ] an unknown tag exits non-zero with a message naming it
