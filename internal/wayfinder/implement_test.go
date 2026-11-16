package wayfinder

import (
	"strings"
	"testing"
)

func TestParseImplementTicket(t *testing.T) {
	src := `---
type: implement
blocked_by: [01]
pr: https://example.com/o/r/pull/12 # the one PR
---

# Ship the thing

## What to build
Behaviour.

- [ ] criterion

## Answer
Shipped.
`
	tk, err := ParseTicket("p", "02-ship.md", src)
	if err != nil {
		t.Fatal(err)
	}
	if tk.Type != TypeImplement || tk.Status != StatusResolved {
		t.Errorf("type/status = %q/%q", tk.Type, tk.Status)
	}
	if tk.PR != "https://example.com/o/r/pull/12" {
		t.Errorf("pr = %q", tk.PR)
	}
	if tk.Title != "Ship the thing" || !tk.HasAnswer {
		t.Errorf("title/answer = %q/%v", tk.Title, tk.HasAnswer)
	}
}

func TestParseAFKField(t *testing.T) {
	parse := func(front string) *Ticket {
		tk, err := ParseTicket("p", "01-a.md", "---\n"+front+"\n---\n\n# T\n")
		if err != nil {
			t.Fatal(err)
		}
		return tk
	}
	if tk := parse("type: task\nafk: true"); !tk.AFKSet || !tk.AFKFlag {
		t.Errorf("afk: true parsed as set=%v flag=%v", tk.AFKSet, tk.AFKFlag)
	}
	if tk := parse("type: task\nafk: false"); !tk.AFKSet || tk.AFKFlag {
		t.Errorf("afk: false parsed as set=%v flag=%v", tk.AFKSet, tk.AFKFlag)
	}
	if tk := parse("type: task"); tk.AFKSet || tk.AFKFlag || tk.PR != "" {
		t.Errorf("absent fields parsed as set=%v flag=%v pr=%q", tk.AFKSet, tk.AFKFlag, tk.PR)
	}
}

func TestTicketAFK(t *testing.T) {
	tests := []struct {
		typ  Type
		flag bool
		want bool
	}{
		{TypeResearch, false, true},
		{TypeImplement, false, true},
		{TypeGrilling, false, false},
		{TypePrototype, false, false},
		{TypeTask, false, false},
		{TypeTask, true, true},
		{TypeGrilling, true, false},
		{TypeResearch, true, true},
	}
	for _, tc := range tests {
		tk := &Ticket{Type: tc.typ, AFKFlag: tc.flag}
		if got := tk.AFK(); got != tc.want {
			t.Errorf("%s afk=%v: AFK() = %v, want %v", tc.typ, tc.flag, got, tc.want)
		}
	}
}

func TestParseMapImplementedSection(t *testing.T) {
	src := "# M\n\n## Destination\nD.\n\n## Notes\n[Elsewhere](./tickets/09-x.md)\n\n" +
		"## Decisions so far\n\n- [Dec](./tickets/01-a.md): g.\n\n" +
		"## Implemented so far\n\n- [Built](./tickets/02-b.md): shipped.\n- [Also](./tickets/03-c.md): shipped too.\n\n" +
		"```markdown\n- [Quoted](./tickets/07-q.md): an example.\n```\n\n## Out of scope\n"
	m := ParseMap("map.md", src)
	if len(m.Implemented) != 2 || m.Implemented[0].TicketNum != 2 || m.Implemented[1].TicketNum != 3 {
		t.Fatalf("implemented = %+v, want [2 3] (Notes link and fenced bullet must not count)", m.Implemented)
	}
	if m.Implemented[0].Line == 0 {
		t.Error("implemented line not recorded")
	}
	if len(m.Decisions) != 1 || m.Decisions[0].TicketNum != 1 {
		t.Errorf("decisions = %+v, want [1]", m.Decisions)
	}
}

func TestFencedImplementedHeadingIsNotStructure(t *testing.T) {
	src := "# M\n\n## Destination\nD.\n\n```markdown\n## Implemented so far\n\n- [Quoted](./tickets/07-q.md): x.\n```\n"
	if m := ParseMap("map.md", src); len(m.Implemented) != 0 {
		t.Errorf("implemented = %+v, want none", m.Implemented)
	}
}

func implTk(num int, status Status, pr string, blocked ...int) *Ticket {
	t := tk(num, status, blocked...)
	t.Type = TypeImplement
	t.PR = pr
	return t
}

const (
	okPR       = "https://example.com/o/r/pull/1"
	implMap    = "# M\n\n## Destination\nD.\n\n## Decisions so far\n\n## Implemented so far\n\n- [T](./tickets/01-a.md): g.\n\n## Out of scope\n"
	decMapOne  = "# M\n\n## Destination\nD.\n\n## Decisions so far\n\n- [T](./tickets/01-a.md): g.\n\n## Implemented so far\n"
	implMapDup = "# M\n\n## Destination\nD.\n\n## Implemented so far\n\n- [T](./tickets/01-a.md): g.\n- [T](./tickets/01-a.md): again.\n"
)

func levelOf(d []Diagnostic, sub string) (Level, bool) {
	for _, x := range d {
		if strings.Contains(x.Msg, sub) {
			return x.Level, true
		}
	}
	return Warn, false
}

