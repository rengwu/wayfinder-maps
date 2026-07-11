package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"wayfinder/internal/wayfinder"
)

// mapInfo is a one-line summary of a wayfinder map, for the picker's card list.
type mapInfo struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Resolved int    `json:"resolved"`
	Total    int    `json:"total"`
	Frontier int    `json:"frontier"`
}

// isEffort reports whether dir directly holds a map.md — i.e. it is a wayfinder
// map rather than a project root, a .plan folder, or a spec-only effort.
func isEffort(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, "map.md"))
	return err == nil && !fi.IsDir()
}

// planDirOf resolves where a project's effort subdirectories live: the path
// itself if it is already a .plan, its .plan child when present, else the path.
func planDirOf(project string) string {
	if filepath.Base(project) == ".plan" {
		return project
	}
	if fi, err := os.Stat(filepath.Join(project, ".plan")); err == nil && fi.IsDir() {
		return filepath.Join(project, ".plan")
	}
	return project
}

// listMaps returns every wayfinder map reachable from a chosen project path:
// the effort itself if the path is one, otherwise every child of its .plan that
// holds a map.md. Efforts that fail to parse are skipped rather than erroring.
func listMaps(project string) []mapInfo {
	var out []mapInfo
	add := func(path string) {
		e, err := wayfinder.Load(path)
		if err != nil {
			return
		}
		name := e.Map.Name
		if name == "" {
			name = filepath.Base(path)
		}
		out = append(out, mapInfo{
			Slug:     filepath.Base(path),
			Name:     name,
			Path:     path,
			Resolved: e.Count(wayfinder.StatusResolved),
			Total:    len(e.Tickets),
			Frontier: len(e.Frontier()),
		})
	}
	if isEffort(project) {
		add(project)
		return out
	}
	entries, err := os.ReadDir(planDirOf(project))
	if err != nil {
		return out
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		if p := filepath.Join(planDirOf(project), ent.Name()); isEffort(p) {
			add(p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out
}

// initialKind classifies the optional CLI path: an effort opens straight into
// its map, anything else opens its map list, and "" shows the splash.
func initialKind(dir string) (effort, project string) {
	switch {
	case dir == "":
		return "", ""
	case isEffort(dir):
		return dir, ""
	default:
		return "", dir
	}
}

// --- recent projects ------------------------------------------------------

type recentEntry struct {
	Path string `json:"path"`
	Name string `json:"name"`
	Maps int    `json:"maps"`
}

func recentsFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wayfinder", "recents.json"), nil
}

// loadRecents reads the recents list, dropping folders that no longer exist and
// refreshing each one's live map count.
func loadRecents() []recentEntry {
	f, err := recentsFile()
	if err != nil {
		return nil
	}
	b, err := os.ReadFile(f)
	if err != nil {
		return nil
	}
	var rs []recentEntry
	json.Unmarshal(b, &rs)
	var live []recentEntry
	for _, r := range rs {
		if fi, err := os.Stat(r.Path); err == nil && fi.IsDir() {
			r.Maps = len(listMaps(r.Path))
			live = append(live, r)
		}
	}
	return live
}

// addRecent moves a project to the front of the recents list (deduped, capped).
func addRecent(project string) {
	f, err := recentsFile()
	if err != nil || project == "" {
		return
	}
	out := []recentEntry{{Path: project, Name: filepath.Base(project), Maps: len(listMaps(project))}}
	for _, r := range loadRecents() {
		if r.Path != project {
			out = append(out, r)
		}
	}
	if len(out) > 8 {
		out = out[:8]
	}
	os.MkdirAll(filepath.Dir(f), 0o755)
	if b, err := json.MarshalIndent(out, "", "  "); err == nil {
		os.WriteFile(f, b, 0o644)
	}
}

// pickFolder opens the macOS "choose folder" dialog and returns the POSIX path,
// or "" if the user cancelled (or not on macOS). It shells out to osascript so
// there is no dependency and no main-thread requirement.
func pickFolder() string {
	const script = `try
	set p to choose folder with prompt "Open a project folder"
	return POSIX path of p
on error
	return ""
end try`
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
