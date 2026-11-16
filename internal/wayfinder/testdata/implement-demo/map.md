# Release notes generator

## Destination

A `relnotes` CLI that turns a range of merged pull requests into Markdown release notes, published for the whole team to read.

## Notes

Example map for the implement phase: decisions and implement tickets live in one graph.
Implementation is part of this map, so tickets may deliver code.
Work is split across a hub and workers: the hub claims tickets and owns this file, workers write their own ticket.

## Decisions so far

- [Choose the audience and format of release notes](./tickets/01-audience-and-format.md): written for end users, one Markdown file per release.
- [Survey existing changelog tools](./tickets/02-survey-changelog-tools.md): none read our PR labels, so we build.
- [Sketch the release notes layout](./tickets/03-sketch-layout.md): grouped by type with a short summary line.
- [Audit the quality of merged PR titles](./tickets/04-audit-pr-titles.md): titles are free-form, so grouping cannot rely on prefixes.
- [Export last quarter's merged PRs to a fixture file](./tickets/05-export-pr-fixture.md): 212 PRs saved as JSON for offline work.
- [Plan the implementation of the core generator](./tickets/06-plan-core-generator.md): four slices, parse first, then render, ranges and grouping.

## Implemented so far

- [Parse merged PRs into entries](./tickets/07-parse-prs-into-entries.md): the parser turns PR JSON into typed entries.

## Not yet specified

- **Release automation.** Running the generator on every tag, and who reviews the output before it goes out. <clears-with: 18>
- **Truncation rules for the summary line.** Depends on what the renderer shows for very long titles. <clears-with: 11>
- **Translations of release notes.** Nobody has asked yet, and the entry model may need a language field.

## Out of scope

- [Support GitLab merge requests](./tickets/19-support-gitlab.md): the team hosts everything on GitHub, so a second forge is past the destination.
- [Serve release notes from a hosted web page](./tickets/20-hosted-web-page.md): the destination is a published file, not a service to run.
