package main

import (
	"fmt"
	"os"
)

// Version is set by goreleaser at build time.
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("𓂀 Sirsi Anubis %s\n", version)
		fmt.Println("  The Guardian of Infrastructure Hygiene")
		fmt.Println("  \"Weigh. Judge. Purge.\"")
		os.Exit(0)
	}

	fmt.Println("𓂀 Sirsi Anubis — The Guardian of Infrastructure Hygiene")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  anubis weigh          Scan your workstation (The Weighing)")
	fmt.Println("  anubis judge          Clean artifacts (The Judgment)")
	fmt.Println("  anubis guard          Manage RAM pressure (The Guardian)")
	fmt.Println("  anubis sight          Fix ghost apps in Spotlight (The Sight)")
	fmt.Println("  anubis hapi           Optimize VRAM & storage (The Flow)")
	fmt.Println("  anubis scarab         Fleet sweep across networks (The Transformer)")
	fmt.Println("  anubis scales         Enforce policies (The Judgment)")
	fmt.Println("  anubis profile        Manage developer profiles")
	fmt.Println("  anubis version        Show version")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --dry-run             Preview without making changes")
	fmt.Println("  --json                Output in JSON format")
	fmt.Println("  --confirm             Skip confirmation prompts")
	fmt.Println()
	fmt.Println("\"Weigh. Judge. Purge.\"")
}
