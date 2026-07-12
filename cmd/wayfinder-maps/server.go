package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rengwu/wayfinder-maps/internal/wayfinder"
)

// serve runs the map viewer as a local web server, blocking until interrupted.
// The GUI requires a working folder dialog (only Linux can lack one), checked
// up front so the splash's Open Folder button is never silently dead.
func serve(dir string) int {
	if err := dialogToolErr(); err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: %v\n", err)
		return 2
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "7777"
	}
	addr := "localhost:" + port
	if dir == "" {
		fmt.Printf("wayfinder-maps: serving at http://%s\n", addr)
	} else {
		fmt.Printf("wayfinder-maps: serving %s at http://%s\n", dir, addr)
	}
	if err := http.ListenAndServe(addr, newServer(dir)); err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder-maps: %v\n", err)
		return 2
	}
	return 0
}

// newServer serves the app: a single-page shell at "/" (splash, map list, and
// the star-map), plus a small JSON API. `initial` is the optional CLI path — an
// effort opens straight into its map, a project into its list, "" the splash.
// Every effort is re-read on request, so a saved edit shows on the next poll.
func newServer(initial string) http.Handler {
	mux := http.NewServeMux()
	writeJSON := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(v); err != nil {
			fmt.Fprintf(os.Stderr, "wayfinder-maps: encode: %v\n", err)
		}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(shellHTML))
	})

	// Where the app should open: {effort} jumps to a map, {project} to a list.
	mux.HandleFunc("/api/initial", func(w http.ResponseWriter, r *http.Request) {
		effort, project := initialKind(initial)
		writeJSON(w, map[string]string{"effort": effort, "project": project})
	})

	// The native macOS folder picker; returns {path:""} when cancelled.
	mux.HandleFunc("/api/pick", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"path": pickFolder()})
	})

	mux.HandleFunc("/api/recents", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, loadRecents())
	})

	// Forget a recent project, and hand back the trimmed list so the client can
	// re-render from the truth on disk. POST, because it writes. Only the recents
	// entry is dropped; the project folder is never touched.
	mux.HandleFunc("/api/recents/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		removeRecent(r.URL.Query().Get("project"))
		writeJSON(w, loadRecents())
	})

	// Maps in a chosen project; opening one records it in recents.
	mux.HandleFunc("/api/maps", func(w http.ResponseWriter, r *http.Request) {
		project := r.URL.Query().Get("project")
		addRecent(project)
		writeJSON(w, listMaps(project))
	})

	mux.HandleFunc("/api/graph", func(w http.ResponseWriter, r *http.Request) {
		e, err := wayfinder.Load(r.URL.Query().Get("effort"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, buildGraph(e))
	})

	// A cheap change token the client polls: the newest mtime across map.md +
	// tickets/ plus the file count, so an edit, add or delete all move it.
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, effortVersion(r.URL.Query().Get("effort")))
	})

	return mux
}

// effortVersion is the newest mtime across map.md and tickets/*.md, plus the
// ticket count — enough to detect edits, additions and deletions cheaply.
func effortVersion(dir string) string {
	var newest int64
	var count int
	if fi, err := os.Stat(filepath.Join(dir, "map.md")); err == nil {
		newest = fi.ModTime().UnixNano()
	}
	if entries, err := os.ReadDir(filepath.Join(dir, "tickets")); err == nil {
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}
			count++
			if fi, err := e.Info(); err == nil {
				if m := fi.ModTime().UnixNano(); m > newest {
					newest = m
				}
			}
		}
	}
	return fmt.Sprintf("%d-%d", newest, count)
}

// --- the graph document served to the client ------------------------------

type graphDoc struct {
	Name        string     `json:"name"`
	Destination string     `json:"destination"`
	Counts      countsDoc  `json:"counts"`
	Nodes       []nodeDoc  `json:"nodes"`
	Edges       []edgeDoc  `json:"edges"`
	Fog         []fogDoc   `json:"fog"`
}

type countsDoc struct {
	Resolved   int `json:"resolved"`
	Claimed    int `json:"claimed"`
	Open       int `json:"open"`
	OutOfScope int `json:"outOfScope"`
	Total      int `json:"total"`
}

type nodeDoc struct {
	Num        int    `json:"num"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	Status     string `json:"status"` // resolved|claimed|frontier|blocked|out_of_scope
	Rank       int    `json:"rank"`
	Undermined bool   `json:"undermined"`
	ClaimedBy  string `json:"claimedBy,omitempty"`
	Blockers   []int  `json:"blockers"`
	Body       string `json:"body"`
}

type edgeDoc struct {
	From      int  `json:"from"`
	To        int  `json:"to"`
	Satisfied bool `json:"satisfied"`
}

type fogDoc struct {
	Title      string `json:"title"`
	ClearsWith int    `json:"clearsWith"`
}

// displayStatus folds the frontier — a subset of open tickets — into a flat set
// of visual categories the client draws directly.
func displayStatus(t *wayfinder.Ticket, frontier map[int]bool) string {
	switch t.Status {
	case wayfinder.StatusResolved:
		return "resolved"
	case wayfinder.StatusClaimed:
		return "claimed"
	case wayfinder.StatusOutOfScope:
		return "out_of_scope"
	default:
		if frontier[t.Num] {
			return "frontier"
		}
		return "blocked"
	}
}

func buildGraph(e *wayfinder.Effort) graphDoc {
	name := e.Map.Name
	if name == "" {
		name = e.Name
	}

	frontier := map[int]bool{}
	for _, t := range e.Frontier() {
		frontier[t.Num] = true
	}

	rank := map[int]int{}
	for ri, layer := range e.Layers() {
		for _, t := range layer {
			rank[t.Num] = ri
		}
	}

	g := graphDoc{
		Name:        name,
		Destination: e.Map.Destination,
		Counts: countsDoc{
			Resolved:   e.Count(wayfinder.StatusResolved),
			Claimed:    e.Count(wayfinder.StatusClaimed),
			Open:       e.Count(wayfinder.StatusOpen),
			OutOfScope: e.Count(wayfinder.StatusOutOfScope),
			Total:      len(e.Tickets),
		},
	}

	for _, t := range e.Tickets {
		g.Nodes = append(g.Nodes, nodeDoc{
			Num:        t.Num,
			Title:      t.Title,
			Type:       string(t.Type),
			Status:     displayStatus(t, frontier),
			Rank:       rank[t.Num],
			Undermined: len(t.UnderminedBy) > 0,
			ClaimedBy:  t.ClaimedBy,
			Blockers:   t.BlockedBy,
			Body:       ticketBody(t.Path),
		})
		for _, b := range t.BlockedBy {
			satisfied := false
			if dep := e.ByNum(b); dep != nil {
				satisfied = dep.Status == wayfinder.StatusResolved
			}
			g.Edges = append(g.Edges, edgeDoc{From: b, To: t.Num, Satisfied: satisfied})
		}
	}

	for _, f := range e.Map.Fog {
		g.Fog = append(g.Fog, fogDoc{Title: f.Title, ClearsWith: f.ClearsWith})
	}
	return g
}

// ticketBody returns a ticket's prose with the YAML frontmatter stripped, for
// the detail panel. Reading fails soft: a missing file yields an empty body
// rather than a failed request.
func ticketBody(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				lines = lines[i+1:]
				break
			}
		}
	}
	// Drop the leading H1 title — the panel header already shows it.
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "# ") {
		lines = lines[1:]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
