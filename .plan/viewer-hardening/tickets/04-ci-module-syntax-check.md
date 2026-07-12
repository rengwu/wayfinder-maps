---
type: task
blocked_by: []
---

# CI guard for the embedded frontend

## Question

Since the web/ refactor, a syntax error in a JS module no longer fails
`go build` — it ships embedded in the binary and breaks only at runtime, in
the browser. The repo currently has no push/PR CI at all (only
`release.yml`), so nothing would catch it before a release.

Add a CI workflow that runs on push/PR: `go build ./...`, `go vet`, `go test
./...`, and a syntax pass over the frontend (`node --check` per module under
`web/js/`, or equivalent). Keep it minimal — the point is the embedded-JS
hole, not a full pipeline.
