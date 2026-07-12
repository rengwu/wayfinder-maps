//go:build darwin

package main

import (
	"os/exec"
	"strings"
)

// pickFolder opens the macOS "choose folder" dialog and returns the POSIX
// path, or "" if the user cancelled. It shells out to osascript so there is no
// dependency and no main-thread requirement.
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

// osascript is part of macOS; a folder dialog is always available.
func dialogToolErr() error { return nil }
