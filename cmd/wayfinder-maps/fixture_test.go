package main

import (
	"path/filepath"
	"testing"

	"github.com/RamsesKools/wayfinder-maps/internal/wayfinder"
)

func loadImplementDemo(t *testing.T) *wayfinder.Effort {
	t.Helper()
	e, err := wayfinder.Load(filepath.Join("..", "..", "internal", "wayfinder", "testdata", "implement-demo"))
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestStatusImplementDemo(t *testing.T) {
	got := runStatus(t, loadImplementDemo(t))
	want := `Release notes generator
Decisions:      6 resolved · 1 claimed · 7 open · 1 out of scope
Implementation: 1 resolved · 1 claimed · 2 open · 1 out of scope

Frontier — ready to claim, first by number wins (3 AFK, 2 HITL):
  09  Add a since flag to select a release range      implement  AFK
  13  Get a publishing token from the platform team   task       HITL
  14  Create the empty releases repository            task       AFK
  16  Decide how breaking changes are flagged         grilling   HITL
  17  Compare hosting options for published notes     research   AFK

Claimed — in flight:
  12  Check API rate limits for listing PRs           by research-agent

Implementation — in flight and shipped:
  07  Parse merged PRs into entries                   https://github.com/example/relnotes/pull/12
  08  Render entries as Markdown                      by worker-2

Blocked:
  10  Group entries by label                          waits on 06, 08
  11  Decide how long entries are shortened           waits on 08
  15  Sketch the breaking-change banner               waits on 16
  18  Plan the implementation of publishing           waits on 13, 16, 17

Undermined — resolved on a premise that later changed:
  03  Sketch the release notes layout                 broken by 04

Fog: 3 patches, 2 anchored to a ticket
`
	if got != want {
		t.Errorf("status output:\n%s\nwant:\n%s", got, want)
	}
}

func TestGraphImplementDemo(t *testing.T) {
	g := buildGraph(loadImplementDemo(t))
	c := g.Counts
	if c.Total != 20 || c.Resolved != 7 || c.Claimed != 2 || c.Open != 9 || c.OutOfScope != 2 {
		t.Errorf("overall counts = %+v", c)
	}
	if want := (kindCountsDoc{Resolved: 6, Claimed: 1, Open: 7, OutOfScope: 1, Total: 15}); c.Decisions != want {
		t.Errorf("decision counts = %+v, want %+v", c.Decisions, want)
	}
	if want := (kindCountsDoc{Resolved: 1, Claimed: 1, Open: 2, OutOfScope: 1, Total: 5}); c.Implement != want {
		t.Errorf("implement counts = %+v, want %+v", c.Implement, want)
	}
	if want := (frontierDoc{AFK: 3, HITL: 2, Total: 5}); c.Frontier != want {
		t.Errorf("frontier counts = %+v, want %+v", c.Frontier, want)
	}

	statuses := map[string]int{}
	for _, n := range g.Nodes {
		statuses[n.Status]++
	}
	want := map[string]int{"resolved": 7, "claimed": 2, "frontier": 5, "blocked": 4, "out_of_scope": 2}
	for k, v := range want {
		if statuses[k] != v {
			t.Errorf("%d nodes %s, want %d (all: %v)", statuses[k], k, v, statuses)
		}
	}
	if len(g.Fog) != 3 {
		t.Errorf("fog = %d, want 3", len(g.Fog))
	}
}
