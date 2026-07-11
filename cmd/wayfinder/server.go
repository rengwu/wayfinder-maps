package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"wayfinder/internal/wayfinder"
)

// serve runs the map viewer as a local web server, blocking until interrupted.
func serve(dir string) int {
	port := os.Getenv("PORT")
	if port == "" {
		port = "7777"
	}
	addr := "localhost:" + port
	fmt.Printf("wayfinder: serving %s at http://%s\n", dir, addr)
	if err := http.ListenAndServe(addr, newServer(dir)); err != nil {
		fmt.Fprintf(os.Stderr, "wayfinder: %v\n", err)
		return 2
	}
	return 0
}

// newServer serves the star-map: a static canvas shell at "/", and the graph as
// JSON at "/graph.json". The effort is re-read on every /graph.json request, so
// a saved edit shows on the next refresh; the shell itself is static.
func newServer(dir string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(shellHTML))
	})

	mux.HandleFunc("/graph.json", func(w http.ResponseWriter, r *http.Request) {
		e, err := wayfinder.Load(dir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(buildGraph(e)); err != nil {
			fmt.Fprintf(os.Stderr, "wayfinder: encode: %v\n", err)
		}
	})

	return mux
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
