// Package main — pantheon-menubar
//
// 𓂀 Pantheon Menu Bar Application (ADR-010)
//
// A native macOS menu bar application that gives Pantheon a persistent visual
// presence. Appears as an ankh icon (☥) in the macOS menu bar with:
//
//   - Live stats panel (RAM, Git status, accelerator, active deities)
//   - Command shortcuts (Scan, Judge, Guard, Ka, Mirror)
//   - Quick actions (Start Watchdog, Open Build Log)
//   - Osiris checkpoint warnings
//
// Build: go build -o bin/pantheon-menubar ./cmd/pantheon-menubar/
// Bundle: make bundle (creates Pantheon.app)
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/systray"
)

var version = "v0.4.0-alpha"

func main() {
	// Check if we should run in headless mode (for testing / CI)
	if os.Getenv("PANTHEON_HEADLESS") == "1" {
		fmt.Println("𓂀 Pantheon Menu Bar " + version)
		fmt.Println("  Running in headless mode (no systray)")
		runHeadless()
		return
	}

	// Launch the real macOS menu bar app
	systray.Run(onReady, onExit)
}

// onReady is called when systray is ready — builds the menu.
func onReady() {
	fmt.Println("𓂀 systray: onReady called, setting template icon...")
	icon := getIcon()
	fmt.Printf("𓂀 systray: icon size = %d bytes\n", len(icon))
	systray.SetTemplateIcon(icon, icon) // Template icon for macOS — system handles tinting
	systray.SetTooltip("𓂀 Pantheon — Active")
	fmt.Println("𓂀 systray: template icon set, building menu...")

	// ── Stats section (refreshed on a timer) ────────────────────────
	cfg := DefaultStatsConfig()
	if root, err := findRepoRoot(); err == nil {
		cfg.RepoDir = root
	}

	snap := CollectStats(cfg)
	items := snap.FormatMenuItems()

	// Create stat menu items (non-clickable info display)
	statItems := make([]*systray.MenuItem, len(items))
	for i, label := range items {
		statItems[i] = systray.AddMenuItem(label, "")
		statItems[i].Disable()
	}

	systray.AddSeparator()

	// ── Commands section ────────────────────────────────────────────
	handlers := PantheonHandlers()
	for _, h := range handlers {
		handler := h
		item := systray.AddMenuItem("  "+handler.Name, "Run "+handler.Name)
		go func() {
			for range item.ClickedCh {
				_ = handler.Execute()
			}
		}()
	}

	systray.AddSeparator()

	// ── Quick actions ───────────────────────────────────────────────
	mBuildLog := systray.AddMenuItem("📄 Open Build Log", "Open the build log in browser")
	mCaseStudies := systray.AddMenuItem("📚 Open Case Studies", "Open case studies in browser")

	systray.AddSeparator()

	// ── Status line ─────────────────────────────────────────────────
	mStatus := systray.AddMenuItem(snap.StatusLine(), "")
	mStatus.Disable()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Pantheon", "Quit the menu bar app")

	// ── Background stats refresh ────────────────────────────────────
	go func() {
		ticker := time.NewTicker(cfg.Interval)
		defer ticker.Stop()
		for range ticker.C {
			s := CollectStats(cfg)
			newItems := s.FormatMenuItems()
			for i, mi := range statItems {
				if i < len(newItems) {
					mi.SetTitle(newItems[i])
				}
			}
			mStatus.SetTitle(s.StatusLine())
		}
	}()

	// ── Click handlers ──────────────────────────────────────────────
	go func() {
		for {
			select {
			case <-mBuildLog.ClickedCh:
				_ = OpenBuildLog()
			case <-mCaseStudies.ClickedCh:
				_ = OpenCaseStudies()
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	fmt.Println("𓂀 Pantheon menu bar exited")
}

// ── Headless fallback ───────────────────────────────────────────────────

// runHeadless runs the stats collector in a terminal-friendly mode.
func runHeadless() {
	cfg := DefaultStatsConfig()
	if root, err := findRepoRoot(); err == nil {
		cfg.RepoDir = root
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	printStats(cfg)

	for {
		select {
		case <-ticker.C:
			printStats(cfg)
		case sig := <-sigCh:
			fmt.Printf("\n𓂀 Pantheon shutting down (signal: %s)\n", sig)
			return
		}
	}
}

func printStats(cfg StatsConfig) {
	snap := CollectStats(cfg)
	fmt.Println("─── 𓂀 Pantheon Status ───────────────────")
	for _, item := range snap.FormatMenuItems() {
		fmt.Printf("  %s\n", item)
	}
	fmt.Printf("  %s\n", snap.StatusLine())
	fmt.Println("──────────────────────────────────────────")
	fmt.Println()
}