func wantDiag(t *testing.T, d []Diagnostic, sub string, want Level) {
	t.Helper()
	got, ok := levelOf(d, sub)
	if !ok {
		t.Errorf("no diagnostic containing %q; got %v", sub, d)
	} else if got != want {
		t.Errorf("diagnostic %q is %s, want %s", sub, got, want)
	}
}

func wantNoDiag(t *testing.T, d []Diagnostic, sub string) {
	t.Helper()
	if _, ok := levelOf(d, sub); ok {
		t.Errorf("unexpected diagnostic containing %q; got %v", sub, d)
	}
}

func lintIt(src string, ts ...*Ticket) []Diagnostic {
	return Lint(effort(src, ts...), DefaultOptions())
}

func TestLintImplementAccepted(t *testing.T) {
	if d := lintIt(implMap, implTk(1, StatusResolved, okPR)); len(d) != 0 {
		t.Errorf("clean implement map produced %v", d)
	}
	wantNoDiag(t, lintIt(okMap, implTk(1, StatusOpen, "")), "is not research")
}

func TestLintPR(t *testing.T) {
	for _, bad := range []string{"not-a-url", "ftp://x/y", "https://", "[https://a/1, https://b/2]", "https://a/1, https://b/2", "https://a/1 https://b/2"} {
		t.Run("error "+bad, func(t *testing.T) {
			wantDiag(t, lintIt(okMap, implTk(1, StatusOpen, bad)), "single http(s) URL", Error)
		})
	}
	for _, ok := range []string{"http://h/p/1", okPR} {
		t.Run("ok "+ok, func(t *testing.T) {
			wantNoDiag(t, lintIt(okMap, implTk(1, StatusOpen, ok)), "single http(s) URL")
		})
	}

	t1 := tk(1, StatusOpen)
	t1.PR = okPR
	wantDiag(t, lintIt(okMap, t1), "only implement tickets ship a PR", Warn)
	wantNoDiag(t, lintIt(okMap, implTk(1, StatusOpen, okPR)), "only implement tickets ship a PR")
}

func TestLintAFK(t *testing.T) {
	t1 := tk(1, StatusOpen)
	t1.AFKSet, t1.AFKFlag = true, true
	wantDiag(t, lintIt(okMap, t1), "derived from the type", Warn)

	t2 := tk(1, StatusOpen)
	t2.Type, t2.AFKSet, t2.AFKFlag = TypeTask, true, true
	wantNoDiag(t, lintIt(okMap, t2), "derived from the type")
}

func TestLintResolvedImplementWithoutPR(t *testing.T) {
	wantDiag(t, lintIt(implMap, implTk(1, StatusResolved, "")), "has no pr", Warn)
	wantNoDiag(t, lintIt(implMap, implTk(1, StatusResolved, okPR)), "has no pr")
	wantNoDiag(t, lintIt(okMap, implTk(1, StatusOpen, "")), "has no pr")
}

func TestLintImplementIndexSections(t *testing.T) {
	t.Run("implement ticket in Decisions-so-far warns", func(t *testing.T) {
		d := lintIt(decMapOne, implTk(1, StatusResolved, okPR))
		wantDiag(t, d, "lists implement ticket 01", Warn)
		wantDiag(t, d, "absent from the map's Implemented-so-far", Error)
	})
	t.Run("non-implement ticket in Implemented-so-far warns", func(t *testing.T) {
		d := lintIt(implMap, tk(1, StatusResolved))
		wantDiag(t, d, "only implement tickets belong there", Warn)
		wantDiag(t, d, "absent from the map's Decisions-so-far", Error)
	})
	t.Run("resolved implement missing from Implemented-so-far errors, not Decisions", func(t *testing.T) {
		d := lintIt(okMap, implTk(1, StatusResolved, okPR))
		wantDiag(t, d, "absent from the map's Implemented-so-far", Error)
		wantNoDiag(t, d, "absent from the map's Decisions-so-far")
	})
	t.Run("resolved decision missing from Decisions-so-far still errors", func(t *testing.T) {
		wantDiag(t, lintIt(okMap, tk(1, StatusResolved)), "absent from the map's Decisions-so-far", Error)
	})
	t.Run("Implemented-so-far pointing at an unresolved ticket errors", func(t *testing.T) {
		wantDiag(t, lintIt(implMap, implTk(1, StatusOpen, "")), "which is open, not resolved", Error)
	})
	t.Run("Implemented-so-far pointing at a missing ticket errors", func(t *testing.T) {
		wantDiag(t, lintIt(implMap), "Implemented-so-far links ticket 01, which does not exist", Error)
	})
	t.Run("out-of-scope ticket in Implemented-so-far errors", func(t *testing.T) {
		wantDiag(t, lintIt(implMap, implTk(1, StatusOutOfScope, "")), "Implemented-so-far lists 01, which is out of scope", Error)
	})
	t.Run("a ticket listed twice warns", func(t *testing.T) {
		wantDiag(t, lintIt(implMapDup, implTk(1, StatusResolved, okPR)), "Implemented-so-far lists ticket 01 2 times", Warn)
	})
}
