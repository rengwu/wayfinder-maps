//go:build !darwin && !windows && !linux

package main

// No native dialog story on this platform; the button reads as a cancel.
func pickFolder() string { return "" }

func dialogToolErr() error { return nil }
