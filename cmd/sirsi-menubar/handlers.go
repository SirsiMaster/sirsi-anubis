// Package main — sirsi-menubar
//
// handlers.go — Binary discovery and shared utilities.
//
// Since ADR-016 (TUI as Primary Interface), all menu actions route through
// spawnTUIWithCommand() in main.go. This file retains only binary-location
// logic needed by the TUI bridge.
package main

import (
	"os/exec"
)

// findSirsiBinary locates the sirsi binary.
func findSirsiBinary() string {
	// Check PATH first
	if p, err := exec.LookPath("sirsi"); err == nil {
		return p
	}
	// Check Homebrew location
	if p, err := exec.LookPath("/opt/homebrew/bin/sirsi"); err == nil {
		return p
	}
	// Check local bin
	if p, err := exec.LookPath("./bin/sirsi"); err == nil {
		return p
	}
	// Fallback to just "sirsi" and hope it's in PATH
	return "sirsi"
}
