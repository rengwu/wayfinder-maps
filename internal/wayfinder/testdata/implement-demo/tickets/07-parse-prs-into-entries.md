---
type: implement
blocked_by: [06]
pr: https://github.com/example/relnotes/pull/12
---

# Parse merged PRs into entries

## What to build

Given a file of merged PRs, the tool produces one typed entry per PR with a title, a type taken from the labels, and the author.
A PR with no usable label becomes an entry of type `other` rather than failing the run.

- [x] every PR in the fixture file becomes an entry
- [x] an unlabelled PR becomes type `other`
- [x] a malformed record is reported with its position and skipped

## Answer

Shipped the parser in [PR 12](https://github.com/example/relnotes/pull/12).
All 212 fixture PRs parse, and 9 of them land in `other`.
