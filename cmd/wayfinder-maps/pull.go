package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/RamsesKools/wayfinder-maps/scripts"
)

// pullSource is what wayfinder-pull records in <effort>/.source.json.
type pullSource struct {
	Repo    string `json:"repo"`
	Map     int    `json:"map"`
	Auth    string `json:"auth"`
	Account string `json:"account"`
}

// readPullSource reports whether the effort is a GitHub mirror made by
// wayfinder-pull, and where from.
func readPullSource(effort string) (pullSource, bool) {
	var src pullSource
	if effort == "" {
		return src, false
	}
	b, err := os.ReadFile(filepath.Join(effort, ".source.json"))
	if err != nil || json.Unmarshal(b, &src) != nil || src.Repo == "" || src.Map == 0 {
		return src, false
	}
	return src, true
}

var pullMu sync.Mutex

// runPull re-pulls the effort from GitHub by running the embedded
// wayfinder-pull in --refresh mode, which reuses the gh account recorded
// at the first pull. It returns the script's stderr on failure.
var runPull = func(effort string) error {
	pullMu.Lock()
	defer pullMu.Unlock()
	cmd := exec.Command("bash", "-c", scripts.WayfinderPull, "wayfinder-pull", "--refresh", effort)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}
