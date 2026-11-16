---
type: research
blocked_by: []
assets: [.plan/release-notes/assets/changelog-tools.md]
---

# Survey existing changelog tools

## Question

Does an existing tool already turn merged PRs into release notes the way we want?

## Answer

Four tools were compared in [the survey](../assets/changelog-tools.md).
All of them group by commit prefix, and none reads PR labels, so we build our own.
