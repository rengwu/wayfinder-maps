//go:build windows

package main

import (
	"os/exec"
	"strings"
)

// pickFolder opens the Windows folder dialog via PowerShell, which is part of
// the OS. -Sta because FolderBrowserDialog requires a single-threaded
// apartment; empty output means the user cancelled.
func pickFolder() string {
	const script = `Add-Type -AssemblyName System.Windows.Forms;` +
		`$d = New-Object System.Windows.Forms.FolderBrowserDialog;` +
		`$d.Description = 'Open a project folder';` +
		`if ($d.ShowDialog() -eq 'OK') { Write-Output $d.SelectedPath }`
	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Sta", "-Command", script).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// PowerShell is part of Windows; a folder dialog is always available.
func dialogToolErr() error { return nil }
