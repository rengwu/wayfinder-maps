package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RamsesKools/wayfinder-maps/internal/wayfinder"
)

func graphTicket(num int, typ wayfinder.Type, answered bool, pr string, afk bool, blocked ...int) *wayfinder.Ticket {
	t := &wayfinder.Ticket{Num: num, Path: "missing.md", Title: "T", Type: typ, BlockedBy: blocked, PR: pr, AFKFlag: afk}
	t.HasAnswer, t.AnswerHeading = answered, answered
	t.Derive()
	return t
}

func TestBuildGraphImplementFields(t *testing.T) {
	e := &wayfinder.Effort{
		Name: "e",
		Map:  wayfinder.ParseMap("map.md", "# M\n\n## Destination\nD.\n"),
		Tickets: []*wayfinder.Ticket{
			graphTicket(1, wayfinder.TypeGrilling, true, "", false),
			graphTicket(2, wayfinder.TypeImplement, true, "https://example.com/o/r/pull/2", false, 1),
			graphTicket(3, wayfinder.TypeImplement, false, "", false, 1),
			graphTicket(4, wayfinder.TypeGrilling, false, "", false),
			graphTicket(5, wayfinder.TypeTask, false, "", true),
			graphTicket(6, wayfinder.TypeResearch, false, "", false, 4),
		},
	}
	g := buildGraph(e)

	byNum := map[int]nodeDoc{}
	for _, n := range g.Nodes {
		byNum[n.Num] = n
	}
	if n := byNum[1]; n.Kind != "decision" || n.AFK || n.PR != "" {
		t.Errorf("grilling node = %+v", n)
	}
	if n := byNum[2]; n.Kind != "implement" || !n.AFK || n.PR != "https://example.com/o/r/pull/2" {
		t.Errorf("resolved implement node = %+v", n)
	}
	if n := byNum[5]; n.Kind != "decision" || !n.AFK {
		t.Errorf("afk task node = %+v", n)
	}

	c := g.Counts
	if c.Total != 6 || c.Resolved != 2 || c.Open != 4 {
		t.Errorf("overall counts = %+v", c)
	}
	if c.Decisions.Total != 4 || c.Decisions.Resolved != 1 || c.Decisions.Open != 3 {
		t.Errorf("decision counts = %+v", c.Decisions)
	}
	if c.Implement.Total != 2 || c.Implement.Resolved != 1 || c.Implement.Open != 1 {
		t.Errorf("implement counts = %+v", c.Implement)
	}
	if c.Frontier != (frontierDoc{AFK: 2, HITL: 1, Total: 3}) {
		t.Errorf("frontier counts = %+v, want afk 2 (03, 05) hitl 1 (04); 06 is blocked", c.Frontier)
	}
}

func TestGraphJSONKeepsExistingFieldsAndOmitsEmptyPR(t *testing.T) {
	e := &wayfinder.Effort{
		Name:    "e",
		Map:     wayfinder.ParseMap("map.md", "# M\n\n## Destination\nD.\n"),
		Tickets: []*wayfinder.Ticket{graphTicket(1, wayfinder.TypeGrilling, false, "", false)},
	}
	b, err := json.Marshal(buildGraph(e))
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Counts map[string]json.RawMessage   `json:"counts"`
		Nodes  []map[string]json.RawMessage `json:"nodes"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"resolved", "claimed", "open", "outOfScope", "total", "decisions", "implement", "frontier"} {
		if _, ok := doc.Counts[k]; !ok {
			t.Errorf("counts.%s missing", k)
		}
	}
	for _, k := range []string{"num", "title", "type", "status", "rank", "undermined", "blockers", "body", "kind", "afk"} {
		if _, ok := doc.Nodes[0][k]; !ok {
			t.Errorf("node.%s missing", k)
		}
	}
	if _, ok := doc.Nodes[0]["pr"]; ok {
		t.Error("empty pr should be omitted")
	}
}

func TestOpenEndpoint(t *testing.T) {
	var opened []string
	old := openURL
	openURL = func(u string) error { opened = append(opened, u); return nil }
	t.Cleanup(func() { openURL = old })
	h := newServer("")

	do := func(method, body string, hdr bool) int {
		req := httptest.NewRequest(method, "/api/open", strings.NewReader(body))
		if hdr {
			req.Header.Set("X-Wayfinder", "1")
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	if c := do("POST", "https://example.com/o/r/pull/2", true); c != http.StatusNoContent {
		t.Errorf("valid open = %d", c)
	}
	for name, c := range map[string]int{
		"no header":  do("POST", "https://example.com/", false),
		"GET":        do("GET", "https://example.com/", true),
		"file URL":   do("POST", "file:///etc/passwd", true),
		"javascript": do("POST", "javascript:alert(1)", true),
		"empty":      do("POST", "", true),
	} {
		if c != http.StatusBadRequest {
			t.Errorf("%s = %d, want 400", name, c)
		}
	}
	if len(opened) != 1 || opened[0] != "https://example.com/o/r/pull/2" {
		t.Errorf("opened = %v", opened)
	}
}

func TestPullEndpoint(t *testing.T) {
	dir := t.TempDir()
	plain := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".source.json"), []byte(`{"repo":"o/r","map":3,"auth":"gh","account":"me"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	var pulled []string
	old := runPull
	runPull = func(e string) error { pulled = append(pulled, e); return nil }
	defer func() { runPull = old }()
	h := newServer("")

	post := func(effort string, hdr bool) int {
		req := httptest.NewRequest(http.MethodPost, "/api/pull?effort="+url.QueryEscape(effort), nil)
		if hdr {
			req.Header.Set("X-Wayfinder", "1")
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}
	if c := post(dir, false); c != http.StatusBadRequest {
		t.Errorf("no header: %d", c)
	}
	if c := post(plain, true); c != http.StatusBadRequest {
		t.Errorf("not a mirror: %d", c)
	}
	if c := post(dir, true); c != http.StatusNoContent || len(pulled) != 1 || pulled[0] != dir {
		t.Errorf("pull: %d %v", c, pulled)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/source?effort="+url.QueryEscape(dir), nil))
	if !strings.Contains(rec.Body.String(), `"pullable":true`) || !strings.Contains(rec.Body.String(), `"account":"me"`) {
		t.Errorf("source = %s", rec.Body)
	}
}
