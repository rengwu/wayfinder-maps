package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// settings are the viewer preferences kept across restarts. They live in a file
// rather than localStorage, which is per origin and app mode picks a random port.
type settings struct {
	Direction string `json:"direction"`
}

const defaultDirection = "E"

// validDirection reports whether d is a compass point or "none" (radial layout).
func validDirection(d string) bool {
	switch d {
	case "N", "NE", "E", "SE", "S", "SW", "W", "NW", "none":
		return true
	}
	return false
}

func settingsFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wayfinder-maps", "settings.json"), nil
}

// loadSettings reads the file, falling back to defaults for a missing,
// unreadable or invalid one.
func loadSettings() settings {
	s := settings{Direction: defaultDirection}
	f, err := settingsFile()
	if err != nil {
		return s
	}
	b, err := os.ReadFile(f)
	if err != nil {
		return s
	}
	var got settings
	if json.Unmarshal(b, &got) == nil && validDirection(got.Direction) {
		s.Direction = got.Direction
	}
	return s
}

func saveSettings(s settings) error {
	f, err := settingsFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f, b, 0o644)
}
