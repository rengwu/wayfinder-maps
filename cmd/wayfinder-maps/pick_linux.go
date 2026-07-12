//go:build linux

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// pickFolder tries the webkit helper's native GTK dialog first — it ships in
// the same archive, so no extra install — then zenity, then kdialog. Empty
// means cancelled (or nothing available; serve refuses to start in that case,
// see dialogToolErr).
func pickFolder() string {
	if h := helperPath(); h != "" {
		if out, err := exec.Command(h, "--pick-folder").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	if _, err := exec.LookPath("zenity"); err == nil {
		out, err := exec.Command("zenity", "--file-selection", "--directory", "--title=Open a project folder").Output()
		if err != nil {
			return "" // cancel exits nonzero
		}
		return strings.TrimSpace(string(out))
	}
	if _, err := exec.LookPath("kdialog"); err == nil {
		out, err := exec.Command("kdialog", "--getexistingdirectory", ".").Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	return ""
}

// dialogToolErr enforces the GUI's one Linux requirement at startup: some way
// to show a folder dialog. The helper counts when its webkit libraries load;
// zenity or kdialog count on their own. status and lint never call this.
func dialogToolErr() error {
	if h := helperPath(); h != "" {
		if err := exec.Command(h, "--check").Run(); err == nil {
			return nil
		}
	}
	for _, tool := range []string{"zenity", "kdialog"} {
		if _, err := exec.LookPath(tool); err == nil {
			return nil
		}
	}
	return fmt.Errorf("the GUI needs a folder dialog: install libwebkit2gtk-4.1 (and keep wayfinder-maps-webview next to this binary), or install zenity or kdialog")
}
