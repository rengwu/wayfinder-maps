package wayfinder

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func loadImplementDemo(t *testing.T) *Effort {
	t.Helper()
	e, err := Load(filepath.Join("testdata", "implement-demo"))
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestImplementDemoLintsClean(t *testing.T) {
	e := loadImplementDemo(t)
	opt := DefaultOptions()
	opt.Now = time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)
	if d := Lint(e, opt); len(d) != 0 {
		t.Errorf("fixture produced diagnostics: %v", d)
	}
}

func TestImplementDemoStaleClaimsWarnLater(t *testing.T) {
	e := loadImplementDemo(t)
	opt := DefaultOptions()
	opt.Now = time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	d := Lint(e, opt)
	if len(d) != 2 || !diagsContain(d, "claimed") {
		t.Errorf("want two stale-claim warnings, got %v", d)
	}
	for _, x := range d {
		if x.Level != Warn {
			t.Errorf("stale claim must warn, got %v", x)
		}
	}
}

func TestImplementDemoFrontier(t *testing.T) {
	e := loadImplementDemo(t)
	var afk, hitl []int
	for _, tk := range e.Frontier() {
		if tk.AFK() {
			afk = append(afk, tk.Num)
		} else {
			hitl = append(hitl, tk.Num)
		}
	}
	if want := []int{9, 14, 17}; !reflect.DeepEqual(afk, want) {
		t.Errorf("AFK frontier = %v, want %v", afk, want)
	}
	if want := []int{13, 16}; !reflect.DeepEqual(hitl, want) {
		t.Errorf("HITL frontier = %v, want %v", hitl, want)
	}
}

func TestImplementDemoCoversEveryCase(t *testing.T) {
	e := loadImplementDemo(t)

	type count struct{ resolved, claimed, open, oos int }
	var decisions, implement count
	byType := map[Type]int{}
	for _, tk := range e.Tickets {
		byType[tk.Type]++
		c := &decisions
		if tk.Type == TypeImplement {
			c = &implement
		}
		switch tk.Status {
		case StatusResolved:
			c.resolved++
		case StatusClaimed:
			c.claimed++
		case StatusOpen:
			c.open++
		case StatusOutOfScope:
			c.oos++
		}
	}
	if want := (count{6, 1, 7, 1}); decisions != want {
		t.Errorf("decision counts = %+v, want %+v", decisions, want)
	}
	if want := (count{1, 1, 2, 1}); implement != want {
		t.Errorf("implement counts = %+v, want %+v", implement, want)
	}
	wantTypes := map[Type]int{TypeResearch: 5, TypePrototype: 2, TypeGrilling: 5, TypeTask: 3, TypeImplement: 5}
	if !reflect.DeepEqual(byType, wantTypes) {
		t.Errorf("tickets per type = %v, want %v", byType, wantTypes)
	}

	if got := len(e.Map.Decisions); got != 6 {
		t.Errorf("Decisions so far = %d, want 6", got)
	}
	if got := len(e.Map.Implemented); got != 1 {
		t.Errorf("Implemented so far = %d, want 1", got)
	}
	if got := len(e.Map.OutOfScope); got != 2 {
		t.Errorf("Out of scope = %d, want 2", got)
	}

	anchored := 0
	for _, f := range e.Map.Fog {
		if f.ClearsWith != 0 {
			anchored++
		}
	}
	if len(e.Map.Fog) != 3 || anchored != 2 {
		t.Errorf("fog = %d patches, %d anchored, want 3 and 2", len(e.Map.Fog), anchored)
	}

	if tk := e.ByNum(3); len(tk.UnderminedBy) != 1 || tk.UnderminedBy[0] != 4 {
		t.Errorf("ticket 03 undermined_by = %v, want [4]", tk.UnderminedBy)
	}
	if tk := e.ByNum(7); tk.PR == "" || tk.Status != StatusResolved {
		t.Errorf("ticket 07 should be resolved with a pr, got %q %s", tk.PR, tk.Status)
	}
	if tk := e.ByNum(14); !tk.AFKFlag || tk.Type != TypeTask {
		t.Errorf("ticket 14 should be an afk task")
	}
	if tk := e.ByNum(11); !reflect.DeepEqual(tk.BlockedBy, []int{8}) || tk.Type == TypeImplement || e.ByNum(8).Type != TypeImplement {
		t.Errorf("ticket 11 should be a decision blocked by implement ticket 08")
	}
	for _, n := range []int{7, 8, 9, 10} {
		if !reflect.DeepEqual(e.ByNum(n).BlockedBy[:1], []int{6}) {
			t.Errorf("implement ticket %02d should be blocked by the planning ticket 06", n)
		}
	}
}
