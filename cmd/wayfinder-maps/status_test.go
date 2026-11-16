package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/RamsesKools/wayfinder-maps/internal/wayfinder"
)

type tk struct {
	num     int
	title   string
	typ     wayfinder.Type
	answer  bool
	claimed string
	pr      string
	afk     bool
	blocked []int
}

func statusEffort(tks ...tk) *wayfinder.Effort {
	e := &wayfinder.Effort{
		Name: "e",
		Map:  wayfinder.ParseMap("map.md", "# Sample map\n\n## Destination\nD.\n"),
	}
	for _, k := range tks {
		t := &wayfinder.Ticket{Num: k.num, Path: "missing.md", Title: k.title, Type: k.typ, BlockedBy: k.blocked, PR: k.pr, AFKFlag: k.afk, ClaimedBy: k.claimed}
		t.HasAnswer, t.AnswerHeading = k.answer, k.answer
		t.Derive()
		e.Tickets = append(e.Tickets, t)
	}
	return e
}

func runStatus(t *testing.T, e *wayfinder.Effort) string {
	t.Helper()
	var buf bytes.Buffer
	if code := status(&buf, e); code != 0 {
		t.Fatalf("status exit = %d", code)
	}
	return buf.String()
}

func TestStatusWithoutImplementTickets(t *testing.T) {
	got := runStatus(t, statusEffort(
		tk{num: 1, title: "Settled", typ: wayfinder.TypeGrilling, answer: true},
		tk{num: 2, title: "Pick a name", typ: wayfinder.TypeGrilling},
		tk{num: 3, title: "Survey options", typ: wayfinder.TypeResearch, blocked: []int{1}},
		tk{num: 4, title: "Waits", typ: wayfinder.TypeGrilling, blocked: []int{2}},
		tk{num: 5, title: "In flight", typ: wayfinder.TypePrototype, claimed: "ana"},
	))
	want := `Sample map
1 resolved · 1 claimed · 3 open · 0 out of scope

Frontier — ready to claim, first by number wins (1 AFK, 1 HITL):
  02  Pick a name                                     grilling   HITL
  03  Survey options                                  research   AFK

Claimed — in flight:
  05  In flight                                       by ana

Blocked:
  04  Waits                                           waits on 02

Fog: 0 patches, 0 anchored to a ticket
`
	if got != want {
		t.Errorf("status output:\n%s\nwant:\n%s", got, want)
	}
	for _, s := range []string{"Decisions:", "Implementation"} {
		if strings.Contains(got, s) {
			t.Errorf("output without implement tickets contains %q", s)
		}
	}
}

func TestStatusWithImplementTickets(t *testing.T) {
	got := runStatus(t, statusEffort(
		tk{num: 1, title: "Settled", typ: wayfinder.TypeGrilling, answer: true},
		tk{num: 2, title: "Shipped with PR", typ: wayfinder.TypeImplement, answer: true, pr: "https://github.com/o/r/pull/2"},
		tk{num: 3, title: "Shipped without PR", typ: wayfinder.TypeImplement, answer: true},
		tk{num: 4, title: "Being built", typ: wayfinder.TypeImplement, claimed: "worker-1"},
		tk{num: 5, title: "Ready to build", typ: wayfinder.TypeImplement, blocked: []int{1}},
		tk{num: 6, title: "Not yet", typ: wayfinder.TypeImplement, blocked: []int{4}},
	))
	want := `Sample map
Decisions:      1 resolved · 0 claimed · 0 open · 0 out of scope
Implementation: 2 resolved · 1 claimed · 2 open · 0 out of scope

Frontier — ready to claim, first by number wins (1 AFK, 0 HITL):
  05  Ready to build                                  implement  AFK

Implementation — in flight and shipped:
  02  Shipped with PR                                 https://github.com/o/r/pull/2
  03  Shipped without PR                              no pr
  04  Being built                                     by worker-1

Blocked:
  06  Not yet                                         waits on 04

Fog: 0 patches, 0 anchored to a ticket
`
	if got != want {
		t.Errorf("status output:\n%s\nwant:\n%s", got, want)
	}
}

func TestStatusImplementOutOfScopeAndClaimedDecision(t *testing.T) {
	ruled := tk{num: 2, title: "Dropped", typ: wayfinder.TypeImplement}
	e := statusEffort(
		tk{num: 1, title: "Decide", typ: wayfinder.TypeGrilling, claimed: "me"},
		ruled,
	)
	e.Tickets[1].HasRuledOut, e.Tickets[1].RuledOutHeading = true, true
	e.Tickets[1].Derive()
	got := runStatus(t, e)
	for _, s := range []string{
		"Decisions:      0 resolved · 1 claimed · 0 open · 0 out of scope",
		"Implementation: 0 resolved · 0 claimed · 0 open · 1 out of scope",
		"Claimed — in flight:\n  01  Decide",
		"Frontier: empty",
	} {
		if !strings.Contains(got, s) {
			t.Errorf("output missing %q:\n%s", s, got)
		}
	}
	if strings.Contains(got, "Implementation — in flight") {
		t.Errorf("out-of-scope implement ticket should not be listed:\n%s", got)
	}
}

func TestStatusMarksAFKAndHITL(t *testing.T) {
	got := runStatus(t, statusEffort(
		tk{num: 1, title: "Grill", typ: wayfinder.TypeGrilling},
		tk{num: 2, title: "Proto", typ: wayfinder.TypePrototype},
		tk{num: 3, title: "Manual task", typ: wayfinder.TypeTask},
		tk{num: 4, title: "Agent task", typ: wayfinder.TypeTask, afk: true},
		tk{num: 5, title: "Research", typ: wayfinder.TypeResearch},
		tk{num: 6, title: "Build", typ: wayfinder.TypeImplement},
	))
	if !strings.Contains(got, "(3 AFK, 3 HITL)") {
		t.Errorf("missing frontier summary:\n%s", got)
	}
	marks := map[string]string{
		"01": "HITL", "02": "HITL", "03": "HITL",
		"04": "AFK", "05": "AFK", "06": "AFK",
	}
	for _, line := range strings.Split(got, "\n") {
		if len(line) < 4 || !strings.HasPrefix(line, "  0") {
			continue
		}
		num := strings.TrimSpace(line[:4])
		if !strings.HasSuffix(line, "  "+marks[num]) {
			t.Errorf("ticket %s: line %q should end with %s", num, line, marks[num])
		}
	}
}
